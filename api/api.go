package api

import (
	"context"

	"github.com/ONSdigital/dp-dataset-api/store"
	"github.com/ONSdigital/dp-search-api/auth"
	"github.com/ONSdigital/dp-search-api/searchoutputqueue"
	"github.com/ONSdigital/go-ns/healthcheck"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/server"
	"github.com/gorilla/mux"
)

var httpServer *server.Server

// API provides an interface for the routes
type API interface {
	CreateSearchAPI(string, *mux.Router, store.DataStore) *SearchAPI
}

// DownloadsGenerator pre generates full file downloads for the specified dataset/edition/version
type DownloadsGenerator interface {
	Generate(datasetID, instanceID, edition, version string) error
}

// OutputQueue - An interface used to queue search outputs
type OutputQueue interface {
	Queue(output *searchoutputqueue.Search) error
}

// SearchAPI manages searches across indices
type SearchAPI struct {
	datasetAPI          DatasetAPIer
	datasetAPISecretKey string
	defaultMaxResults   int
	elasticsearch       Elasticsearcher
	hasPrivateEndpoints bool
	host                string
	internalToken       string
	privateAuth         *auth.Authenticator
	router              *mux.Router
	searchOutputQueue   OutputQueue
}

// CreateSearchAPI manages all the routes configured to API
func CreateSearchAPI(host, bindAddr, secretKey, datasetAPISecretKey string, errorChan chan error, searchOutputQueue OutputQueue, datasetAPI DatasetAPIer, elasticsearch Elasticsearcher, defaultMaxResults int, hasPrivateEndpoints bool) {
	router := mux.NewRouter()
	routes(host, secretKey, datasetAPISecretKey, router, searchOutputQueue, datasetAPI, elasticsearch, defaultMaxResults, hasPrivateEndpoints)

	httpServer = server.New(bindAddr, router)
	// Disable this here to allow service to manage graceful shutdown of the entire app.
	httpServer.HandleOSSignals = false

	go func() {
		log.Debug("Starting api...", nil)
		if err := httpServer.ListenAndServe(); err != nil {
			log.ErrorC("api http server returned error", err, nil)
			errorChan <- err
		}
	}()
}

func routes(host, secretKey, datasetAPISecretKey string, router *mux.Router, searchOutputQueue OutputQueue, datasetAPI DatasetAPIer, elasticsearch Elasticsearcher, defaultMaxResults int, hasPrivateEndpoints bool) *SearchAPI {
	api := SearchAPI{
		datasetAPI:          datasetAPI,
		datasetAPISecretKey: datasetAPISecretKey,
		defaultMaxResults:   defaultMaxResults,
		elasticsearch:       elasticsearch,
		hasPrivateEndpoints: hasPrivateEndpoints,
		searchOutputQueue:   searchOutputQueue,
		host:                host,
		internalToken:       secretKey,
		privateAuth:         &auth.Authenticator{SecretKey: secretKey, HeaderName: "internal-token"},
		router:              router,
	}

	router.Path("/healthcheck").Methods("GET").HandlerFunc(healthcheck.Do)

	api.router.HandleFunc("/search/datasets/{id}/editions/{edition}/versions/{version}/dimensions/{name}", api.getSearch).Methods("GET")

	if hasPrivateEndpoints {
		api.router.HandleFunc("/search/instances/{instance_id}/dimensions/{dimension}", api.privateAuth.Check(api.createSearchIndex)).Methods("PUT")
		api.router.HandleFunc("/search/instances/{instance_id}/dimensions/{dimension}", api.privateAuth.Check(api.deleteSearchIndex)).Methods("DELETE")
	}

	return &api
}

// Close represents the graceful shutting down of the http server
func Close(ctx context.Context) error {
	if err := httpServer.Shutdown(ctx); err != nil {
		return err
	}
	log.Info("graceful shutdown of http server complete", nil)
	return nil
}
