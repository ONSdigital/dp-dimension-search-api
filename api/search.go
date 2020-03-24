package api

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	errs "github.com/ONSdigital/dp-search-api/apierrors"
	"github.com/ONSdigital/dp-search-api/models"
	"github.com/ONSdigital/dp-search-api/searchoutputqueue"
	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
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

	log.InfoCtx(ctx, "getSearch endpoint: incoming request", logData)

	auditParams := common.Params{"dataset_id": datasetID, "edition": edition, "version": version, "dimension": dimension}

	if auditError := api.auditor.Record(ctx, models.AuditTaskGetSearch, models.AuditActionAttempted, auditParams); auditError != nil {
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}

	serviceAuthToken := ""
	if api.hasPrivateEndpoints && common.IsCallerPresent(ctx) {
		// Authorised to search against an unpublished version
		// and exposes private endpoints
		serviceAuthToken = api.serviceAuthToken
	}

	b, err := func() ([]byte, error) {

		// Get instanceID from datasetAPI
		versionDoc, err := api.datasetAPIClient.GetVersion(ctx, "", serviceAuthToken, "", "", datasetID, edition, version)
		if err != nil {
			log.ErrorCtx(ctx, errors.WithMessage(err, "getSearch endpoint: failed to get version of a dataset from the dataset API"), logData)
			return nil, setError(err)
		}

		logData["version_doc"] = versionDoc

		instanceID := versionDoc.ID
		logData["instance_id"] = instanceID

		limit := defaultLimit
		if requestedLimit != "" {
			limit, err = strconv.Atoi(requestedLimit)
			if err != nil {
				log.ErrorCtx(ctx, errors.WithMessage(err, "getSearch endpoint: request limit parameter error"), logData)
				return nil, errs.ErrParsingQueryParameters
			}
		}

		offset := defaultOffset
		if requestedOffset != "" {
			offset, err = strconv.Atoi(requestedOffset)
			if err != nil {
				log.ErrorCtx(ctx, errors.WithMessage(err, "getSearch endpoint: request offset parameter error"), logData)
				return nil, errs.ErrParsingQueryParameters
			}
		}

		page := &models.PageVariables{
			DefaultMaxResults: api.defaultMaxResults,
			Limit:             limit,
			Offset:            offset,
		}

		if err = page.ValidateQueryParameters(term); err != nil {
			log.ErrorCtx(ctx, errors.WithMessage(err, "getSearch endpoint: failed query parameter validation"), logData)
			return nil, err
		}

		logData["limit"] = page.Limit
		logData["offset"] = page.Offset

		log.InfoCtx(ctx, "getSearch endpoint: just before querying search index", logData)

		response, _, err := api.elasticsearch.QuerySearchIndex(ctx, instanceID, dimension, term, page.Limit, page.Offset)
		if err != nil {
			log.ErrorCtx(ctx, errors.WithMessage(err, "getSearch endpoint: failed to query elastic search index"), logData)
			return nil, err
		}

		searchResults := &models.SearchResults{
			Count:  response.Hits.Total,
			Limit:  page.Limit,
			Offset: page.Offset,
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
			log.ErrorCtx(ctx, errors.WithMessage(err, "getSearch endpoint: failed to marshal search resource into bytes"), logData)
			return nil, errs.ErrInternalServer
		}

		return b, nil
	}()
	if err != nil {
		if auditErr := api.auditor.Record(ctx, models.AuditTaskGetSearch, models.AuditActionUnsuccessful, auditParams); auditErr != nil {
			err = auditErr
		}

		setErrorCode(ctx, w, err)
		return
	}

	if auditError := api.auditor.Record(ctx, models.AuditTaskGetSearch, models.AuditActionSuccessful, auditParams); auditError != nil {
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}

	setJSONContentType(w)
	_, err = w.Write(b)
	if err != nil {
		log.ErrorCtx(ctx, err, logData)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	log.InfoCtx(ctx, "getSearch endpoint: successfully searched index", logData)
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
			log.InfoCtx(ctx, "getSearch endpoint: added code snippet", logData)

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
			log.InfoCtx(ctx, "getSearch endpoint: added label snippet", logData)

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
	auditParams := common.Params{"instance_id": instanceID, "dimension": dimension}

	log.InfoCtx(ctx, "createSearchIndex endpoint: attempting to enqueue a new search index", logData)

	output := &searchoutputqueue.Search{
		Dimension:  dimension,
		InstanceID: instanceID,
	}

	if err := api.searchOutputQueue.Queue(output); err != nil {
		if auditError := api.auditor.Record(ctx, models.AuditTaskCreateIndex, models.AuditActionUnsuccessful, auditParams); auditError != nil {
			http.Error(w, internalError, http.StatusInternalServerError)
			return
		}

		setErrorCode(ctx, w, err)
		return
	}

	api.auditor.Record(ctx, models.AuditTaskCreateIndex, models.AuditActionSuccessful, auditParams)

	setJSONContentType(w)
	w.WriteHeader(http.StatusOK)

	log.InfoCtx(ctx, "createSearchIndex endpoint: index creation queued", logData)
}

func (api *SearchAPI) deleteSearchIndex(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	instanceID := vars["instance_id"]
	dimension := vars["dimension"]

	logData := log.Data{"instance_id": instanceID, "dimension": dimension}
	auditParams := common.Params{"instance_id": instanceID, "dimension": dimension}

	log.InfoCtx(ctx, "deleteSearchIndex endpoint: attempting to delete search index", logData)

	status, err := api.elasticsearch.DeleteSearchIndex(ctx, instanceID, dimension)
	logData["status"] = status
	if err != nil {
		if auditError := api.auditor.Record(ctx, models.AuditTaskDeleteIndex, models.AuditActionUnsuccessful, auditParams); auditError != nil {
			http.Error(w, internalError, http.StatusInternalServerError)
			return
		}

		setErrorCode(ctx, w, err)
		return
	}

	api.auditor.Record(ctx, models.AuditTaskDeleteIndex, models.AuditActionSuccessful, auditParams)

	setJSONContentType(w)
	w.WriteHeader(http.StatusOK)

	log.InfoCtx(ctx, "deleteSearchIndex endpoint: search index deleted", logData)
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

func setErrorCode(ctx context.Context, w http.ResponseWriter, err error) {

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
