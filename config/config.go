package config

import (
	"encoding/json"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config is the filing resource handler config
type Config struct {
	AuthAPIURL                 string        `envconfig:"ZEBEDEE_URL"`
	AwsRegion                  string        `envconfig:"AWS_REGION"`
	AwsService                 string        `envconfig:"AWS_SERVICE"`
	BindAddr                   string        `envconfig:"BIND_ADDR"`
	Brokers                    []string      `envconfig:"KAFKA_ADDR"                 json:"-"`
	DatasetAPIURL              string        `envconfig:"DATASET_API_URL"`
	ElasticSearchAPIURL        string        `envconfig:"ELASTIC_SEARCH_URL"         json:"-"`
	GracefulShutdownTimeout    time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	HasPrivateEndpoints        bool          `envconfig:"ENABLE_PRIVATE_ENDPOINTS"`
	HealthCheckInterval        time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	HealthCheckCriticalTimeout time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
	HierarchyBuiltTopic        string        `envconfig:"HIERARCHY_BUILT_TOPIC"`
	KafkaMaxBytes              int           `envconfig:"KAFKA_MAX_BYTES"`
	KafkaVersion               string        `envconfig:"KAFKA_VERSION"`
	MaxRetries                 int           `envconfig:"REQUEST_MAX_RETRIES"`
	MaxSearchResultsOffset     int           `envconfig:"MAX_SEARCH_RESULTS_OFFSET"`
	SearchAPIURL               string        `envconfig:"SEARCH_API_URL"`
	ServiceAuthToken           string        `envconfig:"SERVICE_AUTH_TOKEN"         json:"-"`
	SignElasticsearchRequests  bool          `envconfig:"SIGN_ELASTICSEARCH_REQUESTS"`
}

var cfg *Config

// Get configures the application and returns the configuration
func Get() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	cfg = &Config{
		AuthAPIURL:                 "http://localhost:8082",
		AwsRegion:                  "eu-west-1",
		AwsService:                 "es",
		BindAddr:                   ":23100",
		Brokers:                    []string{"localhost:9092"},
		DatasetAPIURL:              "http://localhost:22000",
		ElasticSearchAPIURL:        "http://localhost:10200",
		GracefulShutdownTimeout:    5 * time.Second,
		HasPrivateEndpoints:        true,
		HealthCheckInterval:        30 * time.Second,
		HealthCheckCriticalTimeout: 90 * time.Second,
		HierarchyBuiltTopic:        "hierarchy-built",
		KafkaMaxBytes:              2000000,
		KafkaVersion:               "1.0.2",
		MaxRetries:                 3,
		MaxSearchResultsOffset:     1000,
		SearchAPIURL:               "http://localhost:23100",
		ServiceAuthToken:           "a507f722-f25a-4889-9653-23a2655b925c",
		SignElasticsearchRequests:  false,
	}

	return cfg, envconfig.Process("", cfg)
}

// String is implemented to prevent sensitive fields being logged.
// The config is returned as JSON with sensitive fields omitted.
func (config Config) String() string {
	json, _ := json.Marshal(config)
	return string(json)
}
