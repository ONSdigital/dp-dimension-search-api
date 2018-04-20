package main

import (
	"context"
	"os"

	"github.com/ONSdigital/dp-search-api/config"
	"github.com/ONSdigital/dp-search-api/elasticsearch"
	"github.com/ONSdigital/dp-search-api/searchoutputqueue"
	"github.com/ONSdigital/dp-search-api/service"
	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/rchttp"
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

	elasticClient := rchttp.NewClient()
	elasticsearch := elasticsearch.NewElasticSearchAPI(elasticClient, cfg.ElasticSearchAPIURL, cfg.SignElasticsearchRequests)
	_, status, err := elasticsearch.CallElastic(context.Background(), cfg.ElasticSearchAPIURL, "GET", nil)
	if err != nil {
		log.ErrorC("failed to start up, unable to connect to elastic search instance", err, log.Data{"http_status": status})
		os.Exit(1)
	}

	producer, err := kafka.NewProducer(cfg.Brokers, cfg.HierarchyBuiltTopic, cfg.KafkaMaxBytes)
	if err != nil {
		log.ErrorC("error creating kafka producer", err, nil)
		os.Exit(1)
	}

	outputQueue := searchoutputqueue.CreateOutputQueue(producer.Output())

	svc := &service.Service{
		AuthAPIURL:                cfg.AuthAPIURL,
		BindAddr:                  cfg.BindAddr,
		DatasetAPIURL:             cfg.DatasetAPIURL,
		DefaultMaxResults:         cfg.MaxSearchResultsOffset,
		Elasticsearch:             elasticsearch,
		ElasticsearchURL:          cfg.ElasticSearchAPIURL,
		HasPrivateEndpoints:       cfg.HasPrivateEndpoints,
		HealthCheckInterval:       cfg.HealthCheckInterval,
		HealthCheckTimeout:        cfg.HealthCheckTimeout,
		MaxRetries:                cfg.MaxRetries,
		OutputQueue:               outputQueue,
		SearchAPIURL:              cfg.SearchAPIURL,
		SearchIndexProducer:       producer,
		ServiceAuthToken:          cfg.ServiceAuthToken,
		Shutdown:                  cfg.GracefulShutdownTimeout,
		SignElasticsearchRequests: cfg.SignElasticsearchRequests,
	}

	svc.Start()
}
