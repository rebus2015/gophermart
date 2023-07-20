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
	LogLevel         string        `env:"LOG_LEVEL"`              // уровень логирования crit | error | info | debug"
	Mode             string        `env:"LOG_MODE"`               // уровень логирования dev | prod
}

func GetConfig() (*Config, error) {
	conf := Config{}

	flag.StringVar(&conf.RunAddress, "a", "127.0.0.1:8080", "Server address")
	flag.DurationVar(&conf.SyncInterval, "i", time.Second*30, "Accrual system data request interval")
	flag.StringVar(&conf.AccruralAddr, "r", "127.0.0.1:8088", "Accrual system address")
	flag.StringVar(&conf.ConnectionString, "d", "postgresql://pguser:pgpwd@localhost:5432/gophermart?sslmode=disable", "Database connection string(PostgreSql)")
	// postgresql://pguser:pgpwd@localhost:5432/devops?sslmode=disable
	flag.StringVar(&conf.LogLevel, "l", "info",
		"logger verbosity level, crit | error | info | debug")
	flag.StringVar(&conf.Mode, "m", "dev",
		"logger mode, possible values:\n- dev, for colored unstructured output\n- prod, for colorless JSON output\n")
	flag.Parse()

	err := env.Parse(&conf)

	return &conf, err
}

func (conf *Config) GetLogLevel() string {
	return conf.LogLevel
}
func (conf *Config) GetLogMode() string {
	return conf.Mode
}

func (conf *Config) GetDbConnection() string {
	return conf.ConnectionString
}
