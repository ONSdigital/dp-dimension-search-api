package api

import (
	"context"
	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	rchttp "github.com/ONSdigital/dp-rchttp"

	identityclient "github.com/ONSdigital/dp-api-clients-go/identity"
	"github.com/ONSdigital/dp-search-api/models"
	"github.com/ONSdigital/dp-search-api/searchoutputqueue"
	"github.com/ONSdigital/go-ns/audit"
	"github.com/ONSdigital/go-ns/identity"
	"github.com/ONSdigital/go-ns/server"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
)

var httpServer *server.Server

type DatasetAPIClient interface {
	GetVersion(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceAuthToken, collectionID, datasetID, edition, version string) (m dataset.Version, err error)
}

// DownloadsGenerator pre generates full file downloads for the specified dataset/edition/version
type DownloadsGenerator interface {
	Generate(datasetID, instanceID, edition, version string) error
}

// OutputQueue - An interface used to queue search outputs
type OutputQueue interface {
	Queue(output *searchoutputqueue.Search) error
}

type HealthCheck interface {
	Start(ctx context.Context)
	Stop()
}

// SearchAPI manages searches across indices
type SearchAPI struct {
	auditor             audit.AuditorService
	datasetAPIClient    DatasetAPIClient
	serviceAuthToken    string
	defaultMaxResults   int
	elasticsearch       Elasticsearcher
	hasPrivateEndpoints bool
	host                string
	internalToken       string
	router              *mux.Router
	searchOutputQueue   OutputQueue
}

// CreateSearchAPI manages all the routes configured to API
func CreateSearchAPI(ctx context.Context,
	host, bindAddr, authAPIURL string, errorChan chan error, searchOutputQueue OutputQueue,
	datasetAPIClient DatasetAPIClient, serviceAuthToken string, elasticsearch Elasticsearcher,
	defaultMaxResults int, hasPrivateEndpoints bool, auditor audit.AuditorService,
	healthCheck *healthcheck.HealthCheck) {

	router := mux.NewRouter()
	routes(host,
		router,
		searchOutputQueue,
		datasetAPIClient,
		serviceAuthToken,
		elasticsearch,
		defaultMaxResults,
		hasPrivateEndpoints,
		auditor,
		healthCheck)

	middlewareChain := alice.New()

	if hasPrivateEndpoints {
		log.Event(ctx, "private endpoints are enabled. using identity middleware", log.INFO)
		identityHTTPClient := rchttp.NewClient()
		identityClient := identityclient.NewAPIClient(identityHTTPClient, authAPIURL)

		identityHandler := identity.HandlerForHTTPClient(identityClient)
		middlewareChain = middlewareChain.Append(identityHandler)
	}

	alice := middlewareChain.Then(router)
	httpServer = server.New(bindAddr, alice)

	// Disable this here to allow service to manage graceful shutdown of the entire app.
	httpServer.HandleOSSignals = false

	go func() {
		log.Event(ctx, "Starting api...", log.INFO)
		if err := httpServer.ListenAndServe(); err != nil {
			log.Event(ctx, "api http server returned error", log.ERROR, log.Error(err))
			errorChan <- err
		}
	}()
}

func routes(host string,
	router *mux.Router,
	searchOutputQueue OutputQueue,
	datasetAPIClient DatasetAPIClient,
	serviceAuthToken string,
	elasticsearch Elasticsearcher,
	defaultMaxResults int,
	hasPrivateEndpoints bool,
	auditor audit.AuditorService,
	healthCheck *healthcheck.HealthCheck) *SearchAPI {

	api := SearchAPI{
		auditor:             auditor,
		datasetAPIClient:    datasetAPIClient,
		serviceAuthToken:    serviceAuthToken,
		defaultMaxResults:   defaultMaxResults,
		elasticsearch:       elasticsearch,
		hasPrivateEndpoints: hasPrivateEndpoints,
		searchOutputQueue:   searchOutputQueue,
		host:                host,
		router:              router,
	}

	api.router.HandleFunc("/health", healthCheck.Handler)
	api.router.HandleFunc("/search/datasets/{id}/editions/{edition}/versions/{version}/dimensions/{name}", api.getSearch).Methods("GET")

	if hasPrivateEndpoints {
		api.router.HandleFunc("/search/instances/{instance_id}/dimensions/{dimension}", identity.Check(auditor, models.AuditTaskCreateIndex, api.createSearchIndex)).Methods("PUT")
		api.router.HandleFunc("/search/instances/{instance_id}/dimensions/{dimension}", identity.Check(auditor, models.AuditTaskDeleteIndex, api.deleteSearchIndex)).Methods("DELETE")
	}

	return &api
}

// Close represents the graceful shutting down of the http server
func Close(ctx context.Context) error {
	if err := httpServer.Shutdown(ctx); err != nil {
		return err
	}
	log.Event(ctx, "graceful shutdown of http server complete", log.INFO)
	return nil
}
