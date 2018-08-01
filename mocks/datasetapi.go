package mocks

import (
	"context"

	errs "github.com/ONSdigital/dp-search-api/apierrors"
	datasetclient "github.com/ONSdigital/go-ns/clients/dataset"
)

// DatasetAPI represents a list of error flags to set error in mocked dataset API
type DatasetAPI struct {
	InternalServerError bool
	VersionNotFound     bool
	RequireAuth         bool
	RequireNoAuth       bool
	SvcAuth             string
	Calls               int
}

// GetVersion represents the mocked version that queries the dataset API to get a version resource
func (api *DatasetAPI) GetVersion(ctx context.Context, datasetID, edition, version string) (ver datasetclient.Version, err error) {
	isAuthenticated := len(api.SvcAuth) > 0
	isBadAuthExpectation := (api.RequireNoAuth && isAuthenticated) || (api.RequireAuth && !isAuthenticated)
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
