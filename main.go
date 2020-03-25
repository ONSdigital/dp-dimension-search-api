package main

import (
	"context"
	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-api-clients-go/zebedee"
	elastic "github.com/ONSdigital/dp-elasticsearch"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	rchttp "github.com/ONSdigital/dp-rchttp"
	"github.com/ONSdigital/dp-search-api/kafkaadapter"
	"os"

	"github.com/ONSdigital/dp-search-api/config"
	"github.com/ONSdigital/dp-search-api/elasticsearch"
	"github.com/ONSdigital/dp-search-api/searchoutputqueue"
	"github.com/ONSdigital/dp-search-api/service"
	"github.com/ONSdigital/go-ns/audit"
	"github.com/ONSdigital/go-ns/log"
	"github.com/pkg/errors"

	kafka "github.com/ONSdigital/dp-kafka"
)

var (
	// BuildTime represents the time in which the service was built
	BuildTime string
	// GitCommit represents the commit (SHA-1) hash of the service that is running
	GitCommit string
	// Version represents the version of the service that is running
	Version string
)

func main() {
	log.Namespace = "dp-search-api"

	ctx := context.Background()

	cfg, err := config.Get()
	if err != nil {
		log.Error(err, nil)
		os.Exit(1)
	}

	// sensitive fields are omitted from config.String().
	log.Info("config on startup", log.Data{"config": cfg})

	elasticHTTPClient := rchttp.NewClient()
	elasticsearch := elasticsearch.NewElasticSearchAPI(elasticHTTPClient, cfg.ElasticSearchAPIURL, cfg.SignElasticsearchRequests)
	_, status, err := elasticsearch.CallElastic(context.Background(), cfg.ElasticSearchAPIURL, "GET", nil)
	if err != nil {
		log.ErrorC("failed to start up, unable to connect to elastic search instance", err, log.Data{"http_status": status})
		os.Exit(1)
	}

	producerChannels := kafka.CreateProducerChannels()
	kafkaProducer, err := kafka.NewProducer(ctx, cfg.Brokers, cfg.HierarchyBuiltTopic, cfg.KafkaMaxBytes, producerChannels)
	if err != nil {
		log.ErrorC("error creating kafka producer", err, nil)
		os.Exit(1)
	}

	var auditor audit.AuditorService
	var auditProducer *kafka.Producer

	if cfg.HasPrivateEndpoints {
		log.Info("private endpoints enabled, enabling action auditing", log.Data{"auditTopicName": cfg.AuditEventsTopic})

		auditProducerChannels := kafka.CreateProducerChannels()
		auditProducer, err = kafka.NewProducer(ctx, cfg.Brokers, cfg.AuditEventsTopic, 0, auditProducerChannels)
		if err != nil {
			log.Error(errors.Wrap(err, "error creating kakfa audit producer"), nil)
			os.Exit(1)
		}

		auditProducerAdapter := kafkaadapter.NewProducerAdapter(auditProducer)
		auditor = audit.New(auditProducerAdapter, "dp-search-api")
	} else {
		log.Info("private endpoints disabled, auditing will not be enabled", nil)
		auditor = &audit.NopAuditor{}
	}

	versionInfo, err := healthcheck.NewVersionInfo(BuildTime, GitCommit, Version)
	if err != nil {
		log.ErrorC("error creating kafka producer", err, nil)
		os.Exit(1)
	}
	exitIfError(err, "error creating version info")
	hc := healthcheck.New(versionInfo, cfg.HealthCheckCriticalTimeout, cfg.HealthCheckInterval)

	datasetAPIClient := dataset.NewAPIClient(cfg.DatasetAPIURL)
	if err = hc.AddCheck("Dataset API", datasetAPIClient.Checker); err != nil {
		log.ErrorC("error creating dataset API health check", err, nil)
	}

	// zebedee is used only for identity checking
	zebedeeClient := zebedee.New(cfg.AuthAPIURL)
	if err = hc.AddCheck("Zebedee", zebedeeClient.Checker); err != nil {
		log.ErrorC("error creating zebedee health check", err, nil)
	}

	elasticClient := elastic.NewClientWithHTTPClient(cfg.ElasticSearchAPIURL, cfg.SignElasticsearchRequests, elasticHTTPClient)
	if err = hc.AddCheck("Elasticsearch", elasticClient.Checker); err != nil {
		log.ErrorC("error creating elasticsearch health check", err, nil)
	}

	outputQueue := searchoutputqueue.CreateOutputQueue(kafkaProducer.Channels().Output)

	svc := &service.Service{
		Auditor:                   auditor,
		AuthAPIURL:                cfg.AuthAPIURL,
		BindAddr:                  cfg.BindAddr,
		DatasetAPIClient:          datasetAPIClient,
		DefaultMaxResults:         cfg.MaxSearchResultsOffset,
		Elasticsearch:             elasticsearch,
		ElasticsearchURL:          cfg.ElasticSearchAPIURL,
		HasPrivateEndpoints:       cfg.HasPrivateEndpoints,
		HealthCheck:               &hc,
		MaxRetries:                cfg.MaxRetries,
		OutputQueue:               outputQueue,
		SearchAPIURL:              cfg.SearchAPIURL,
		SearchIndexProducer:       kafkaProducer,
		ServiceAuthToken:          cfg.ServiceAuthToken,
		Shutdown:                  cfg.GracefulShutdownTimeout,
		SignElasticsearchRequests: cfg.SignElasticsearchRequests,
	}

	svc.Start(ctx)
}

func exitIfError(err error, message string) {
	if err != nil {
		log.ErrorC(message, err, nil)
		os.Exit(1)
	}
}
