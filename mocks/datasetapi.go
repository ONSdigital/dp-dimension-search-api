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
	RequireAuth         bool
	RequireNoAuth       bool
}

var (
	errorInternalServer = errors.New("Internal server error")
	errorNotFound       = errors.New("Not found")
)

// GetVersion represents the mocked version that queries the dataset API to get a version resource
func (api *DatasetAPI) GetVersion(ctx context.Context, datasetID, edition, version, authToken string) (*models.Version, error) {
	isAuthenticated := len(authToken) > 0
	isBadAuthExpectation := (api.RequireNoAuth && isAuthenticated) || (api.RequireAuth && !isAuthenticated)

	if api.InternalServerError {
		if isBadAuthExpectation {
			return nil, errorNotFound
		}
		return nil, errorInternalServer
	}

	if api.VersionNotFound {
		if isBadAuthExpectation {
			return nil, errorInternalServer
		}
		return nil, errorNotFound
	}

	if isBadAuthExpectation {
		return nil, errorNotFound
	}
	return &models.Version{}, nil
}

// GetHealthCheckClient represents the mocked version of the healthcheck client
func (api *DatasetAPI) GetHealthCheckClient() *datasetclient.Client {
	return &datasetclient.Client{}
}
