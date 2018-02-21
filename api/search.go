package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/ONSdigital/dp-search-api/models"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/mux"
)

type pageVariables struct {
	limit  int
	offset int
}

const (
	internalTokenHeader = "Internal-Token"

	defaultLimit  = 20
	defaultOffset = 0
)

var (
	internalError = "Failed to process the request due to an internal error"
	notFoundError = "Resource not found"

	err error
)

func (api *SearchAPI) getSearch(w http.ResponseWriter, r *http.Request, hasAuth bool) {
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

	if hasAuth == true {
		// Authorised to search against an unpublished version
		authToken = api.datasetAPISecretKey
	}

	// Get instanceID from datasetAPI
	versionDoc, err := api.datasetAPI.GetVersion(r.Context(), datasetID, edition, version, authToken)
	if err != nil {
		setErrorCode(w, err)
		return
	}

	logData["version_doc"] = versionDoc

	instanceID := versionDoc.ID
	logData["instance_id"] = instanceID

	limit := defaultLimit
	if requestedLimit != "" {
		limit, err = strconv.Atoi(requestedLimit)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	offset := defaultOffset
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

	response, _, err := api.elasticsearch.QuerySearchIndex(r.Context(), instanceID, dimension, term, page.Limit, page.Offset)
	if err != nil {
		setErrorCode(w, err)
		return
	}

	searchResults := &models.SearchResults{
		Count:  response.Hits.Total,
		Limit:  page.Limit,
		Offset: page.Offset,
	}

	for _, result := range response.Hits.HitList {
		result.Source.DimensionOptionURL = result.Source.URL
		result.Source.URL = ""

		result = getSnippets(result)

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

func getSnippets(result models.HitList) models.HitList {

	if len(result.Highlight.Code) > 0 {
		highlightedCode := result.Highlight.Code[0]
		var prevEnd int
		logData := log.Data{}
		for {
			start := prevEnd + strings.Index(highlightedCode, "\u0001S") + 1

			logData["start"] = start

			end := strings.Index(highlightedCode, "\u0001E")
			if end == -1 {
				break
			}
			logData["end"] = prevEnd + end - 2

			snippet := models.Snippet{
				Start: start,
				End:   prevEnd + end - 2,
			}

			prevEnd = snippet.End

			result.Source.Matches.Code = append(result.Source.Matches.Code, snippet)
			log.Info("added code snippet", logData)

			highlightedCode = string(highlightedCode[end+2:])
		}
	}

	if len(result.Highlight.Label) > 0 {
		highlightedLabel := result.Highlight.Label[0]
		var prevEnd int
		logData := log.Data{}
		for {
			start := prevEnd + strings.Index(highlightedLabel, "\u0001S") + 1

			logData["start"] = start

			end := strings.Index(highlightedLabel, "\u0001E")
			if end == -1 {
				break
			}
			logData["end"] = prevEnd + end - 2

			snippet := models.Snippet{
				Start: start,
				End:   prevEnd + end - 2,
			}

			prevEnd = snippet.End

			result.Source.Matches.Label = append(result.Source.Matches.Label, snippet)
			log.Info("added label snippet", logData)

			highlightedLabel = string(highlightedLabel[end+2:])
		}
	}

	return result
}

func (api *SearchAPI) deleteSearchIndex(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	instanceID := vars["instance_id"]
	dimension := vars["dimension"]

	logData := log.Data{"instance_id": instanceID, "dimension": dimension}

	status, err := api.elasticsearch.DeleteSearchIndex(r.Context(), instanceID, dimension)
	logData["status"] = status
	if err != nil {
		log.ErrorC("failed to delete index", err, logData)
		setErrorCode(w, err)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusOK)

	log.Info("index deleted", logData)
}

func setJSONContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}

func setErrorCode(w http.ResponseWriter, err error, typ ...string) {
	switch err.Error() {
	case "Not found":
		http.Error(w, notFoundError, http.StatusNotFound)
	case "Version not found":
		http.Error(w, notFoundError, http.StatusNotFound)
	case "Index not found":
		http.Error(w, notFoundError, http.StatusNotFound)
	default:
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}
}
