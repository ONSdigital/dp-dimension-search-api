package api

import (
	"context"

	"github.com/ONSdigital/dp-dataset-api/models"
	datasetclient "github.com/ONSdigital/go-ns/clients/dataset"
)

// DatasetAPIer - An interface used to access the DatasetAPI
type DatasetAPIer interface {
	GetVersion(ctx context.Context, datasetID, edition, version, authToken string) (*models.Version, error)
	GetHealthCheckClient() *datasetclient.Client
}
