package migrations

import (
	"fmt"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/hook"
	"github.com/pocketbase/pocketbase/tools/types"
)

// =============================================================================
// ACTIVITIES
//

func addActivityTypeField(app core.App) error {
	collection, err := app.FindCollectionByNameOrId("activities")
	if err != nil {
		return err
	}

	collection.Fields.Add(&core.SelectField{
		Name:      "type",
		Required:  true,
		MaxSelect: 1,
		Values: []string{
			"daily",
			"weekly",
			"monthly",
			"yearly",
		},
	})

	return app.Save(collection)
}

func removeActivityTypeField(app core.App) error {
	collection, err := app.FindCollectionByNameOrId("activities")
	if err != nil {
		return err
	}

	collection.Fields.RemoveById("type")

	return app.Save(collection)
}

func addActivityAPIRules(app core.App) error {
	collection, err := app.FindCollectionByNameOrId("activities")
	if err != nil {
		return err
	}

	collection.ListRule = types.Pointer("@request.auth.id = '' || @request.auth.id = user")
	collection.ViewRule = types.Pointer("@request.auth.id = '' || @request.auth.id = user")
	collection.CreateRule = types.Pointer("@request.auth.id = '' || @request.auth.id = user")
	collection.UpdateRule = types.Pointer("@request.auth.id = '' || @request.auth.id = user")
	collection.DeleteRule = types.Pointer("@request.auth.id = '' || @request.auth.id = user")

	return app.Save(collection)
}

func removeActivityAPIRules(app core.App) error {
	collection, err := app.FindCollectionByNameOrId("activities")
	if err != nil {
		return err
	}

	collection.ListRule = types.Pointer("")
	collection.ViewRule = types.Pointer("")
	collection.CreateRule = types.Pointer("")
	collection.UpdateRule = types.Pointer("")
	collection.DeleteRule = types.Pointer("")

	return app.Save(collection)
}

// Hooks -----------------------------------------------------------------------

func createActivityEntryHookBind(app core.App) {
	app.OnRecordAfterCreateSuccess("activities").Unbind("activities_onCreate")
	app.OnRecordAfterCreateSuccess("activities").Bind(&hook.Handler[*core.RecordEvent]{
		Id: "activities-onCreateSuccess_createEntry",
		Func: func(e *core.RecordEvent) error {
			typ := e.Record.GetString("type")
			if typ != "daily" && typ != "weekly" && typ != "monthly" && typ != "yearly" {
				return fmt.Errorf("Not known activity type '%s'\n", typ)
			}

			collection, err := e.App.FindCollectionByNameOrId(fmt.Sprintf("%s_entries", typ))
			if err != nil {
				return err
			}

			record := core.NewRecord(collection)
			record.Set("activity", e.Record.Id)
			record.Set("progress", 0)
			record.Set("goal", e.Record.GetInt("goal"))
			record.Set("closed", false)

			if err := e.App.Save(record); err != nil {
				return err
			}

			return e.Next()
		},
	})
}

func createActivityEntryHookUnbind(app core.App) {
	app.OnRecordAfterCreateSuccess("activities").Unbind("activities-onCreateSuccess_createEntry")
}

func preventActivityOwnerChangeHookBind(app core.App) {
	app.OnRecordUpdateRequest("activities").Bind(&hook.Handler[*core.RecordRequestEvent]{
		Id: "activities-onUpdateRequest_changeOwner",
		Func: func(e *core.RecordRequestEvent) error {
			activity, err := e.App.FindRecordById("activities", e.Record.Id)
			if err != nil {
				return err
			}

			if activity.GetString("user") != e.Record.GetString("user") {
				return fmt.Errorf("Cannot change activity owner")
			}

			return e.Next()
		},
	})
}

func preventActivityOwnerChangeHookUnbind(app core.App) {
	app.OnRecordUpdateRequest("activities").Unbind("activities-onUpdateRequest_changeOwner")
}

