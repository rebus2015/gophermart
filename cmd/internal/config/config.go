package config

import (
	"flag"
	"time"

	"github.com/caarlos0/env"
)


type Config struct {
	RunAddress    string        `env:"RUN_ADDRESS"`
	SyncInterval    time.Duration `env:"SYNC_INTERVAL"`   // период синхронизации с системой начислений
    AccruralAddr    string `env:"ACCRUAL_SYSTEM_ADDRESS"` //адрес системы расчета начислений
	ConnectionString string        `env:"DATABASE_URI"`   // Cтрока подключения к БД
}

func GetConfig() (*Config, error) {
	conf := Config{}

	flag.StringVar(&conf.RunAddress, "a", "127.0.0.1:8080", "Server address")
	flag.DurationVar(&conf.SyncInterval, "i", time.Second*30, "Accrual system data request interval")
	flag.StringVar(&conf.AccruralAddr, "r", "127.0.0.1:8088", "Accrual system address")
    flag.StringVar(&conf.ConnectionString, "d", "",
		"Database connection string(PostgreSql)") // postgresql://pguser:pgpwd@localhost:5432/devops?sslmode=disable
	flag.Parse()

	err := env.Parse(&conf)

	return &conf, err
}