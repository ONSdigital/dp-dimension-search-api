package main

import (
	"context"
	"errors"
	"os"

	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-api-clients-go/zebedee"
	esauth "github.com/ONSdigital/dp-elasticsearch/v2/awsauth"
	elastic "github.com/ONSdigital/dp-elasticsearch/v2/elasticsearch"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/http"

	"github.com/ONSdigital/dp-dimension-search-api/config"
	"github.com/ONSdigital/dp-dimension-search-api/elasticsearch"
	"github.com/ONSdigital/dp-dimension-search-api/searchoutputqueue"
	"github.com/ONSdigital/dp-dimension-search-api/service"
	kafka "github.com/ONSdigital/dp-kafka/v4"
	dpotelgo "github.com/ONSdigital/dp-otel-go"
	"github.com/ONSdigital/log.go/v2/log"
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
	log.Namespace = "dp-dimension-search-api"

	ctx := context.Background()

	cfg, err := config.Get()
	if err != nil {
		log.Fatal(ctx, "failed to retrieve configuration", err)
	}

	// sensitive fields are omitted from config.String().
	log.Info(ctx, "config on startup", log.Data{"config": cfg})

	var esSigner *esauth.Signer
	if cfg.SignElasticsearchRequests {
		esSigner, err = esauth.NewAwsSigner("", "", cfg.AwsRegion, cfg.AwsService)
		if err != nil {
			log.Error(ctx, "failed to create aws v4 signer", err)
			os.Exit(1)
		}
	}

	// Set up OpenTelemetry
	otelConfig := dpotelgo.Config{
		OtelServiceName:          cfg.OTServiceName,
		OtelExporterOtlpEndpoint: cfg.OTExporterOTLPEndpoint,
		OtelBatchTimeout:         cfg.OTBatchTimeout,
	}

	otelShutdown, oErr := dpotelgo.SetupOTelSDK(ctx, otelConfig)
	if oErr != nil {
		log.Error(ctx, "error setting up OpenTelemetry - hint: ensure OTEL_EXPORTER_OTLP_ENDPOINT is set", oErr)
	}
	// Handle shutdown properly so nothing leaks.
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	elasticHTTPClient := dphttp.NewClient()
	elasticsearch := elasticsearch.NewElasticSearchAPI(elasticHTTPClient, cfg.ElasticSearchAPIURL, cfg.SignElasticsearchRequests, esSigner, cfg.AwsService, cfg.AwsRegion)

	pConfig := &kafka.ProducerConfig{
		KafkaVersion:    &cfg.KafkaVersion,
		MaxMessageBytes: &cfg.KafkaMaxBytes,
		BrokerAddrs:     cfg.Brokers,
		Topic:           cfg.HierarchyBuiltTopic,
	}
	if cfg.KafkaSecProtocol == "TLS" {
		pConfig.SecurityConfig = kafka.GetSecurityConfig(
			cfg.KafkaSecCACerts,
			cfg.KafkaSecClientCert,
			cfg.KafkaSecClientKey,
			cfg.KafkaSecSkipVerify,
		)
	}
	hierarchyBuiltProducer, err := kafka.NewProducer(
		ctx,
		pConfig,
	)
	exitIfError(ctx, err, "error creating kafka hierarchyBuiltProducer")

	hierarchyBuiltProducer.LogErrors(ctx)

	outputQueue := searchoutputqueue.CreateOutputQueue(hierarchyBuiltProducer.Channels().Output)

	datasetAPIClient := dataset.NewAPIClient(cfg.DatasetAPIURL)

	hc := configureHealthChecks(ctx, cfg, elasticHTTPClient, esSigner, hierarchyBuiltProducer, datasetAPIClient)

	svc := &service.Service{
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
		HierarchyBuiltProducer:    hierarchyBuiltProducer,
		OTServiceName:             cfg.OTServiceName,
		ServiceAuthToken:          cfg.ServiceAuthToken,
		Shutdown:                  cfg.GracefulShutdownTimeout,
		SignElasticsearchRequests: cfg.SignElasticsearchRequests,
		EnableURLRewriting:        cfg.EnableURLRewriting,
	}

	svc.Start(ctx)
}

func configureHealthChecks(ctx context.Context,
	cfg *config.Config,
	elasticHTTPClient dphttp.Clienter,
	esSigner *esauth.Signer,
	producer *kafka.Producer,
	datasetAPIClient *dataset.Client) *healthcheck.HealthCheck {
	hasErrors := false

	versionInfo, err := healthcheck.NewVersionInfo(BuildTime, GitCommit, Version)
	if err != nil {
		log.Fatal(ctx, "error creating version info", err)
	}

	hc := healthcheck.New(versionInfo, cfg.HealthCheckCriticalTimeout, cfg.HealthCheckInterval)

	if err = hc.AddCheck("Dataset API", datasetAPIClient.Checker); err != nil {
		log.Error(ctx, "error creating dataset API health check", err)
		hasErrors = true
	}

	elasticClient := elastic.NewClientWithHTTPClientAndAwsSigner(cfg.ElasticSearchAPIURL, esSigner, cfg.SignElasticsearchRequests, elasticHTTPClient)
	if err = hc.AddCheck("Elasticsearch", elasticClient.Checker); err != nil {
		log.Error(ctx, "error creating elasticsearch health check", err)
		hasErrors = true
	}

	if err = hc.AddCheck("Kafka Producer", producer.Checker); err != nil {
		log.Error(ctx, "error adding check for kafka producer", err)
		hasErrors = true
	}

	if cfg.HasPrivateEndpoints {
		// zebedee is used only for identity checking
		zebedeeClient := zebedee.New(cfg.AuthAPIURL)
		if err = hc.AddCheck("Zebedee", zebedeeClient.Checker); err != nil {
			log.Error(ctx, "error creating zebedee health check", err)
			hasErrors = true
		}
	}

	if hasErrors {
		os.Exit(1)
	}

	return &hc
}

func exitIfError(ctx context.Context, err error, message string) {
	if err != nil {
		log.Fatal(ctx, message, err)
	}
}