func changeActivityTypeHookBind(app core.App) {
	app.OnRecordAfterCreateSuccess("activities").Bind(&hook.Handler[*core.RecordEvent]{
		Id: "activities-onCreateSuccess_changeActivity",
		Func: func(e *core.RecordEvent) error {
			activity, err := e.App.FindRecordById("activities", e.Record.Id)
			if err != nil {
				return err
			}

			oldType := activity.GetString("type")
			newType := e.Record.GetString("type")
			if oldType == newType {
				return e.Next()
			}

			if oldType != "daily" && oldType != "weekly" && oldType != "monthly" &&
				oldType != "yearly" {
				return fmt.Errorf("Unknown activity type of '%s'", oldType)
			}

			oldRecord, err := e.App.FindFirstRecordByFilter(
				fmt.Sprintf("%s_entries", oldType),
				"closed = False",
			)
			if err != nil {
				return err
			}

			collection, err := e.App.FindCollectionByNameOrId(fmt.Sprintf("%s_entries", newType))
			if err != nil {
				return err
			}

			newRecord := core.NewRecord(collection)
			newRecord.Set("activity", e.Record.Id)
			newRecord.Set("progress", oldRecord.GetInt("progress"))
			newRecord.Set("goal", oldRecord.GetInt("goal"))
			newRecord.Set("closed", false)

			if err := e.App.Delete(oldRecord); err != nil {
				return err
			}

			if err := e.App.Save(newRecord); err != nil {
				return err
			}

			return e.Next()
		},
	})
}

func changeActivityTypeHookUnbind(app core.App) {
	app.OnRecordAfterCreateSuccess("activities").Unbind("activities-onCreateSuccess_changeActivity")
}

// =============================================================================
// REWARDS
//

func addRewardsAPIRules(app core.App) error {
	collection, err := app.FindCollectionByNameOrId("rewards")
	if err != nil {
		return err
	}

	collection.ListRule = types.Pointer("@request.auth.id = '' || @request.auth.id = user")
	collection.ViewRule = types.Pointer("@request.auth.id = '' || @request.auth.id = user")
	collection.CreateRule = types.Pointer("@request.auth.id = '' || @request.auth.id = user")
	collection.UpdateRule = types.Pointer("@request.auth.id = '' || @request.auth.id = user")
	collection.DeleteRule = types.Pointer("@request.auth.id = '' || @request.auth.id = user")

	return app.Save(collection)
}

func removeRewardsAPIRules(app core.App) error {
	collection, err := app.FindCollectionByNameOrId("rewards")
	if err != nil {
		return err
	}

	collection.ListRule = types.Pointer("")
	collection.ViewRule = types.Pointer("")
	collection.CreateRule = types.Pointer("")
	collection.UpdateRule = types.Pointer("")
	collection.DeleteRule = types.Pointer("")

	return app.Save(collection)
}

// =============================================================================
// DAILY ENTRIES
//

func addDailyEntriesAPIRules(app core.App) error {
	collection, err := app.FindCollectionByNameOrId("daily_entries")
	if err != nil {
		return err
	}

	collection.ListRule = types.Pointer(
		"@request.auth.id = '' || (@request.auth.id = activity.user && closed = false)",
	)
	collection.ViewRule = types.Pointer(
		"@request.auth.id = '' || (@request.auth.id = activity.user && closed = false)",
	)
	collection.CreateRule = types.Pointer("@request.auth.id = ''")
	collection.UpdateRule = types.Pointer(
		"@request.auth.id = '' || (@request.auth.id = activity.user && closed = false)",
	)
	collection.DeleteRule = types.Pointer("@request.auth.id = ''")

	return app.Save(collection)
}

func removeDailyEntriesAPIRules(app core.App) error {
	collection, err := app.FindCollectionByNameOrId("daily_entries")
	if err != nil {
		return err
	}

	collection.ListRule = types.Pointer("")
	collection.ViewRule = types.Pointer("")
	collection.CreateRule = types.Pointer("")
	collection.UpdateRule = types.Pointer("")
	collection.DeleteRule = types.Pointer("")

	return app.Save(collection)
}

// Hooks -----------------------------------------------------------------------

func checkClosedDailyEntryOnUpdateHookBind(app core.App) {
	app.OnRecordUpdateRequest("daily_entries").Unbind("daily_entries_onUpdate")
	app.OnRecordUpdateRequest("daily_entries").Bind(&hook.Handler[*core.RecordRequestEvent]{
		Id: "daily_entries-onUpdateRequest_closed",
		Func: func(e *core.RecordRequestEvent) error {
			rec, err := e.App.FindRecordById("daily_entries", e.Record.Id)
			if err != nil {
				return err
			}

			if rec.GetBool("closed") {
				return fmt.Errorf("Is not possible to reopen a closed entry")
			}

			return e.Next()
		},
	})
}

func checkClosedDailyEntryOnUpdateHookUnbind(app core.App) {
	app.OnRecordUpdateRequest("daily_entries").Unbind("daily_entries-onUpdateRequest_closed")
}

func checkClosedDailyEntryOnDeleteHookBind(app core.App) {
	app.OnRecordUpdateRequest("daily_entries").Unbind("daily_entries_onDelete")
	app.OnRecordDeleteRequest("daily_entries").Bind(&hook.Handler[*core.RecordRequestEvent]{
		Id: "daily_entries-onDeleteRequest_closed",
		Func: func(e *core.RecordRequestEvent) error {
			if e.Record.GetBool("closed") {
				return fmt.Errorf("Is not possible to delete a closed entry")
			}

			return e.Next()
		},
	})
}

