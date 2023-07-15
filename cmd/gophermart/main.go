package main

import (
	"gophermart/cmd/internal/config"
	"log"
	"net/http"
	"time"
)

func main() {
	cfg, err := config.GetConfig()
	if err != nil {
		log.Panicf("Error reading configuration from env variables: %v", err)
		return
	}
	log.Printf("server started on %v with \n accrualService: '%v', \n database:%v,\n restore: %v ",
		cfg.RunAddress, cfg.AccruralAddr, cfg.ConnectionString, cfg.SyncInterval)

	srv := &http.Server{
		Addr:         cfg.RunAddress,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
	}
	err = srv.ListenAndServe()
	if err != nil {
		log.Printf("server exited with %v", err)
	}
}
