package migrations

import (
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

func RunMigrations(lg *logger.Logger, config dbConfig) error {

	goose.SetBaseFS(embedMigrations)

	db, err := goose.OpenDBWithDriver("pgx", config.GetDbConnection())
	if err != nil {
		lg.Error().Err(err).Msgf("goose: failed to open DB: %v\n", config.GetDbConnection())
		return fmt.Errorf("goose: failed to open DB: %v\n", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			lg.Error().Err(err).Msgf("goose: failed to open DB: %v\n", config.GetDbConnection())
		}
	}()

	if err := goose.Up(db, "."); err != nil {
		lg.Error().Err(err).Msg("goose: failed to Up migrations\n")
		return fmt.Errorf("goose: failed to Up migrations: %v\n", err)

	}
	return nil
}
