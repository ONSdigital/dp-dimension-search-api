package main

import (
	"context"
	"os"

	"github.com/ONSdigital/dp-search-api/config"
	dataset "github.com/ONSdigital/dp-search-api/dataset"
	"github.com/ONSdigital/dp-search-api/elasticsearch"
	"github.com/ONSdigital/dp-search-api/service"
	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/rchttp"
	"github.com/pkg/errors"
)

func main() {
	log.Namespace = "dp-search-api"

	cfg, err := config.Get()
	if err != nil {
		log.Error(err, nil)
		os.Exit(1)
	}

	client := rchttp.DefaultClient
	elasticsearch := elasticsearch.NewElasticSearchAPI(client, cfg.ElasticSearchAPIURL)
	_, status, err := elasticsearch.CallElastic(context.Background(), cfg.ElasticSearchAPIURL, "GET", nil)
	if err != nil {
		log.ErrorC("failed to start up, unable to connect to elastic search instance", err, log.Data{"http_status": status})
		os.Exit(1)
	}

	createSearchIndexProducer, err := kafka.NewProducer(cfg.Brokers, cfg.HierarchyBuiltTopic, 0)
	if err != nil {
		log.Error(errors.Wrap(err, "error creating kakfa producer"), nil)
		os.Exit(1)
	}

	datasetAPI := dataset.NewDatasetAPI(client, cfg.DatasetAPIURL)

	// TODO This needs to be sanitised to hide sensitive configs
	log.Info("config on startup", log.Data{"config": cfg})

	svc := &service.Service{
		BindAddr:            cfg.BindAddr,
		DatasetAPI:          datasetAPI,
		DatasetAPISecretKey: cfg.DatasetAPISecretKey,
		DefaultMaxResults:   cfg.MaxSearchResultsOffset,
		Elasticsearch:       elasticsearch,
		ElasticsearchURL:    cfg.ElasticSearchAPIURL,
		EnvMax:              cfg.KafkaMaxBytes,
		HealthCheckInterval: cfg.HealthCheckInterval,
		HealthCheckTimeout:  cfg.HealthCheckTimeout,
		HTTPClient:          client,
		MaxRetries:          cfg.MaxRetries,
		SearchAPIURL:        cfg.SearchAPIURL,
		SearchIndexProducer: createSearchIndexProducer,
		SecretKey:           cfg.SecretKey,
		Shutdown:            cfg.GracefulShutdownTimeout,
	}

	svc.Start()
}
