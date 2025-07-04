package cron

import (
	"log"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

func calculatePointsCron(app core.App) func() {
	return func() {
		app.RunInTransaction(func(txApp core.App) error {
			entries, err := txApp.FindAllRecords(
				"daily_entries",
				dbx.NewExp("closed = False"),
			)
			log.Printf("Entries to update %d\n", len(entries))
			if err != nil {
				log.Println(err)
				return err
			}

			for _, entry := range entries {
				entry.Set("closed", true)
				if err := txApp.Save(entry); err != nil {
					log.Println(err)
					return err
				}

				if entry.GetInt("progress") < entry.GetInt("goal") {
					log.Printf(
						"No progress for %s\n: %d/%d",
						entry.Id,
						entry.GetInt("progress"),
						entry.GetInt("goal"),
					)
					continue
				}

				activity, err := txApp.FindRecordById("activities", entry.GetString("activity"))
				if err != nil {
					log.Println(err)
					return err
				}

				user, err := txApp.FindRecordById("users", activity.GetString("user"))
				if err != nil {
					log.Println(err)
					return err
				}

				user.Set("points", user.GetInt("points")+activity.GetInt("points"))
				if err := txApp.Save(user); err != nil {
					log.Println(err)
					return err
				}

			}

			return nil
		})
	}
}

func createNewDailyEntriesCron(app core.App) func() {
	return func() {
		app.RunInTransaction(func(txApp core.App) error {
			activities, err := txApp.FindAllRecords("activities")
			log.Println(len(activities))
			if err != nil {
				return err
			}

			entries, err := txApp.FindCollectionByNameOrId("daily_entries")
			if err != nil {
				return err
			}

			for _, activity := range activities {
				log.Printf("Inserting %s\n", activity.Id)
				record := core.NewRecord(entries)

				record.Set("activity", activity.Id)
				record.Set("progress", 0)
				record.Set("goal", activity.GetInt("goal"))
				record.Set("closed", false)

				if err := txApp.Save(record); err != nil {
					log.Println(err)
					return err
				}
			}

			return nil
		})
	}
}

func updateRedeemedRewardsCron(app core.App) func() {
	return func() {
		app.RunInTransaction(func(txApp core.App) error {
			rewards, err := txApp.FindAllRecords("rewards")
			if err != nil {
				return err
			}

			for _, reward := range rewards {
				reward.Set("redeemed", 0)
				reward.Set("used", 0)

				if err := txApp.Save(reward); err != nil {
					return err
				}
			}

			return nil
		})
	}
}
