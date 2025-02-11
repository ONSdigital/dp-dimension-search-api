package service

import (
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	kafka "github.com/ONSdigital/dp-kafka/v4"

	"golang.org/x/net/context"

	"github.com/ONSdigital/dp-dimension-search-api/api"
	"github.com/ONSdigital/dp-dimension-search-api/searchoutputqueue"

	"github.com/ONSdigital/log.go/v2/log"
)

// Service represents the necessary config for dp-dimension-search-api
type Service struct {
	AuthAPIURL                 string
	BindAddr                   string
	DatasetAPIClient           api.DatasetAPIClient
	DefaultMaxResults          int
	Elasticsearch              api.Elasticsearcher
	ElasticsearchURL           string
	EnvMax                     int
	HealthCheck                *healthcheck.HealthCheck
	HealthCheckCriticalTimeout time.Duration
	MaxRetries                 int
	OutputQueue                searchoutputqueue.Output
	SearchAPIURL               string
	HierarchyBuiltProducer     *kafka.Producer
	OTServiceName              string
	ServiceAuthToken           string
	Shutdown                   time.Duration
	SignElasticsearchRequests  bool
	HasPrivateEndpoints        bool
	EnableURLRewriting         bool
}

// Start handles consumption of events
func (svc *Service) Start(ctx context.Context) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	apiErrors := make(chan error, 1)

	svc.HealthCheck.Start(ctx)
	SearchAPIURL, err := url.Parse(svc.SearchAPIURL)
	if err != nil {
		log.Fatal(ctx, "error parsing SearchAPIURL API URL", err, log.Data{"url": svc.SearchAPIURL})
		os.Exit(1)
	}

	api.CreateSearchAPI(
		ctx,
		SearchAPIURL,
		svc.BindAddr,
		svc.AuthAPIURL,
		apiErrors,
		&svc.OutputQueue,
		svc.DatasetAPIClient,
		svc.ServiceAuthToken,
		svc.Elasticsearch,
		svc.DefaultMaxResults,
		svc.HasPrivateEndpoints,
		svc.HealthCheck,
		svc.OTServiceName,
		svc.EnableURLRewriting,
	)

	go func() {
		for err := range apiErrors {
			log.Error(ctx, "api error received", err)
		}
	}()

	<-signals
	log.Info(ctx, "os signal received")

	// Gracefully shutdown the application closing any open resources.
	log.Info(ctx, fmt.Sprintf("shutdown with timeout: %s", svc.Shutdown))
	ctx, cancel := context.WithTimeout(context.Background(), svc.Shutdown)

	// stop any incoming requests before closing any outbound connections
	api.Close(ctx)
	svc.HealthCheck.Stop()

	if err := svc.HierarchyBuiltProducer.Close(ctx); err != nil {
		log.Error(ctx, "error while attempting to shutdown hierarchy built kafka producer", err)
	}

	log.Info(ctx, "shutdown complete")

	cancel()
	os.Exit(0)
}
