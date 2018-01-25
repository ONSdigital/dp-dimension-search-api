package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/ONSdigital/dp-search-api/models"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/mux"
)

type pageVariables struct {
	limit  int
	offset int
}

const (
	internalToken = "Internal-Token"
)

var (
	err error

	limit  = 20
	offset = 0
)

func (api *SearchAPI) getSearch(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	datasetID := vars["id"]
	edition := vars["edition"]
	version := vars["version"]
	dimension := vars["name"]

	term := r.FormValue("q")
	requestedLimit := r.FormValue("limit")
	requestedOffset := r.FormValue("offset")

	logData := log.Data{
		"dataset_id":       datasetID,
		"edition":          edition,
		"version":          version,
		"dimension":        dimension,
		"query_term":       term,
		"requested_limit":  requestedLimit,
		"requested_offset": requestedOffset,
	}

	log.Info("incoming request", logData)

	var authToken string

	if r.Header.Get(internalToken) == api.internalToken {
		// Authorised to search against an unpublished version
		authToken = api.datasetAPISecretKey
	}

	// Get instanceID from datasetAPI
	versionDoc, err := api.datasetAPI.GetVersion(context.Background(), datasetID, edition, version, authToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	logData["version_doc"] = versionDoc

	instanceID := versionDoc.ID
	logData["instance_id"] = instanceID

	if requestedLimit != "" {
		limit, err = strconv.Atoi(requestedLimit)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	if requestedOffset != "" {
		offset, err = strconv.Atoi(requestedOffset)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	page := &models.PageVariables{
		DefaultMaxResults: api.defaultMaxResults,
		Limit:             limit,
		Offset:            offset,
	}

	if err = page.ValidateQueryParameters(term); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	logData["limit"] = page.Limit
	logData["offset"] = page.Offset

	log.Info("just before calling query search index", logData)

	response, _, err := api.elasticsearch.QuerySearchIndex(context.Background(), instanceID, dimension, term, page.Limit, page.Offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Trace("nathan here", log.Data{"response": response})

	searchResults := &models.SearchResults{
		Count:      response.Hits.Total,
		Limit:      page.Limit,
		Offset:     page.Offset,
		TotalCount: response.Hits.Total,
	}

	for _, result := range response.Hits.HitList {
		result.Source.DimensionOptionURL = result.Source.URL
		result.Source.URL = ""

		// TODO Matches will need to be worked on here?

		doc := result.Source
		searchResults.Items = append(searchResults.Items, doc)
	}

	searchResults.Count = len(searchResults.Items)

	bytes, err := json.Marshal(searchResults)
	if err != nil {
		log.ErrorC("failed to marshal dataset resource into bytes", err, nil)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	setJSONContentType(w)
	_, err = w.Write(bytes)
	if err != nil {
		log.Error(err, nil)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	log.Debug("get all datasets", nil)
}

func setJSONContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}
