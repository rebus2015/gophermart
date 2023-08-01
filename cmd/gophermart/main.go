package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/rebus2015/gophermart/cmd/internal/api/handlers"
	"github.com/rebus2015/gophermart/cmd/internal/api/middleware"
	"github.com/rebus2015/gophermart/cmd/internal/client"
	"github.com/rebus2015/gophermart/cmd/internal/config"
	"github.com/rebus2015/gophermart/cmd/internal/logger"
	m "github.com/rebus2015/gophermart/cmd/internal/migrations"
	"github.com/rebus2015/gophermart/cmd/internal/router"
	"github.com/rebus2015/gophermart/cmd/internal/storage/dbstorage"
	"github.com/rebus2015/gophermart/cmd/internal/storage/memstorage"
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
		//lg.Fatal().Err(err).Msgf("Migrations retuned error")
		//return
		log.Panicf("Migrations retuned error: %v", err)
		return
	}
	repo, err := dbstorage.NewStorage(ctx, lg, cfg)
	if err != nil {
		//lg.Fatal().Err(err).Msgf("Error creating dbStorage, with conn: %s", cfg.ConnectionString)
		log.Panicf("Error creating dbStorage, with conn: %s", cfg.ConnectionString)
		return
	}
	orders := memstorage.NewStorage(ctx, repo, cfg, lg)
	err = orders.Restore()
	if err != nil {
		log.Panicf("MemStorage Restore failed: %v", err)
		//lg.Fatal().Err(err).Msgf("MemStorage Restore failed")
		return
	}

	h := handlers.NewAPI(repo, lg, orders, cfg)
	m := middleware.NewMiddlewares(repo, lg)
	handle := router.NewRouter(m, h)
	accrualClient := client.NewClient(ctx, orders, cfg, lg)
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
		//lg.Fatal().Err(err).Msg("server exited with error")
		log.Panicf("MemStorage Restore failed: %v", err)
	}
}
