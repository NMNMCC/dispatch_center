package auth

import (
	"encore.dev/storage/sqldb"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"rezics.com/task-queue/service/auth/ent"
)

var _database = sqldb.NewDatabase("auth", sqldb.DatabaseConfig{
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

	return &service, nil
}

var _ = initService
