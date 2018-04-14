package api

import (
	"context"

	"github.com/ONSdigital/go-ns/clients/dataset"
)

// DatasetAPIer - An interface used to access the DatasetAPI
type DatasetAPIer interface {
	GetVersion(ctx context.Context, datasetID, edition, version string) (dataset.Version, error)
	Healthcheck() (string, error)
}
