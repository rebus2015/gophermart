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
	Debug            bool          `env:"DBUG_MODE"`              // уровень логирования
	RateLimit        int           `env:"RATE_LIMIT"`             // частота запросов
	SecretKey        string        `env:"KEY"`                    //JWT ключ для формирования подписи
}

func GetConfig() (*Config, error) {
	conf := Config{}

	flag.StringVar(&conf.RunAddress, "a", "", "Server address")
	flag.DurationVar(&conf.SyncInterval, "i", time.Second*1, "Accrual system data request interval")
	flag.StringVar(&conf.AccruralAddr, "r", "", "Accrual system address")
	flag.StringVar(&conf.ConnectionString, "d", "", "Database connection string(PostgreSql)")
	// postgresql://pguser:pgpwd@localhost:5432/gophermart?sslmode=disable
	flag.BoolVar(&conf.Debug, "l", true,
		"logger mode")
	flag.IntVar(&conf.RateLimit, "m", 3, "Rate limit for accrual client")
	flag.StringVar(&conf.ConnectionString, "л", "", "JWT Key to create signature")
	flag.Parse()

	err := env.Parse(&conf)

	return &conf, err
}

func (conf *Config) IsDebug() bool {
	return conf.Debug
}

func (conf *Config) GetDBConnection() string {
	return conf.ConnectionString
}

func (conf *Config) GetAccruralAddr() string {
	return conf.AccruralAddr
}

func (conf *Config) GetSyncInterval() time.Duration {
	return conf.SyncInterval
}

func (conf *Config) GetRateLimit() int {
	return conf.RateLimit
}
func (conf *Config) GetSecretKey() string {
	return conf.SecretKey
}

