package service

import (
	"fmt"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	kafka "github.com/ONSdigital/dp-kafka"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/net/context"

	"github.com/ONSdigital/dp-search-api/api"
	"github.com/ONSdigital/dp-search-api/searchoutputqueue"
	"github.com/ONSdigital/go-ns/audit"

	"github.com/ONSdigital/log.go/log"
)

// Service represents the necessary config for dp-search-api
type Service struct {
	Auditor                    audit.AuditorService
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
	SearchIndexProducer        *kafka.Producer
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
		svc.Auditor,
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
	if err := svc.SearchIndexProducer.Close(ctx); err != nil {
		log.Event(ctx, "error while attempting to shutdown kafka producer", log.ERROR, log.Error(err))
	}

	svc.HealthCheck.Stop()

	log.Event(ctx, "shutdown complete", log.INFO)

	cancel()
	os.Exit(1)
}
