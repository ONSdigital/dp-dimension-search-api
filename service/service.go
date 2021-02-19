package service

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	kafka "github.com/ONSdigital/dp-kafka"

	"golang.org/x/net/context"

	"github.com/ONSdigital/dp-dimension-search-api/api"
	"github.com/ONSdigital/dp-dimension-search-api/searchoutputqueue"

	"github.com/ONSdigital/log.go/log"
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
	ServiceAuthToken           string
	Shutdown                   time.Duration
	SignElasticsearchRequests  bool
	HasPrivateEndpoints        bool
}

// Start handles consumption of events
func (svc *Service) Start(ctx context.Context) {

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	apiErrors := make(chan error, 1)

	svc.HealthCheck.Start(ctx)

	api.CreateSearchAPI(
		ctx,
		svc.SearchAPIURL,
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
	)

	go func() {
		for {
			select {
			case err := <-apiErrors:
				log.Event(ctx, "api error received", log.ERROR, log.Error(err))
			}
		}
	}()

	<-signals
	log.Event(ctx, "os signal received", log.INFO)

	// Gracefully shutdown the application closing any open resources.
	log.Event(ctx, fmt.Sprintf("shutdown with timeout: %s", svc.Shutdown), log.INFO)
	ctx, cancel := context.WithTimeout(context.Background(), svc.Shutdown)

	// stop any incoming requests before closing any outbound connections
	api.Close(ctx)
	svc.HealthCheck.Stop()

	if err := svc.HierarchyBuiltProducer.Close(ctx); err != nil {
		log.Event(ctx, "error while attempting to shutdown hierarchy built kafka producer", log.ERROR, log.Error(err))
	}

	log.Event(ctx, "shutdown complete", log.INFO)

	cancel()
	os.Exit(0)
}
