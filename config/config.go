package config

import (
	"encoding/json"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config is the filing resource handler config
type Config struct {
	AuthAPIURL                string        `envconfig:"ZEBEDEE_URL"`
	BindAddr                  string        `envconfig:"BIND_ADDR"                  json:"-"`
	Brokers                   []string      `envconfig:"KAFKA_ADDR"                 json:"-"`
	DatasetAPIURL             string        `envconfig:"DATASET_API_URL"`
	DatasetAPIAuthToken       string        `envconfig:"DATASET_API_AUTH_TOKEN"     json:"-"`
	ElasticSearchAPIURL       string        `envconfig:"ELASTIC_SEARCH_URL"         json:"-"`
	GracefulShutdownTimeout   time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	HasPrivateEndpoints       bool          `envconfig:"ENABLE_PRIVATE_ENDPOINTS"`
	HealthCheckInterval       time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	HealthCheckTimeout        time.Duration `envconfig:"HEALTHCHECK_TIMEOUT"`
	HierarchyBuiltTopic       string        `envconfig:"HIERARCHY_BUILT_TOPIC"`
	KafkaMaxBytes             int           `envconfig:"KAFKA_MAX_BYTES"`
	MaxRetries                int           `envconfig:"REQUEST_MAX_RETRIES"`
	MaxSearchResultsOffset    int           `envconfig:"MAX_SEARCH_RESULTS_OFFSET"`
	SearchAPIURL              string        `envconfig:"SEARCH_API_URL"`
	ServiceAuthToken          string        `envconfig:"SERVICE_AUTH_TOKEN"         json:"-"`
	SignElasticsearchRequests bool          `envconfig:"SIGN_ELASTICSEARCH_REQUESTS"`
}

var cfg *Config

// Get configures the application and returns the configuration
func Get() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	cfg = &Config{
		AuthAPIURL:                "http://localhost:8082",
		BindAddr:                  ":23100",
		Brokers:                   []string{"localhost:9092"},
		DatasetAPIURL:             "http://localhost:22000",
		DatasetAPIAuthToken:       "FD0108EA-825D-411C-9B1D-41EF7727F465",
		ElasticSearchAPIURL:       "http://localhost:9200",
		GracefulShutdownTimeout:   5 * time.Second,
		HasPrivateEndpoints:       false,
		HealthCheckInterval:       1 * time.Minute,
		HealthCheckTimeout:        2 * time.Second,
		HierarchyBuiltTopic:       "hierarchy-built",
		KafkaMaxBytes:             2000000,
		MaxRetries:                3,
		MaxSearchResultsOffset:    1000,
		SearchAPIURL:              "http://localhost:23100",
		ServiceAuthToken:          "SD0108EA-825D-411C-45J3-41EF7727F123",
		SignElasticsearchRequests: false,
	}

	return cfg, envconfig.Process("", cfg)
}

// String is implemented to prevent sensitive fields being logged.
// The config is returned as JSON with sensitive fields omitted.
func (config Config) String() string {
	json, _ := json.Marshal(config)
	return string(json)
}
