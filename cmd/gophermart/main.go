package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/rebus2015/gophermart/cmd/internal/config"
	"github.com/rebus2015/gophermart/cmd/internal/logger"
)

func main() {
	cfg, err := config.GetConfig()

	if err != nil {
		log.Panicf("Error reading configuration from env variables: %v", err)
		return
	}

	ctx := logger.InitZerolog(context.Background(), cfg)
	logger.New(ctx).Infof("Logger started")

	logger.New(ctx).Infof("server started on %v with \n accrualService: '%v', \n database:%v,\n restore: %v ",
		cfg.RunAddress, cfg.AccruralAddr, cfg.ConnectionString, cfg.SyncInterval)

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
