package task

import (
	"context"
	"time"

	"encore.dev/rlog"
	"encore.dev/storage/sqldb"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"rezics.com/task-queue/service/task/ent"

	"github.com/go-co-op/gocron"
)

var _database = sqldb.NewDatabase("task", sqldb.DatabaseConfig{
	Migrations: "migrations",
})

//encore:service
type Service struct {
	Database *ent.Client
}

func initService() (*Service, error) {
	service := Service{
		Database: ent.NewClient(
			ent.Driver(
				sql.OpenDB(
					dialect.Postgres,
					_database.Stdlib(),
				),
			),
		),
	}

	service.Clean(context.TODO())
	scheduler := gocron.NewScheduler(time.UTC)
	scheduler.Every(time.Second * 5).Do(func() {
		if count, err := service.Clean(context.TODO()); err == nil && count > 0 {
			rlog.Info("cleaned up worker", "count", count)
		}
	})
	scheduler.StartAsync()

	return &service, nil
}

var _ = initService