func checkClosedDailyEntryOnDeleteHookUnbind(app core.App) {
	app.OnRecordDeleteRequest("daily_entries").Unbind("daily_entries-onDeleteRequest_closed")
}

// =============================================================================
// WEEKLY ACTIVITIES
//

func createWeeklyEntries(app core.App) error {
	collection := core.NewBaseCollection("weekly_entries")

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

func deleteWeeklyEntries(app core.App) error {
	collection, err := app.FindCollectionByNameOrId("weekly_entries")
	if err != nil {
		return err
	}

	return app.Delete(collection)
}

func addWeeklyEntriesAPIRules(app core.App) error {
	collection, err := app.FindCollectionByNameOrId("daily_entries")
	if err != nil {
		return err
	}

	collection.ListRule = types.Pointer(
		"@request.auth.id = '' || (@request.auth.id = activity.user && closed = false)",
	)
	collection.ViewRule = types.Pointer(
		"@request.auth.id = '' || (@request.auth.id = activity.user && closed = false)",
	)
	collection.CreateRule = types.Pointer("@request.auth.id = ''")
	collection.UpdateRule = types.Pointer(
		"@request.auth.id = '' || (@request.auth.id = activity.user && closed = false)",
	)
	collection.DeleteRule = types.Pointer("@request.auth.id = ''")

	return app.Save(collection)
}

func removeWeeklyEntriesAPIRules(app core.App) error {
	collection, err := app.FindCollectionByNameOrId("daily_entries")
	if err != nil {
		return err
	}

	collection.ListRule = types.Pointer("")
	collection.ViewRule = types.Pointer("")
	collection.CreateRule = types.Pointer("")
	collection.UpdateRule = types.Pointer("")
	collection.DeleteRule = types.Pointer("")

	return app.Save(collection)
}

// Hooks -----------------------------------------------------------------------

func checkClosedWeeklyEntryOnUpdateHookBind(app core.App) {
	app.OnRecordUpdateRequest("weekly_entries").Bind(&hook.Handler[*core.RecordRequestEvent]{
		Id: "weekly_entries-onUpdateRequest_closed",
		Func: func(e *core.RecordRequestEvent) error {
			rec, err := e.App.FindRecordById("weekly_entries", e.Record.Id)
			if err != nil {
				return err
			}

			if rec.GetBool("closed") {
				return fmt.Errorf("Is not possible to reopen a closed entry")
			}

			return e.Next()
		},
	})
}

func checkClosedWeeklyEntryOnUpdateHookUnbind(app core.App) {
	app.OnRecordUpdateRequest("weekly_entries").Unbind("weekly_entries-onUpdateRequest_closed")
}

func checkClosedWeeklyEntryOnDeleteHookBind(app core.App) {
	app.OnRecordDeleteRequest("weekly_entries").Bind(&hook.Handler[*core.RecordRequestEvent]{
		Id: "weekly_entries-onDeleteRequest_closed",
		Func: func(e *core.RecordRequestEvent) error {
			if e.Record.GetBool("closed") {
				return fmt.Errorf("Is not possible to delete a closed entry")
			}

			return e.Next()
		},
	})
}

func checkClosedWeeklyEntryOnDeleteHookUnbind(app core.App) {
	app.OnRecordDeleteRequest("weekly_entries").Unbind("weekly_entries-onDeleteRequest_closed")
}

// =============================================================================
// MONTHLY ACTIVITIES
//

func createMonthlyEntries(app core.App) error {
	collection := core.NewBaseCollection("monthly_entries")

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

func deleteMonthlyEntries(app core.App) error {
	collection, err := app.FindCollectionByNameOrId("monthly_entries")
	if err != nil {
		return err
	}

	return app.Delete(collection)
}

func addMonthlyEntriesAPIRules(app core.App) error {
	collection, err := app.FindCollectionByNameOrId("daily_entries")
	if err != nil {
		return err
	}

	collection.ListRule = types.Pointer(
		"@request.auth.id = '' || (@request.auth.id = activity.user && closed = false)",
	)
	collection.ViewRule = types.Pointer(
		"@request.auth.id = '' || (@request.auth.id = activity.user && closed = false)",
	)
	collection.CreateRule = types.Pointer("@request.auth.id = ''")
	collection.UpdateRule = types.Pointer(
		"@request.auth.id = '' || (@request.auth.id = activity.user && closed = false)",
	)
	collection.DeleteRule = types.Pointer("@request.auth.id = ''")

	return app.Save(collection)
}

func removeMonthlyEntriesAPIRules(app core.App) error {
	collection, err := app.FindCollectionByNameOrId("daily_entries")
	if err != nil {
		return err
	}

	collection.ListRule = types.Pointer("")
	collection.ViewRule = types.Pointer("")
	collection.CreateRule = types.Pointer("")
	collection.UpdateRule = types.Pointer("")
	collection.DeleteRule = types.Pointer("")

	return app.Save(collection)
}

// Hooks -----------------------------------------------------------------------

func checkClosedMonthlyEntryOnUpdateHookBind(app core.App) {
	app.OnRecordUpdateRequest("monthly_entries").Bind(&hook.Handler[*core.RecordRequestEvent]{
		Id: "monthly_entries-onUpdateRequest_closed",
		Func: func(e *core.RecordRequestEvent) error {
			rec, err := e.App.FindRecordById("monthly_entries", e.Record.Id)
			if err != nil {
				return err
			}

			if rec.GetBool("closed") {
				return fmt.Errorf("Is not possible to reopen a closed entry")
			}

			return e.Next()
		},
	})
}

func checkClosedMonthlyEntryOnUpdateHookUnbind(app core.App) {
	app.OnRecordUpdateRequest("monthly_entries").Unbind("monthly_entries-onUpdateRequest_closed")
}

func checkClosedMonthlyEntryOnDeleteHookBind(app core.App) {
	app.OnRecordDeleteRequest("monthly_entries").Bind(&hook.Handler[*core.RecordRequestEvent]{
		Id: "monthly_entries-onDeleteRequest_closed",
		Func: func(e *core.RecordRequestEvent) error {
			if e.Record.GetBool("closed") {
				return fmt.Errorf("Is not possible to delete a closed entry")
			}

			return e.Next()
		},
	})
}

func checkClosedMonthlyEntryOnDeleteHookUnbind(app core.App) {
	app.OnRecordDeleteRequest("monthly_entries").Unbind("monthly_entries-onDeleteRequest_closed")
}

// =============================================================================
// YEARLY ACTIVITIES
//

func createYearlyEntries(app core.App) error {
	collection := core.NewBaseCollection("yearly_entries")

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

func deleteYearlyEntries(app core.App) error {
	collection, err := app.FindCollectionByNameOrId("yearly_entries")
	if err != nil {
		return nil
	}

	return app.Delete(collection)
}

func addYearlyEntriesAPIRules(app core.App) error {
	collection, err := app.FindCollectionByNameOrId("daily_entries")
	if err != nil {
		return err
	}

	collection.ListRule = types.Pointer(
		"@request.auth.id = '' || (@request.auth.id = activity.user && closed = false)",
	)
	collection.ViewRule = types.Pointer(
		"@request.auth.id = '' || (@request.auth.id = activity.user && closed = false)",
	)
	collection.CreateRule = types.Pointer("@request.auth.id = ''")
	collection.UpdateRule = types.Pointer(
		"@request.auth.id = '' || (@request.auth.id = activity.user && closed = false)",
	)
	collection.DeleteRule = types.Pointer("@request.auth.id = ''")

	return app.Save(collection)
}

func removeYearlyEntriesAPIRules(app core.App) error {
	collection, err := app.FindCollectionByNameOrId("daily_entries")
	if err != nil {
		return err
	}

	collection.ListRule = types.Pointer("")
	collection.ViewRule = types.Pointer("")
	collection.CreateRule = types.Pointer("")
	collection.UpdateRule = types.Pointer("")
	collection.DeleteRule = types.Pointer("")

	return app.Save(collection)
}

// Hooks -----------------------------------------------------------------------

func checkClosedYearlyEntryOnUpdateHookBind(app core.App) {
	app.OnRecordUpdateRequest("yearly_entries").Bind(&hook.Handler[*core.RecordRequestEvent]{
		Id: "yearly_entries-onUpdateRequest_closed",
		Func: func(e *core.RecordRequestEvent) error {
			rec, err := e.App.FindRecordById("yearly_entries", e.Record.Id)
			if err != nil {
				return err
			}

			if rec.GetBool("closed") {
				return fmt.Errorf("Is not possible to reopen a closed entry")
			}

			return e.Next()
		},
	})
}

func checkClosedYearlyEntryOnUpdateHookUnbind(app core.App) {
	app.OnRecordUpdateRequest("yearly_entries").Unbind("yearly_entries-onUpdateRequest_closed")
}

func checkClosedYearlyEntryOnDeleteHookBind(app core.App) {
	app.OnRecordDeleteRequest("yearly_entries").Bind(&hook.Handler[*core.RecordRequestEvent]{
		Id: "yearly_entries-onDeleteRequest_closed",
		Func: func(e *core.RecordRequestEvent) error {
			if e.Record.GetBool("closed") {
				return fmt.Errorf("Is not possible to delete a closed entry")
			}

			return e.Next()
		},
	})
}

func checkClosedYearlyEntryOnDeleteHookUnbind(app core.App) {
	app.OnRecordDeleteRequest("yearly_entries").Unbind("yearly_entries-onDeleteRequest_closed")
}

// =============================================================================
// MIGRATIONS
//

func init() {
	m.Register(
		func(app core.App) error {
			// Tables
			{ // Activities
				if err := addActivityTypeField(app); err != nil {
					return err
				}

				if err := addActivityAPIRules(app); err != nil {
					return err
				}
			}

			{ // Rewards
				if err := addRewardsAPIRules(app); err != nil {
					return err
				}
			}

			{ // Daily entries
				if err := addDailyEntriesAPIRules(app); err != nil {
					return err
				}
			}

			{ // Weekly entries
				if err := createWeeklyEntries(app); err != nil {
					return err
				}

				if err := addWeeklyEntriesAPIRules(app); err != nil {
					return err
				}
			}

			{ // Monthly entries
				if err := createMonthlyEntries(app); err != nil {
					return err
				}

				if err := addMonthlyEntriesAPIRules(app); err != nil {
					return err
				}
			}

			{ // Yearly entries
				if err := createYearlyEntries(app); err != nil {
					return err
				}

				if err := addYearlyEntriesAPIRules(app); err != nil {
					return err
				}
			}

			// Hooks
			{ // Activities
				createActivityEntryHookBind(app)
				preventActivityOwnerChangeHookBind(app)
				changeActivityTypeHookBind(app)
			}

			{ // Daily entries
				checkClosedDailyEntryOnUpdateHookBind(app)
				checkClosedDailyEntryOnDeleteHookBind(app)
			}

			{ // Weekly entries
				checkClosedWeeklyEntryOnUpdateHookBind(app)
				checkClosedWeeklyEntryOnDeleteHookBind(app)
			}

			{ // Monthly entries
				checkClosedMonthlyEntryOnUpdateHookBind(app)
				checkClosedMonthlyEntryOnDeleteHookBind(app)
			}

			{ // Yearly entries
				checkClosedYearlyEntryOnUpdateHookBind(app)
				checkClosedYearlyEntryOnDeleteHookBind(app)
			}

			return nil
		},
		func(app core.App) error {
			// Tables
			{ // Activities
				if err := removeActivityTypeField(app); err != nil {
					return err
				}

				if err := removeActivityAPIRules(app); err != nil {
					return err
				}
			}

			{ // Rewards
				if err := removeRewardsAPIRules(app); err != nil {
					return err
				}
			}

			{ // Daily entries
				if err := removeDailyEntriesAPIRules(app); err != nil {
					return err
				}
			}

			{ // Weekly entries
				if err := deleteWeeklyEntries(app); err != nil {
					return err
				}

				if err := removeWeeklyEntriesAPIRules(app); err != nil {
					return err
				}
			}

			{ // Monthly entries
				if err := deleteMonthlyEntries(app); err != nil {
					return err
				}

				if err := removeMonthlyEntriesAPIRules(app); err != nil {
					return err
				}
			}

			{ // Yearly entries
				if err := deleteYearlyEntries(app); err != nil {
					return err
				}

				if err := removeYearlyEntriesAPIRules(app); err != nil {
					return err
				}
			}

			// Hooks
			{ // Activities
				createActivityEntryHookUnbind(app)
				preventActivityOwnerChangeHookUnbind(app)
				changeActivityTypeHookUnbind(app)
			}

			{ // Daily entries
				checkClosedDailyEntryOnUpdateHookUnbind(app)
				checkClosedDailyEntryOnDeleteHookUnbind(app)
			}

			{ // Weekly entries
				checkClosedWeeklyEntryOnUpdateHookUnbind(app)
				checkClosedWeeklyEntryOnDeleteHookUnbind(app)
			}

			{ // Monthly entries
				checkClosedMonthlyEntryOnUpdateHookUnbind(app)
				checkClosedMonthlyEntryOnDeleteHookUnbind(app)
			}

			{ // Yearly entries
				checkClosedYearlyEntryOnUpdateHookUnbind(app)
				checkClosedYearlyEntryOnDeleteHookUnbind(app)
			}

			return nil
		},
	)
}
