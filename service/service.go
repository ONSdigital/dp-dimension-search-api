package service

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/net/context"

	"github.com/ONSdigital/dp-search-api/api"
	"github.com/ONSdigital/go-ns/elasticsearch"
	"github.com/ONSdigital/go-ns/healthcheck"
	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/rchttp"
	"github.com/pkg/errors"
)

// Service represents the necessary config for dp-dimension-extractor
type Service struct {
	BindAddr            string
	DatasetAPI          api.DatasetAPIer
	DatasetAPISecretKey string
	DefaultMaxResults   int
	Elasticsearch       api.Elasticsearcher
	ElasticsearchURL    string
	EnvMax              int64
	HealthCheckInterval time.Duration
	HealthCheckTimeout  time.Duration
	HTTPClient          *rchttp.Client
	MaxRetries          int
	SearchAPIURL        string
	SearchIndexProducer kafka.Producer
	SecretKey           string
	Shutdown            time.Duration
}

// Start handles consumption of events
func (svc *Service) Start() {

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	apiErrors := make(chan error, 1)

	svc.HTTPClient = rchttp.DefaultClient

	healthTicker := healthcheck.NewTicker(
		svc.HealthCheckInterval,
		svc.DatasetAPI.GetHealthCheckClient(),
		elasticsearch.NewHealthCheckClient(svc.ElasticsearchURL),
	)

	api.CreateSearchAPI(svc.SearchAPIURL, svc.BindAddr, svc.SecretKey, svc.DatasetAPISecretKey, apiErrors, svc.SearchIndexProducer, svc.DatasetAPI, svc.Elasticsearch, svc.DefaultMaxResults)

	// Gracefully shutdown the application closing any open resources.
	gracefulShutdown := func() {
		log.Info(fmt.Sprintf("shutdown with timeout: %s", svc.Shutdown), nil)
		ctx, cancel := context.WithTimeout(context.Background(), svc.Shutdown)

		// stop any incoming requests before closing any outbound connections
		api.Close(ctx)

		if err := svc.SearchIndexProducer.Close(ctx); err != nil {
			log.Error(errors.Wrap(err, "error while attempting to shutdown kafka producer"), nil)
		}

		healthTicker.Close()

		log.Info("shutdown complete", nil)

		cancel()
		os.Exit(1)
	}

	for {
		select {
		case err := <-apiErrors:
			log.ErrorC("api error received", err, nil)
			gracefulShutdown()
		case <-signals:
			log.Debug("os signal received", nil)
			gracefulShutdown()
		}
	}
}
