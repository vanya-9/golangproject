package core_pgx_pool

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	Timeout  time.Duration
}

func NewConfig() (Config, error) {
	var config Config

	if err := envconfig.Process("POSTGRES", &config); err != nil {
		return Config{}, fmt.Errorf("process envconfig: %w", err)
	}
	return config, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		panic(fmt.Errorf("create config: %w", err))
	}

	return config
}
