package main

import (
	"context"
	"os"
	"strconv"

	"github.com/ONSdigital/dp-search-api/config"
	"github.com/ONSdigital/dp-search-api/dataset"
	"github.com/ONSdigital/dp-search-api/elasticsearch"
	"github.com/ONSdigital/dp-search-api/searchOutputQueue"
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

	// sensitive fields are omitted from config.String().
	log.Info("config on startup", log.Data{"config": cfg})

	client := rchttp.DefaultClient
	elasticsearch := elasticsearch.NewElasticSearchAPI(client, cfg.ElasticSearchAPIURL, cfg.SignElasticsearchRequests)
	_, status, err := elasticsearch.CallElastic(context.Background(), cfg.ElasticSearchAPIURL, "GET", nil)
	if err != nil {
		log.ErrorC("failed to start up, unable to connect to elastic search instance", err, log.Data{"http_status": status})
		os.Exit(1)
	}

	envMax, err := strconv.ParseInt(cfg.KafkaMaxBytes, 10, 32)
	if err != nil {
		log.ErrorC("encountered error parsing kafka max bytes", err, nil)
		os.Exit(1)
	}

	producer, err := kafka.NewProducer(cfg.Brokers, cfg.HierarchyBuiltTopic, int(envMax))
	if err != nil {
		log.Error(errors.Wrap(err, "error creating kafka producer"), nil)
		os.Exit(1)
	}

	datasetAPI := dataset.NewDatasetAPI(client, cfg.DatasetAPIURL)
	outputQueue := searchOutputQueue.CreateOutputQueue(producer.Output())

	svc := &service.Service{
		BindAddr:                  cfg.BindAddr,
		DatasetAPI:                datasetAPI,
		DatasetAPISecretKey:       cfg.DatasetAPISecretKey,
		DefaultMaxResults:         cfg.MaxSearchResultsOffset,
		Elasticsearch:             elasticsearch,
		ElasticsearchURL:          cfg.ElasticSearchAPIURL,
		HealthCheckInterval:       cfg.HealthCheckInterval,
		HealthCheckTimeout:        cfg.HealthCheckTimeout,
		HTTPClient:                client,
		MaxRetries:                cfg.MaxRetries,
		OutputQueue:               outputQueue,
		SearchAPIURL:              cfg.SearchAPIURL,
		SearchIndexProducer:       producer,
		SecretKey:                 cfg.SecretKey,
		Shutdown:                  cfg.GracefulShutdownTimeout,
		HasPrivateEndpoints:	   cfg.HasPrivateEndpoints,
		SignElasticsearchRequests: cfg.SignElasticsearchRequests
	}

	svc.Start()
}
