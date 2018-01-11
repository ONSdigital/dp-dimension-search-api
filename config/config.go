package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config is the filing resource handler config
type Config struct {
	BindAddr                string        `envconfig:"BIND_ADDR"`
	ElasticSearchAPIURL     string        `envconfig:"ELASTIC_SEARCH_URL"`
	GracefulShutdownTimeout time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	HealthcheckInterval     time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	HealthcheckTimeout      time.Duration `envconfig:"HEALTHCHECK_TIMEOUT"`
	MaxRetries              int           `envconfig:"REQUEST_MAX_RETRIES"`
}

var cfg *Config

// Get configures the application and returns the configuration
func Get() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	cfg = &Config{
		BindAddr:                ":23100",
		ElasticSearchAPIURL:     "http://localhost:9200",
		GracefulShutdownTimeout: 5 * time.Second,
		HealthcheckInterval:     time.Minute,
		HealthcheckTimeout:      2 * time.Second,
		MaxRetries:              3,
	}

	return cfg, envconfig.Process("", cfg)
}
