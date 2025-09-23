package main

import (
	"context"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	"ariga.io/atlas/sql/sqltool"
	"rezics.com/task-queue/service/auth/ent/migrate"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql/schema"

	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
)

const name = "auth"

func main() {
	raw_shadow, err := exec.Command("encore", "db", "conn-uri", name, "--shadow").Output()
	if err != nil {
		slog.Error("failed getting shadow database URL", "error", err)
		return
	}
	shadow := strings.TrimSpace(string(raw_shadow))

	reset := exec.Command("encore", "db", "reset", name, "--shadow")
	if out, err := reset.CombinedOutput(); err != nil {
		slog.Error("failed resetting shadow database", "error", err)
		println(string(out))
		return
	}

	if len(os.Args) != 2 {
		slog.Error("migration name is required. Use: 'go run -mod=mod migrations/main.go <name>'")
		return
	}

	ctx := context.Background()

	dir, err := sqltool.NewGolangMigrateDir("./migrations")
	if err != nil {
		slog.Error("failed creating migration directory", "error", err)
		return
	}

	opts := []schema.MigrateOption{
		schema.WithDir(dir),                         // provide migration directory
		schema.WithMigrationMode(schema.ModeReplay), // provide migration mode
		schema.WithDialect(dialect.Postgres),        // Ent dialect to use
	}

	if err := migrate.NamedDiff(ctx, shadow, os.Args[1], opts...); err != nil {
		slog.Error("failed generating migration file", "error", err)
		return
	}
}
