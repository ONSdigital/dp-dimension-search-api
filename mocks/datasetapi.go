package mocks

import (
	"context"
	"errors"

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

var (
	errorInternalServer = errors.New("Internal server error")
	errorNotFound       = errors.New("Not found")
)

// GetVersion represents the mocked version that queries the dataset API to get a version resource
func (api *DatasetAPI) GetVersion(ctx context.Context, datasetID, edition, version string) (ver datasetclient.Version, err error) {
	isAuthenticated := len(api.SvcAuth) > 0
	isBadAuthExpectation := (api.RequireNoAuth && isAuthenticated) || (api.RequireAuth && !isAuthenticated)
	api.Calls++

	if api.InternalServerError {
		if isBadAuthExpectation {
			return ver, errorNotFound
		}
		return ver, errorInternalServer
	}

	if api.VersionNotFound {
		if isBadAuthExpectation {
			return ver, errorInternalServer
		}
		return ver, errorNotFound
	}

	if isBadAuthExpectation {
		return ver, errorNotFound
	}

	return
}

// Healthcheck represents the mocked version of the healthcheck
func (api *DatasetAPI) Healthcheck() (string, error) {
	return "healthcheckID", nil
}
