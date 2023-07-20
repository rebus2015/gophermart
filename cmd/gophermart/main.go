package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/rebus2015/gophermart/cmd/internal/config"
	"github.com/rebus2015/gophermart/cmd/internal/logger"
	m "github.com/rebus2015/gophermart/cmd/internal/migrations"
)

func main() {
	cfg, err := config.GetConfig()

	if err != nil {
		log.Panicf("Error reading configuration from env variables: %v", err)
		return
	}

	ctx := logger.InitZerolog(context.Background(), cfg)
	logger.New(ctx).Infof("Logger started")

	logger.New(ctx).Infof("server started \n address:%v \n accrualService: '%v', \n database:%v,\n restore interval: %v ",
		cfg.RunAddress, cfg.AccruralAddr, cfg.ConnectionString, cfg.SyncInterval)

	err = m.RunMigrations(ctx, cfg)
	if err != nil {
		logger.New(ctx).Fatalf("MigrationsError:%v", err)
		return
	}
	srv := &http.Server{
		Addr:         cfg.RunAddress,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
	}
	err = srv.ListenAndServe()
	if err != nil {
		logger.New(ctx).Fatalf("server exited with %v", err)
	}
}
