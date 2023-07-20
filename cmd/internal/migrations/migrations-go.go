package migrations

import (
	"context"
	"embed"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/rebus2015/gophermart/cmd/internal/logger"
)

type dbConfig interface {
	GetDbConnection() string
}

//go:embed *.sql
var embedMigrations embed.FS

func RunMigrations(ctx context.Context, config dbConfig) error {

	goose.SetBaseFS(embedMigrations)

	db, err := goose.OpenDBWithDriver("pgx", config.GetDbConnection())
	if err != nil {
		logger.New(ctx).Fatalf("goose: failed to open DB: %v\n", err)
		return fmt.Errorf("goose: failed to open DB: %v\n", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			logger.New(ctx).Fatalf("goose: failed to open DB: %v\n", err)
		}
	}()

	if err := goose.Up(db, "."); err != nil {
		logger.New(ctx).Fatalf("goose: failed to Up migrations: %v\n", err)
		return fmt.Errorf("goose: failed to Up migrations: %v\n", err)

	}
	return nil
}
