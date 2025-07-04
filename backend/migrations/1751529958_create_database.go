package migrations

import (
	"fmt"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/hook"
)

// =============================================================================
// USERS
//

func updateUsers(app core.App) error {
	collection, err := app.FindCollectionByNameOrId("users")
	if err != nil {
		return err
	}

	collection.Fields.Add(&core.NumberField{
		Name: "points",
	})

	return app.Save(collection)
}

func revertUsers(app core.App) error {
	collection, err := app.FindCollectionByNameOrId("users")
	if err != nil {
		return err
	}

	collection.Fields.RemoveByName("points")

	return app.Save(collection)
}

// =============================================================================
// ACTIVITIES
//

func createActivities(app core.App) error {
	collection := core.NewBaseCollection("activities")

	// Fields
	userCollection, err := app.FindCollectionByNameOrId("users")
	if err != nil {
		return err
	}

	collection.Fields.Add(
		&core.TextField{
			Name:     "name",
			Required: true,
		},
		&core.RelationField{
			Name:          "user",
			Required:      true,
			CascadeDelete: true,
			MinSelect:     1,
			MaxSelect:     1,
			CollectionId:  userCollection.Id,
		},
		&core.NumberField{
			Name:    "points",
			OnlyInt: true,
		},
		&core.NumberField{
			Name:     "goal",
			Required: true,
			OnlyInt:  true,
		},
		&core.AutodateField{
			Name:     "created",
			OnCreate: true,
		},
		&core.AutodateField{
			Name:     "updated",
			OnCreate: true,
			OnUpdate: true,
		},
	)
	return app.Save(collection)
}

func deleteActivities(app core.App) error {
	collection, err := app.FindCollectionByNameOrId("activities")
	if err != nil {
		return err
	}

	return app.Delete(collection)
}

func createActivitiesHooks(app core.App) {
	app.OnRecordAfterCreateSuccess("activities").Bind(&hook.Handler[*core.RecordEvent]{
		Id: "activities_onCreate",
		Func: func(e *core.RecordEvent) error {
			collection, err := e.App.FindCollectionByNameOrId("daily_entries")
			if err != nil {
				return err
			}

			record := core.NewRecord(collection)
			record.Set("activity", e.Record.Id)
			record.Set("progress", 0)
			record.Set("goal", e.Record.Get("goal"))
			record.Set("closed", false)

			if err := e.App.Save(record); err != nil {
				return err
			}

			return e.Next()
		},
	})
}

func deleteActivitiesHooks(app core.App) {
	app.OnRecordAfterCreateSuccess("activities").Unbind("activities_onCreate")
}

// =============================================================================
// REWARDS
//

func createRewards(app core.App) error {
	collection := core.NewBaseCollection("rewards")

	// Fields
	userCollection, err := app.FindCollectionByNameOrId("users")
	if err != nil {
		return err
	}

	collection.Fields.Add(
		&core.TextField{
			Name:     "name",
			Required: true,
		},
		&core.RelationField{
			Name:          "user",
			Required:      true,
			CascadeDelete: true,
			MinSelect:     1,
			MaxSelect:     1,
			CollectionId:  userCollection.Id,
		},
		&core.NumberField{
			Name:     "unit_cost",
			Required: true,
			OnlyInt:  true,
		},
		&core.NumberField{
			Name:    "redeemed",
			OnlyInt: true,
		},
		&core.NumberField{
			Name:    "used",
			OnlyInt: true,
		},
		&core.NumberField{
			Name:     "max_redeemables",
			Required: true,
			OnlyInt:  true,
		},
		&core.AutodateField{
			Name:     "created",
			OnCreate: true,
		},
		&core.AutodateField{
			Name:     "updated",
			OnCreate: true,
			OnUpdate: true,
		},
	)

	return app.Save(collection)
}

func deleteRewards(app core.App) error {
	collection, err := app.FindCollectionByNameOrId("rewards")
	if err != nil {
		return err
	}

	return app.Delete(collection)
}

