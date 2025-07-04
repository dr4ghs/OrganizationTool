package cron

import (
	"github.com/pocketbase/pocketbase/core"
)

type MigrationCron struct {
	Name    string
	CronTab string
	Func    func()
}

func NewMigrationCron(
	name string,
	cronTab string,
	fun func(),
) MigrationCron {
	return MigrationCron{
		Name:    name,
		CronTab: cronTab,
		Func:    fun,
	}
}

var migrationCrons map[int][]MigrationCron

func InitMigrationsCron(app core.App) {
	migrationCrons = make(map[int][]MigrationCron)
	migrationCrons[7] = []MigrationCron{
		NewMigrationCron("calculatePoints", "0 6 * * *", calculatePointsCron(app)),
		NewMigrationCron("createNewDailyEntries", "1 6 * * *", createNewDailyEntriesCron(app)),
		NewMigrationCron("updateRedeemedRewards", "0 6 * * *", updateRedeemedRewardsCron(app)),
	}
	migrationCrons[8] = []MigrationCron{
		NewMigrationCron("calculatePoints", "0 6 * * *", calculatePointsV2Cron(app)),
		NewMigrationCron("createNewDailyEntries", "1 6 * * *", createNewDailyEntriesV2Cron(app)),
		NewMigrationCron("createNewWeeklyEntries", "1 6 * * 1", createNewWeeklyEntriesCron(app)),
		NewMigrationCron("createNewMonthlyEntries", "1 6 1 * *", createNewMonthlyEntriesCron(app)),
		NewMigrationCron("createNewYearlyEntries", "1 6 1 1 *", createNewYearlyEntriesCron(app)),
	}

	applyMigrationCron(app)
}

func applyMigrationCron(app core.App) {
	ids, err := getMigrationIDs(app)
	if err != nil {
		app.Logger().Warn("It was not possible to retrieve migrations IDs")
	}

	for _, id := range ids {
		if entry, ok := migrationCrons[id]; ok {
			for _, cron := range entry {
				app.Cron().MustAdd(cron.Name, cron.CronTab, cron.Func)
			}
		}
	}
}

func getMigrationIDs(app core.App) (id []int, err error) {
	migrations := []struct {
		Id      int `db:"id"`
		Applied int `db:"applied"`
	}{}
	err = app.DB().
		NewQuery("SELECT (ROW_NUMBER() OVER()) AS id, applied FROM _migrations ORDER BY applied ASC").
		All(&migrations)

	id = make([]int, 0)
	for _, m := range migrations {
		id = append(id, m.Id)
	}

	return
}
