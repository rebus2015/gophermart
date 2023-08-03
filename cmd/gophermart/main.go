package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/rebus2015/gophermart/cmd/internal/api/auth"
	"github.com/rebus2015/gophermart/cmd/internal/api/handlers"
	"github.com/rebus2015/gophermart/cmd/internal/api/middleware"
	"github.com/rebus2015/gophermart/cmd/internal/client"
	"github.com/rebus2015/gophermart/cmd/internal/config"
	"github.com/rebus2015/gophermart/cmd/internal/logger"
	m "github.com/rebus2015/gophermart/cmd/internal/migrations"
	"github.com/rebus2015/gophermart/cmd/internal/router"
	"github.com/rebus2015/gophermart/cmd/internal/storage/dbstorage"
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
	repo, err := dbstorage.NewStorage(ctx, lg, cfg)
	if err != nil {
		lg.Fatal().Err(err).Msgf("Error creating dbStorage, with conn: %s", cfg.ConnectionString)
		return
	}

	a := auth.NewAuth(lg, cfg)
	h := handlers.NewAPI(repo, lg, a)
	m := middleware.NewMiddlewares(repo, lg, a)
	handle := router.NewRouter(m, h)
	accrualClient := client.NewClient(ctx, repo, cfg, lg)
	accrualClient.Run()

	srv := &http.Server{
		Addr:         cfg.RunAddress,
		ReadTimeout:  160 * time.Second,
		WriteTimeout: 160 * time.Second,
		Handler:      handle,
	}

	lg.Info().Msgf("server started \n address:%v \n accrualService: '%v', \n database:%v,\n restore interval: %v ",
		cfg.RunAddress, cfg.AccruralAddr, cfg.ConnectionString, cfg.SyncInterval)

	err = srv.ListenAndServe()
	if err != nil {
		lg.Fatal().Err(err).Msg("server exited with error")
	}
}
