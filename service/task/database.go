package queue

import (
	"encore.dev/storage/sqldb"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"rezics.com/task-queue/service/task/ent"
)

var Database = ent.NewClient(
	ent.Driver(
		sql.OpenDB(
			dialect.Postgres,
			sqldb.NewDatabase("queue", sqldb.DatabaseConfig{
				Migrations: "migrations",
			}).Stdlib())))
