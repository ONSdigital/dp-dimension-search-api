package api

import (
	"context"

	"github.com/ONSdigital/dp-dimension-search-api/models"
)

// Elasticsearcher - An interface used to access elasticsearch
type Elasticsearcher interface {
	DeleteSearchIndex(ctx context.Context, instanceID, dimension string) (int, error)
	QuerySearchIndex(ctx context.Context, instanceID, dimension, term string, limit, offset int) (*models.SearchResponse, int, error)
}
