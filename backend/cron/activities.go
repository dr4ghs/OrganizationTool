package cron

import (
	"log"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

func calculatePointsV2Cron(app core.App) func() {
	app.Cron().Remove("calculatePoints")

	tables := []string{
		"daily_entries",
		"weekly_entries",
		"monthly_entries",
		"yearly_entries",
	}

	return func() {
		for _, table := range tables {
			app.RunInTransaction(func(txApp core.App) error {
				daily, err := txApp.FindAllRecords(table, dbx.NewExp("closed = False"))
				if err != nil {
					return err
				}

				for _, entry := range daily {
					entry.Set("closed", true)
					if err := txApp.Save(entry); err != nil {
						return err
					}

					if entry.GetInt("progress") < entry.GetInt("goal") {
						continue
					}

					activity, err := txApp.FindRecordById("activities", entry.GetString("activity"))
					if err != nil {
						return err
					}

					user, err := txApp.FindRecordById("users", activity.GetString("user"))
					if err != nil {
						return err
					}

					user.Set("points", user.GetInt("points")+activity.GetInt("points"))
					if err := txApp.Save(user); err != nil {
						return err
					}
				}

				return nil
			})
		}
	}
}

func createNewDailyEntriesV2Cron(app core.App) func() {
	app.Cron().Remove("createNewDailyEntries")

	return func() {
		app.RunInTransaction(func(txApp core.App) error {
			activities, err := txApp.FindAllRecords("activities", dbx.NewExp("type = 'daily'"))
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

func createNewWeeklyEntriesCron(app core.App) func() {
	return func() {
		app.RunInTransaction(func(txApp core.App) error {
			activities, err := txApp.FindAllRecords("activities", dbx.NewExp("type = 'weekly'"))
			log.Println(len(activities))
			if err != nil {
				return err
			}

			entries, err := txApp.FindCollectionByNameOrId("weekly_entries")
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

func createNewMonthlyEntriesCron(app core.App) func() {
	return func() {
		app.RunInTransaction(func(txApp core.App) error {
			activities, err := txApp.FindAllRecords("activities", dbx.NewExp("type = 'monthly'"))
			log.Println(len(activities))
			if err != nil {
				return err
			}

			entries, err := txApp.FindCollectionByNameOrId("monthly_entries")
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

func createNewYearlyEntriesCron(app core.App) func() {
	return func() {
		app.RunInTransaction(func(txApp core.App) error {
			activities, err := txApp.FindAllRecords("activities", dbx.NewExp("type = 'yearly'"))
			log.Println(len(activities))
			if err != nil {
				return err
			}

			entries, err := txApp.FindCollectionByNameOrId("yearly_entries")
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
