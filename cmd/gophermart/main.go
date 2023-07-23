package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/rebus2015/gophermart/cmd/internal/api/handlers"
	"github.com/rebus2015/gophermart/cmd/internal/api/middleware"
	"github.com/rebus2015/gophermart/cmd/internal/config"
	"github.com/rebus2015/gophermart/cmd/internal/logger"
	m "github.com/rebus2015/gophermart/cmd/internal/migrations"
	"github.com/rebus2015/gophermart/cmd/internal/router"
	"github.com/rebus2015/gophermart/cmd/internal/storage"
)

func main() {
	cfg, err := config.GetConfig()
	if err != nil {
		log.Panicf("Error reading configuration from env variables: %v", err)
		return
	}
	lg := logger.NewConsole(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err = m.RunMigrations(lg, cfg)
	if err != nil {
		lg.Fatal().Err(err).Msgf("Migrations retuned error")
		return
	}
	repo, err := storage.NewStorage(ctx, lg, cfg)
	if err != nil {
		lg.Fatal().Err(err).Msgf("Error creating dbStorage, with conn: %s", cfg.ConnectionString)
		return
	}
	h := handlers.NewApi(repo, lg)
	m := middleware.NewMiddlewares(repo, lg)
	handle := router.NewRouter(m, h)

	srv := &http.Server{
		Addr:         cfg.RunAddress,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		Handler:      handle,
	}

	lg.Info().Msgf("server started \n address:%v \n accrualService: '%v', \n database:%v,\n restore interval: %v ",
		cfg.RunAddress, cfg.AccruralAddr, cfg.ConnectionString, cfg.SyncInterval)

	err = srv.ListenAndServe()
	if err != nil {
		lg.Fatal().Err(err).Msg("server exited with error")
	}

}
