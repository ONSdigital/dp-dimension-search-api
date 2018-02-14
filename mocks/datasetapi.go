package mocks

import (
	"context"
	"errors"

	"github.com/ONSdigital/dp-dataset-api/models"
	datasetclient "github.com/ONSdigital/go-ns/clients/dataset"
)

// DatasetAPI represents a list of error flags to set error in mocked dataset API
type DatasetAPI struct {
	InternalServerError bool
	VersionNotFound     bool
}

var (
	errorInternalServer = errors.New("Internal server error")
	errorNotFound       = errors.New("Not found")
)

// GetVersion represents the mocked version that queries the dataset API to get a version resource
func (api *DatasetAPI) GetVersion(ctx context.Context, datasetID, edition, version, authToken string) (*models.Version, error) {
	if api.InternalServerError {
		return nil, errorInternalServer
	}

	if api.VersionNotFound {
		return nil, errorNotFound
	}

	return &models.Version{}, nil
}

// GetHealthCheckClient represents the mocked version of the healthcheck client
func (api *DatasetAPI) GetHealthCheckClient() *datasetclient.Client {
	return &datasetclient.Client{}
}