package mocks

import (
	"context"
	"net/http"

	errs "github.com/ONSdigital/dp-dimension-search-api/apierrors"
	"github.com/ONSdigital/dp-dimension-search-api/models"
)

// Elasticsearch represents a list of error flags to set error in mocked elasticsearch
type Elasticsearch struct {
	InternalServerError bool
	IndexNotFound       bool
}

// QuerySearchIndex represents the mocked version of building a query and then calling elasticsearch index
func (api *Elasticsearch) QuerySearchIndex(ctx context.Context, instanceID, dimension, term string, limit, offset int) (*models.SearchResponse, int, error) {
	if api.InternalServerError {
		return nil, 0, errs.ErrInternalServer
	}

	if api.IndexNotFound {
		return nil, http.StatusInternalServerError, errs.ErrIndexNotFound
	}

	firstHit := models.HitList{
		Highlight: models.Highlight{
			Code:  []string{"\u0001Sfrs34g5t98hdd\u0001E"},
			Label: []string{"\u0001Ssomething\u0001Eand\u0001Ssomeone\u0001E"},
		},
		Score: 3.0678,
		Source: models.SearchResult{
			Code:             "frs34g5t98hdd",
			URL:              "http://localhost:8080/testing/1",
			HasData:          true,
			Label:            "something and someone",
			NumberOfChildren: 3,
		},
	}

	secondHit := models.HitList{
		Highlight: models.Highlight{
			Label: []string{"\u0001Ssomething\u0001E else and\u0001Ssomeone\u0001E else"},
		},
		Score: 2.9782,
		Source: models.SearchResult{
			Code:             "gt534g5t98hs1",
			URL:              "http://localhost:8080/testing/2",
			HasData:          false,
			Label:            "something else and someone else",
			NumberOfChildren: 10,
		},
	}

	return &models.SearchResponse{
		Hits: models.Hits{
			Total: models.Total{
				Value:    22,
				Relation: "eq",
			},
			HitList: []models.HitList{firstHit, secondHit},
		},
	}, http.StatusOK, nil
}

// DeleteSearchIndex represents the mocked version that removes an index from elasticsearch
func (api *Elasticsearch) DeleteSearchIndex(ctx context.Context, instanceID, dimension string) (int, error) {
	if api.InternalServerError {
		return 0, errs.ErrInternalServer
	}

	if api.IndexNotFound {
		return http.StatusNotFound, errs.ErrDeleteIndexNotFound
	}

	return http.StatusOK, nil
}
