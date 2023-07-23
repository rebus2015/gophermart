package config

import (
	"flag"
	"time"

	"github.com/caarlos0/env"
)

type Config struct {
	RunAddress       string        `env:"RUN_ADDRESS"`
	SyncInterval     time.Duration `env:"SYNC_INTERVAL"`          // период синхронизации с системой начислений
	AccruralAddr     string        `env:"ACCRUAL_SYSTEM_ADDRESS"` //адрес системы расчета начислений
	ConnectionString string        `env:"DATABASE_URI"`           // строка подключения к БД	
	Debug             bool        `env:"DBUG_MODE"`               // уровень логирования
}

func GetConfig() (*Config, error) {
	conf := Config{}

	flag.StringVar(&conf.RunAddress, "a", "127.0.0.1:8080", "Server address")
	flag.DurationVar(&conf.SyncInterval, "i", time.Second*30, "Accrual system data request interval")
	flag.StringVar(&conf.AccruralAddr, "r", "127.0.0.1:8088", "Accrual system address")
	flag.StringVar(&conf.ConnectionString, "d", "postgresql://pguser:pgpwd@localhost:5432/gophermart?sslmode=disable", "Database connection string(PostgreSql)")
	// postgresql://pguser:pgpwd@localhost:5432/devops?sslmode=disable
	flag.BoolVar(&conf.Debug, "l", true,
		"logger mode")	
	flag.Parse()

	err := env.Parse(&conf)

	return &conf, err
}

func (conf *Config) IsDebug() bool {
	return conf.Debug
}

func (conf *Config) GetDbConnection() string {
	return conf.ConnectionString
}
