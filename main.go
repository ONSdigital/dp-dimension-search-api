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

	kafka "github.com/ONSdigital/dp-kafka"
	"github.com/ONSdigital/dp-search-api/config"
	"github.com/ONSdigital/dp-search-api/elasticsearch"
	"github.com/ONSdigital/dp-search-api/searchoutputqueue"
	"github.com/ONSdigital/dp-search-api/service"
	"github.com/ONSdigital/go-ns/audit"
	"github.com/ONSdigital/log.go/log"
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
		log.Event(ctx, "failed to retrieve configuration", log.FATAL, log.Error(err))
		os.Exit(1)
	}

	// sensitive fields are omitted from config.String().
	log.Event(ctx, "config on startup", log.INFO, log.Data{"config": cfg})

	elasticHTTPClient := rchttp.NewClient()
	elasticsearch := elasticsearch.NewElasticSearchAPI(elasticHTTPClient, cfg.ElasticSearchAPIURL, cfg.SignElasticsearchRequests)
	_, status, err := elasticsearch.CallElastic(context.Background(), cfg.ElasticSearchAPIURL, "GET", nil)
	if err != nil {
		log.Event(ctx, "failed to start up, unable to connect to elastic search instance", log.FATAL, log.Error(err), log.Data{"http_status": status})
		os.Exit(1)
	}

	producerChannels := kafka.CreateProducerChannels()
	hierarchyBuiltProducer, err := kafka.NewProducer(ctx, cfg.Brokers, cfg.HierarchyBuiltTopic, cfg.KafkaMaxBytes, producerChannels)
	exitIfError(ctx, err, "error creating kafka hierarchyBuiltProducer")

	var auditor audit.AuditorService
	var auditProducer *kafka.Producer

	if cfg.HasPrivateEndpoints {
		log.Event(ctx, "private endpoints enabled, enabling action auditing", log.INFO, log.Data{"auditTopicName": cfg.AuditEventsTopic})

		auditProducerChannels := kafka.CreateProducerChannels()
		auditProducer, err = kafka.NewProducer(ctx, cfg.Brokers, cfg.AuditEventsTopic, 0, auditProducerChannels)
		exitIfError(ctx, err, "error creating kafka hierarchyBuiltProducer")

		auditProducerAdapter := kafkaadapter.NewProducerAdapter(auditProducer)
		auditor = audit.New(auditProducerAdapter, "dp-search-api")
	} else {
		log.Event(ctx, "private endpoints disabled, auditing will not be enabled", log.INFO)
		auditor = &audit.NopAuditor{}
	}

	outputQueue := searchoutputqueue.CreateOutputQueue(hierarchyBuiltProducer.Channels().Output)

	datasetAPIClient := dataset.NewAPIClient(cfg.DatasetAPIURL)

	hc := configureHealthChecks(ctx, cfg, elasticHTTPClient, hierarchyBuiltProducer, auditProducer, datasetAPIClient)

	svc := &service.Service{
		Auditor:                   auditor,
		AuthAPIURL:                cfg.AuthAPIURL,
		BindAddr:                  cfg.BindAddr,
		DatasetAPIClient:          datasetAPIClient,
		DefaultMaxResults:         cfg.MaxSearchResultsOffset,
		Elasticsearch:             elasticsearch,
		ElasticsearchURL:          cfg.ElasticSearchAPIURL,
		HasPrivateEndpoints:       cfg.HasPrivateEndpoints,
		HealthCheck:               hc,
		MaxRetries:                cfg.MaxRetries,
		OutputQueue:               outputQueue,
		SearchAPIURL:              cfg.SearchAPIURL,
		AuditProducer:             auditProducer,
		HierarchyBuiltProducer:    hierarchyBuiltProducer,
		ServiceAuthToken:          cfg.ServiceAuthToken,
		Shutdown:                  cfg.GracefulShutdownTimeout,
		SignElasticsearchRequests: cfg.SignElasticsearchRequests,
	}

	svc.Start(ctx)
}

func configureHealthChecks(ctx context.Context,
	cfg *config.Config,
	elasticHTTPClient rchttp.Clienter,
	producer *kafka.Producer,
	auditProducer *kafka.Producer,
	datasetAPIClient *dataset.Client) *healthcheck.HealthCheck {

	hasErrors := false

	versionInfo, err := healthcheck.NewVersionInfo(BuildTime, GitCommit, Version)
	if err != nil {
		log.Event(ctx, "error creating version info", log.FATAL, log.Error(err))
		hasErrors = true
	}

	hc := healthcheck.New(versionInfo, cfg.HealthCheckCriticalTimeout, cfg.HealthCheckInterval)

	if err = hc.AddCheck("Dataset API", datasetAPIClient.Checker); err != nil {
		log.Event(ctx, "error creating dataset API health check", log.Error(err))
		hasErrors = true
	}

	// zebedee is used only for identity checking
	zebedeeClient := zebedee.New(cfg.AuthAPIURL)
	if err = hc.AddCheck("Zebedee", zebedeeClient.Checker); err != nil {
		log.Event(ctx, "error creating zebedee health check", log.ERROR, log.Error(err))
		hasErrors = true
	}

	elasticClient := elastic.NewClientWithHTTPClient(cfg.ElasticSearchAPIURL, cfg.SignElasticsearchRequests, elasticHTTPClient)
	if err = hc.AddCheck("Elasticsearch", elasticClient.Checker); err != nil {
		log.Event(ctx, "error creating elasticsearch health check", log.ERROR, log.Error(err))
		hasErrors = true
	}

	if err = hc.AddCheck("Kafka Producer", producer.Checker); err != nil {
		log.Event(ctx, "error adding check for kafka producer", log.ERROR, log.Error(err))
		hasErrors = true
	}

	if err = hc.AddCheck("Kafka Audit Producer", auditProducer.Checker); err != nil {
		log.Event(ctx, "error adding check for kafka audit producer", log.ERROR, log.Error(err))
		hasErrors = true
	}

	if hasErrors {
		os.Exit(1)
	}

	return &hc
}

func exitIfError(ctx context.Context, err error, message string) {
	if err != nil {
		log.Event(ctx, message, log.FATAL, log.Error(err))
		os.Exit(1)
	}
}
