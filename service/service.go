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
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/rchttp"
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

	api.CreateSearchAPI(svc.SearchAPIURL, svc.BindAddr, svc.SecretKey, svc.DatasetAPISecretKey, apiErrors, svc.DatasetAPI, svc.Elasticsearch, svc.DefaultMaxResults)

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
	healthTicker.Close()

	log.Info("shutdown complete", nil)

	cancel()
	os.Exit(1)
}
