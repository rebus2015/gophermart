package migrations

import (
	"embed"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/rebus2015/gophermart/cmd/internal/logger"
)

type dbConfig interface {
	GetDBConnection() string
}

//go:embed *.sql
var embedMigrations embed.FS

func RunMigrations(lg *logger.Logger, config dbConfig) error {

	goose.SetBaseFS(embedMigrations)

	db, err := goose.OpenDBWithDriver("pgx", config.GetDBConnection())
	if err != nil {
		lg.Error().Err(err).Msgf("goose: failed to open DB: %v\n", config.GetDBConnection())
		return fmt.Errorf("goose failed to open DB %v", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			lg.Error().Err(err).Msgf("goose: failed to open DB: %v\n", config.GetDBConnection())
		}
	}()

	if err := goose.Up(db, "."); err != nil {
		lg.Error().Err(err).Msg("goose failed to Up migrations")
		return fmt.Errorf("goose failed to Up migrations: %v", err)

	}
	return nil
}
