package main

import (
	"log"
	"os"
	"strings"

	pb "github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"

	"github.com/dr4ghs/orgtool/cron"
	_ "github.com/dr4ghs/orgtool/migrations"
)

func main() {
	app := pb.New()

	isGoRun := strings.HasPrefix(os.Args[0], os.TempDir())

	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		Automigrate: isGoRun,
	})

	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		cron.InitMigrationsCron(app)

		return e.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
