package service

import (
	"fmt"
	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/net/context"

	"github.com/ONSdigital/dp-search-api/api"
	"github.com/ONSdigital/dp-search-api/searchoutputqueue"
	"github.com/ONSdigital/go-ns/audit"
	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/log"
	"github.com/pkg/errors"
)

type HealthCheck interface {
	Start(ctx context.Context)
	Stop()
}

// Service represents the necessary config for dp-search-api
type Service struct {
	Auditor                    audit.AuditorService
	AuthAPIURL                 string
	BindAddr                   string
	DatasetAPIURL              string
	DefaultMaxResults          int
	Elasticsearch              api.Elasticsearcher
	ElasticsearchURL           string
	EnvMax                     int
	HealthCheck                HealthCheck
	HealthCheckCriticalTimeout time.Duration
	MaxRetries                 int
	OutputQueue                searchoutputqueue.Output
	SearchAPIURL               string
	SearchIndexProducer        kafka.Producer
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

	datasetAPIClient := dataset.NewAPIClient(svc.DatasetAPIURL)

	svc.HealthCheck.Start(ctx)

	api.CreateSearchAPI(
		svc.SearchAPIURL,
		svc.BindAddr,
		svc.AuthAPIURL,
		apiErrors,
		&svc.OutputQueue,
		datasetAPIClient,
		svc.ServiceAuthToken,
		svc.Elasticsearch,
		svc.DefaultMaxResults,
		svc.HasPrivateEndpoints,
		svc.Auditor,
	)

	// blocks until a fatal error occurs
	select {
	case err := <-apiErrors:
		log.ErrorC("api error received", err, nil)
	case <-signals:
		log.Debug("os signal received", nil)
	}

	// Gracefully shutdown the application closing any open resources.
	log.Info(fmt.Sprintf("shutdown with timeout: %s", svc.Shutdown), nil)
	ctx, cancel := context.WithTimeout(context.Background(), svc.Shutdown)

	// stop any incoming requests before closing any outbound connections
	api.Close(ctx)
	if err := svc.SearchIndexProducer.Close(ctx); err != nil {
		log.Error(errors.Wrap(err, "error while attempting to shutdown kafka producer"), nil)
	}

	svc.HealthCheck.Stop()

	log.Info("shutdown complete", nil)

	cancel()
	os.Exit(1)
}
