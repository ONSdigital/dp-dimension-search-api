package mocks

import (
	"context"

	"github.com/ONSdigital/dp-api-clients-go/dataset"
	errs "github.com/ONSdigital/dp-dimension-search-api/apierrors"
)

// DatasetAPI represents a list of error flags to set error in mocked dataset API
type DatasetAPI struct {
	InternalServerError bool
	VersionNotFound     bool
	RequireAuth         bool
	RequireNoAuth       bool
	Calls               int
	IsAuthenticated     bool
}

// GetVersion represents the mocked version that queries the dataset API to get a version resource
func (api *DatasetAPI) GetVersion(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceAuthToken, collectionID, datasetID, edition, version string) (ver dataset.Version, err error) {
	api.IsAuthenticated = len(serviceAuthToken) > 0
	isBadAuthExpectation := (api.RequireNoAuth && api.IsAuthenticated) || (api.RequireAuth && !api.IsAuthenticated)
	api.Calls++

	if api.InternalServerError {
		if isBadAuthExpectation {
			return ver, errs.ErrVersionNotFound
		}
		return ver, errs.ErrInternalServer
	}

	if api.VersionNotFound {
		if isBadAuthExpectation {
			return ver, errs.ErrInternalServer
		}
		return ver, errs.ErrVersionNotFound
	}

	if isBadAuthExpectation {
		return ver, errs.ErrVersionNotFound
	}

	return
}

// Healthcheck represents the mocked version of the healthcheck
func (api *DatasetAPI) Healthcheck() (string, error) {
	return "healthcheckID", nil
}