func createRewardsHooks(app core.App) {
	app.OnRecordUpdateRequest("rewards").Bind(&hook.Handler[*core.RecordRequestEvent]{
		Id:       "rewards-OnUpdateRequest_checkRequest",
		Priority: 0,
		Func: func(e *core.RecordRequestEvent) error {
			reward, err := e.App.FindRecordById("rewards", e.Record.Id)
			if err != nil {
				return err
			}

			// Redeem
			redeemed := e.Record.GetInt("redeemed") - reward.GetInt("redeemed")
			if redeemed < 0 {
				return fmt.Errorf(
					"Cannot update redeemed rewards: new value is smaller than the old one",
				)
			}

			if e.Record.GetInt("redeemed") > reward.GetInt("max_redeemables") {
				return fmt.Errorf("Redeemed rewards exceded the max redeemables limit")
			}

			// Use
			used := e.Record.GetInt("used") - reward.GetInt("used")
			if used < 0 {
				return fmt.Errorf("Cannot use %d rewards", used)
			}

			if e.Record.GetInt("used") > reward.GetInt("redeemed") {
				return fmt.Errorf("Already used all redeemed rewards")
			}

			return e.Next()
		},
	})

	app.OnRecordUpdateRequest("rewards").Bind(&hook.Handler[*core.RecordRequestEvent]{
		Id:       "rewards-OnUpdateRequest_redeem",
		Priority: 2,
		Func: func(e *core.RecordRequestEvent) error {
			reward, err := e.App.FindRecordById("rewards", e.Record.Id)
			if err != nil {
				return err
			}

			user, err := e.App.FindRecordById("users", reward.GetString("user"))
			if err != nil {
				return err
			}

			redeemed := e.Record.GetInt("redeemed") - reward.GetInt("redeemed")
			cost := redeemed * reward.GetInt("unit_cost")
			if user.GetInt("points") < cost {
				return fmt.Errorf("Not enough points to redeem reward")
			}

			user.Set("points", user.GetInt("points")-cost)

			if err := e.App.Save(user); err != nil {
				e.App.Logger().Error(err.Error())
				return err
			}

			return e.Next()
		},
	})
}

func deleteRewardsHooks(app core.App) {
	app.OnRecordUpdateRequest("rewards").Unbind("rewards-OnUpdateRequest_checkRequest")
	app.OnRecordUpdateRequest("rewards").Unbind("rewards-OnUpdateRequest_redeem")
}

// =============================================================================
// DAILY ENTRIES
//

func createDailyEntries(app core.App) error {
	collection := core.NewBaseCollection("daily_entries")

	// Fields
	activities, err := app.FindCollectionByNameOrId("activities")
	if err != nil {
		return err
	}

	collection.Fields.Add(
		&core.RelationField{
			Name:          "activity",
			Required:      true,
			CollectionId:  activities.Id,
			MinSelect:     1,
			MaxSelect:     1,
			CascadeDelete: true,
		},
		&core.NumberField{
			Name:    "progress",
			OnlyInt: true,
		},
		&core.NumberField{
			Name:     "goal",
			Required: true,
			OnlyInt:  true,
		},
		&core.BoolField{
			Name: "closed",
		},
		&core.AutodateField{
			Name:     "created",
			OnCreate: true,
		},
		&core.AutodateField{
			Name:     "updated",
			OnCreate: true,
			OnUpdate: true,
		},
	)

	return app.Save(collection)
}

func deleteDailyEntries(app core.App) error {
	collection, err := app.FindCollectionByNameOrId("daily_entries")
	if err != nil {
		return err
	}

	return app.Delete(collection)
}

func createDailyEntriesHooks(app core.App) {
	app.OnRecordUpdateRequest("daily_entries").Bind(&hook.Handler[*core.RecordRequestEvent]{
		Id: "daily_entries_onUpdate",
		Func: func(e *core.RecordRequestEvent) error {
			rec, _ := e.App.FindRecordById("daily_entries", e.Record.Id)
			if rec.GetBool("closed") {
				return fmt.Errorf("Is not possible to reopen a closed entry")
			}

			return e.Next()
		},
	})

	app.OnRecordDeleteRequest("daily_entries").Bind(&hook.Handler[*core.RecordRequestEvent]{
		Id: "daily_entries_onDelete",
		Func: func(e *core.RecordRequestEvent) error {
			if e.Record.GetBool("closed") {
				return fmt.Errorf("Is not possible to delete a closed entry")
			}

			return e.Next()
		},
	})
}

func deleteDailyEntriesHooks(app core.App) {
	app.OnRecordUpdateRequest("daily_entries").Unbind("daily_entries_onUpdate")
	app.OnRecordDeleteRequest("daily_entries").Unbind("daily_entries_onDelete")
}

// =============================================================================
// MIGRATIONS
//

func init() {
	m.Register(
		func(app core.App) error {
			// Tables
			if err := updateUsers(app); err != nil {
				return err
			}

			if err := createActivities(app); err != nil {
				return err
			}

			if err := createRewards(app); err != nil {
				return err
			}

			if err := createDailyEntries(app); err != nil {
				return err
			}

			// Hooks
			createRewardsHooks(app)
			createActivitiesHooks(app)
			createDailyEntriesHooks(app)

			return nil
		},
		func(app core.App) error {
			// Tables
			if err := revertUsers(app); err != nil {
				return err
			}

			if err := deleteActivities(app); err != nil {
				return err
			}

			if err := deleteRewards(app); err != nil {
				return err
			}

			if err := deleteDailyEntries(app); err != nil {
				return err
			}

			// Hooks
			deleteRewardsHooks(app)
			deleteActivitiesHooks(app)
			deleteDailyEntriesHooks(app)

			return nil
		},
	)
}
