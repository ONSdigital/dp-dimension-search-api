package api

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	errs "github.com/ONSdigital/dp-dimension-search-api/apierrors"
	"github.com/ONSdigital/dp-dimension-search-api/models"
	"github.com/ONSdigital/dp-dimension-search-api/searchoutputqueue"
	"github.com/ONSdigital/dp-net/request"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

const (
	defaultLimit  = 50
	defaultOffset = 0

	datasetNotFound = "dataset not found"
	editionNotFound = "edition not found"
	versionNotFound = "version not found"

	internalError         = "internal server error"
	exceedsDefaultMaximum = "the maximum offset has been reached, the offset cannot be more than"
)

var (
	err        error
	reNotFound = regexp.MustCompile(`\bbody: (\w+ not found)[\n$]`)
)

func (api *SearchAPI) getSearch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
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

	log.Info(ctx, "getSearch endpoint: incoming request", logData)

	serviceAuthToken := ""
	if api.hasPrivateEndpoints && request.IsCallerPresent(ctx) {
		// Authorised to search against an unpublished version
		// and exposes private endpoints
		serviceAuthToken = api.serviceAuthToken
	}

	// Get instanceID from datasetAPI
	versionDoc, err := api.datasetAPIClient.GetVersion(ctx, "", serviceAuthToken, "", "", datasetID, edition, version)
	if err != nil {
		log.Error(ctx, "getSearch endpoint: failed to get version of a dataset from the dataset API", err, logData)
		setErrorCode(w, setError(err))
		return
	}

	logData["version_doc"] = versionDoc

	instanceID := versionDoc.ID
	logData["instance_id"] = instanceID

	limit := defaultLimit
	if requestedLimit != "" {
		limit, err = strconv.Atoi(requestedLimit)
		if err != nil {
			log.Error(ctx, "getSearch endpoint: request limit parameter error", err, logData)
			setErrorCode(w, errs.ErrParsingQueryParameters)
			return
		}
	}

	offset := defaultOffset
	if requestedOffset != "" {
		offset, err = strconv.Atoi(requestedOffset)
		if err != nil {
			log.Error(ctx, "getSearch endpoint: request offset parameter error", err, logData)
			setErrorCode(w, errs.ErrParsingQueryParameters)
			return
		}
	}

	page := &models.PageVariables{
		DefaultMaxResults: api.defaultMaxResults,
		Limit:             limit,
		Offset:            offset,
	}

	if err = page.ValidateQueryParameters(term); err != nil {
		log.Error(ctx, "getSearch endpoint: request offset parameter error", err, logData)
		setErrorCode(w, setError(err))
		return
	}

	logData["limit"] = page.Limit
	logData["offset"] = page.Offset

	log.Info(ctx, "getSearch endpoint: just before querying search index", logData)

	response, _, err := api.elasticsearch.QuerySearchIndex(ctx, instanceID, dimension, term, page.Limit, page.Offset)
	if err != nil {
		log.Error(ctx, "getSearch endpoint: failed to query elastic search index", err, logData)
		setErrorCode(w, setError(err))
		return
	}

	searchResults := &models.SearchResults{
		TotalCount: response.Hits.Total,
		Limit:      page.Limit,
		Offset:     page.Offset,
	}

	for _, result := range response.Hits.HitList {
		result.Source.DimensionOptionURL = result.Source.URL
		result.Source.URL = ""

		result = getSnippets(ctx, result)

		doc := result.Source
		searchResults.Items = append(searchResults.Items, doc)
	}

	searchResults.Count = len(searchResults.Items)

	b, err := json.Marshal(searchResults)
	if err != nil {
		log.Error(ctx, "getSearch endpoint: failed to marshal search resource into bytes", err, logData)
		setErrorCode(w, errs.ErrInternalServer)
		return
	}

	setJSONContentType(w)
	_, err = w.Write(b)
	if err != nil {
		log.Error(ctx, "error writing response", err, logData)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	log.Info(ctx, "getSearch endpoint: successfully searched index", logData)
}

func getSnippets(ctx context.Context, result models.HitList) models.HitList {

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
			log.Info(ctx, "getSearch endpoint: added code snippet", logData)

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
			log.Info(ctx, "getSearch endpoint: added label snippet", logData)

			highlightedLabel = string(highlightedLabel[end+2:])
		}
	}

	return result
}

func (api *SearchAPI) createSearchIndex(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	instanceID := vars["instance_id"]
	dimension := vars["dimension"]

	logData := log.Data{"instance_id": instanceID, "dimension": dimension}

	log.Info(ctx, "createSearchIndex endpoint: attempting to enqueue a new search index", logData)

	output := &searchoutputqueue.Search{
		Dimension:  dimension,
		InstanceID: instanceID,
	}

	if err := api.searchOutputQueue.Queue(output); err != nil {
		setErrorCode(w, err)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusOK)

	log.Info(ctx, "createSearchIndex endpoint: index creation queued", logData)
}

func (api *SearchAPI) deleteSearchIndex(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	instanceID := vars["instance_id"]
	dimension := vars["dimension"]

	logData := log.Data{"instance_id": instanceID, "dimension": dimension}

	log.Info(ctx, "deleteSearchIndex endpoint: attempting to delete search index", logData)

	status, err := api.elasticsearch.DeleteSearchIndex(ctx, instanceID, dimension)
	logData["status"] = status
	if err != nil {
		setErrorCode(w, err)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusOK)

	log.Info(ctx, "deleteSearchIndex endpoint: search index deleted", logData)
}

func setJSONContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}

func setError(err error) (searchError error) {
	switch {
	case strings.Contains(err.Error(), versionNotFound):
		searchError = errs.ErrVersionNotFound
	case strings.Contains(err.Error(), editionNotFound):
		searchError = errs.ErrEditionNotFound
	case strings.Contains(err.Error(), datasetNotFound):
		searchError = errs.ErrDatasetNotFound
	default:
		searchError = err
	}

	return searchError
}

func setErrorCode(w http.ResponseWriter, err error) {

	switch {
	case errs.NotFoundMap[err]:
		http.Error(w, err.Error(), http.StatusNotFound)
	case errs.BadRequestMap[err]:
		http.Error(w, err.Error(), http.StatusBadRequest)
	case strings.Contains(err.Error(), exceedsDefaultMaximum):
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		http.Error(w, internalError, http.StatusInternalServerError)
	}
}
