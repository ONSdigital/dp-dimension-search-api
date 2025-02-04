package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"testing"

	errs "github.com/ONSdigital/dp-dimension-search-api/apierrors"
	"github.com/ONSdigital/dp-dimension-search-api/mocks"
	"github.com/ONSdigital/dp-dimension-search-api/models"
	"github.com/ONSdigital/dp-net/request"
	"github.com/gorilla/mux"
	"github.com/smartystreets/goconvey/convey"
)

var (
	defaultMaxResults = 200
	host              = &url.URL{Host: "localhost", Scheme: "http"}
)

type testOpts struct {
	method                string
	url                   string
	serviceAuthToken      string
	maxResults            int
	dsInternalServerError bool
	dsRequireNoAuth       bool
	dsRequireAuth         bool
	dsVersionNotFound     bool
	esIndexNotFound       bool
	esInternalServerError bool
	reqHasAuth            bool
	searchReturnError     bool
	privateSubnet         bool
	enableURLRewriting    bool
	externalRequest       bool
}
type testRes struct {
	w              *httptest.ResponseRecorder
	datasetAPIMock *mocks.DatasetAPI
}

func setupTest(opts testOpts) testRes {
	if opts.method == "" {
		opts.method = "GET"
	}
	r := httptest.NewRequest(opts.method, opts.url, http.NoBody)
	w := httptest.NewRecorder()

	if opts.maxResults == 0 {
		opts.maxResults = defaultMaxResults
	}

	datasetAPIMock := &mocks.DatasetAPI{InternalServerError: opts.dsInternalServerError, VersionNotFound: opts.dsVersionNotFound, RequireNoAuth: opts.dsRequireNoAuth, RequireAuth: opts.dsRequireAuth}

	// fake the auth wrapper by adding user,caller to r.Context() before ServeHTTP() is called
	if opts.reqHasAuth {
		r = r.WithContext(request.SetUser(r.Context(), "coffee@test"))
		r = r.WithContext(request.SetCaller(r.Context(), "APIAmWhoAPIAm"))
		opts.serviceAuthToken = "1234"
	}

	if opts.externalRequest {
		r.Header.Add("X-Forwarded-Proto", "https")
		r.Header.Add("X-Forwarded-Host", "api.example.com")
		r.Header.Add("X-Forwarded-Path-Prefix", "/v1")
	}
	api := routes(host, mux.NewRouter(), &mocks.BuildSearch{ReturnError: opts.searchReturnError}, datasetAPIMock, opts.serviceAuthToken, &mocks.Elasticsearch{InternalServerError: opts.esInternalServerError, IndexNotFound: opts.esIndexNotFound}, opts.maxResults, opts.privateSubnet, nil, opts.enableURLRewriting)

	api.router.ServeHTTP(w, r)

	return testRes{w: w, datasetAPIMock: datasetAPIMock}
}

func TestGetSearchPublishedWithoutAuthReturnsOK(t *testing.T) {
	t.Parallel()
	convey.Convey("Given the search query satisfies the published search index then return OK", t, func() {
		testres := setupTest(testOpts{url: "http://localhost:23100/dimension-search/datasets/123/editions/2017/versions/1/dimensions/aggregate?q=term"})
		convey.So(testres.w.Code, convey.ShouldEqual, http.StatusOK)
		convey.So(testres.datasetAPIMock.Calls, convey.ShouldEqual, 1)
		convey.So(testres.datasetAPIMock.IsAuthenticated, convey.ShouldEqual, false)
	})
}

func TestGetSearchWithAuthReturnsOK(t *testing.T) {
	t.Parallel()

	convey.Convey("Given the search query satisfies the search index then return a status 200", t, func() {
		testres := setupTest(testOpts{
			url:                "http://localhost:23100/dimension-search/datasets/123/editions/2017/versions/1/dimensions/aggregate?q=term",
			reqHasAuth:         true,
			enableURLRewriting: false,
		})

		convey.So(testres.w.Code, convey.ShouldEqual, http.StatusOK)

		// Check response json
		searchResults := getSearchResults(testres.w.Body)

		convey.So(searchResults.Count, convey.ShouldEqual, 2)
		convey.So(len(searchResults.Items), convey.ShouldEqual, 2)
		convey.So(searchResults.Limit, convey.ShouldEqual, 50)
		convey.So(searchResults.Offset, convey.ShouldEqual, 0)
		convey.So(searchResults.Items[0].Code, convey.ShouldEqual, "frs34g5t98hdd")
		convey.So(searchResults.Items[0].DimensionOptionURL, convey.ShouldEqual, "http://localhost:8080/testing/1")
		convey.So(searchResults.Items[0].HasData, convey.ShouldEqual, true)
		convey.So(searchResults.Items[0].Label, convey.ShouldEqual, "something and someone")
		convey.So(searchResults.Items[0].NumberOfChildren, convey.ShouldEqual, 3)
		convey.So(len(searchResults.Items[0].Matches.Code), convey.ShouldEqual, 1)
		convey.So(searchResults.Items[0].Matches.Code[0].Start, convey.ShouldEqual, 1)
		convey.So(searchResults.Items[0].Matches.Code[0].End, convey.ShouldEqual, 13)
		convey.So(len(searchResults.Items[0].Matches.Label), convey.ShouldEqual, 2)
		convey.So(searchResults.Items[0].Matches.Label[0].Start, convey.ShouldEqual, 1)
		convey.So(searchResults.Items[0].Matches.Label[0].End, convey.ShouldEqual, 9)
		convey.So(searchResults.Items[0].Matches.Label[1].Start, convey.ShouldEqual, 13)
		convey.So(searchResults.Items[0].Matches.Label[1].End, convey.ShouldEqual, 19)
		convey.So(searchResults.Items[1].Code, convey.ShouldEqual, "gt534g5t98hs1")
		convey.So(searchResults.Items[1].DimensionOptionURL, convey.ShouldEqual, "http://localhost:8080/testing/2")
		convey.So(searchResults.Items[1].HasData, convey.ShouldEqual, false)
		convey.So(searchResults.Items[1].Label, convey.ShouldEqual, "something else and someone else")
		convey.So(searchResults.Items[1].NumberOfChildren, convey.ShouldEqual, 10)
		convey.So(len(searchResults.Items[1].Matches.Code), convey.ShouldEqual, 0)
		convey.So(len(searchResults.Items[1].Matches.Label), convey.ShouldEqual, 2)
		convey.So(searchResults.Items[1].Matches.Label[0].Start, convey.ShouldEqual, 1)
		convey.So(searchResults.Items[1].Matches.Label[0].End, convey.ShouldEqual, 9)
		convey.So(searchResults.Items[1].Matches.Label[1].Start, convey.ShouldEqual, 19)
		convey.So(searchResults.Items[1].Matches.Label[1].End, convey.ShouldEqual, 25)
		convey.So(searchResults.Items[1].Matches, convey.ShouldResemble, models.Matches{Code: []models.Snippet(nil), Label: []models.Snippet{{Start: 1, End: 9}, {Start: 19, End: 25}}})
	})

	convey.Convey("Given the search query satisfies the search index then return a status 200 with URL rewriting enabled", t, func() {
		testres := setupTest(testOpts{
			url:                "http://localhost:23100/dimension-search/datasets/123/editions/2017/versions/1/dimensions/aggregate?q=term",
			reqHasAuth:         true,
			enableURLRewriting: true,
		})

		convey.So(testres.w.Code, convey.ShouldEqual, http.StatusOK)

		// Check response json
		searchResults := getSearchResults(testres.w.Body)

		convey.So(searchResults.Count, convey.ShouldEqual, 2)
		convey.So(len(searchResults.Items), convey.ShouldEqual, 2)
		convey.So(searchResults.Limit, convey.ShouldEqual, 50)
		convey.So(searchResults.Offset, convey.ShouldEqual, 0)
		convey.So(searchResults.Items[0].Code, convey.ShouldEqual, "frs34g5t98hdd")
		convey.So(searchResults.Items[0].DimensionOptionURL, convey.ShouldEqual, "http://localhost/testing/1")
		convey.So(searchResults.Items[0].HasData, convey.ShouldEqual, true)
		convey.So(searchResults.Items[0].Label, convey.ShouldEqual, "something and someone")
		convey.So(searchResults.Items[0].NumberOfChildren, convey.ShouldEqual, 3)
		convey.So(len(searchResults.Items[0].Matches.Code), convey.ShouldEqual, 1)
		convey.So(searchResults.Items[0].Matches.Code[0].Start, convey.ShouldEqual, 1)
		convey.So(searchResults.Items[0].Matches.Code[0].End, convey.ShouldEqual, 13)
		convey.So(len(searchResults.Items[0].Matches.Label), convey.ShouldEqual, 2)
		convey.So(searchResults.Items[0].Matches.Label[0].Start, convey.ShouldEqual, 1)
		convey.So(searchResults.Items[0].Matches.Label[0].End, convey.ShouldEqual, 9)
		convey.So(searchResults.Items[0].Matches.Label[1].Start, convey.ShouldEqual, 13)
		convey.So(searchResults.Items[0].Matches.Label[1].End, convey.ShouldEqual, 19)
		convey.So(searchResults.Items[1].Code, convey.ShouldEqual, "gt534g5t98hs1")
		convey.So(searchResults.Items[1].DimensionOptionURL, convey.ShouldEqual, "http://localhost/testing/2")
		convey.So(searchResults.Items[1].HasData, convey.ShouldEqual, false)
		convey.So(searchResults.Items[1].Label, convey.ShouldEqual, "something else and someone else")
		convey.So(searchResults.Items[1].NumberOfChildren, convey.ShouldEqual, 10)
		convey.So(len(searchResults.Items[1].Matches.Code), convey.ShouldEqual, 0)
		convey.So(len(searchResults.Items[1].Matches.Label), convey.ShouldEqual, 2)
		convey.So(searchResults.Items[1].Matches.Label[0].Start, convey.ShouldEqual, 1)
		convey.So(searchResults.Items[1].Matches.Label[0].End, convey.ShouldEqual, 9)
		convey.So(searchResults.Items[1].Matches.Label[1].Start, convey.ShouldEqual, 19)
		convey.So(searchResults.Items[1].Matches.Label[1].End, convey.ShouldEqual, 25)
		convey.So(searchResults.Items[1].Matches, convey.ShouldResemble, models.Matches{Code: []models.Snippet(nil), Label: []models.Snippet{{Start: 1, End: 9}, {Start: 19, End: 25}}})
	})

	convey.Convey("Given the search query satisfies the search index then return a status 200 with URL rewriting enabled and X-forwarded added as an external request", t, func() {
		testres := setupTest(testOpts{
			url:                "http://localhost:23100/dimension-search/datasets/123/editions/2017/versions/1/dimensions/aggregate?q=term",
			reqHasAuth:         true,
			enableURLRewriting: true,
			externalRequest:    true,
		})

		convey.So(testres.w.Code, convey.ShouldEqual, http.StatusOK)

		// Check response json
		searchResults := getSearchResults(testres.w.Body)

		convey.So(searchResults.Count, convey.ShouldEqual, 2)
		convey.So(len(searchResults.Items), convey.ShouldEqual, 2)

		convey.So(searchResults.Items[0].DimensionOptionURL, convey.ShouldEqual, "https://api.example.com/v1/testing/1")
		convey.So(searchResults.Items[1].DimensionOptionURL, convey.ShouldEqual, "https://api.example.com/v1/testing/2")
		convey.So(searchResults.Items[1].Matches, convey.ShouldResemble, models.Matches{Code: []models.Snippet(nil), Label: []models.Snippet{{Start: 1, End: 9}, {Start: 19, End: 25}}})
	})

	convey.Convey("Given the search query satisfies the search index when limit and offset parameters are set then return a status 200", t, func() {
		testres := setupTest(testOpts{
			url:                "http://localhost:23100/dimension-search/datasets/123/editions/2017/versions/1/dimensions/aggregate?q=term",
			reqHasAuth:         true,
			enableURLRewriting: true,
			externalRequest:    true,
		})
		convey.So(testres.w.Code, convey.ShouldEqual, http.StatusOK)

		// Check response json
		searchResults := getSearchResults(testres.w.Body)

		convey.So(searchResults.Count, convey.ShouldEqual, 2)
		convey.So(len(searchResults.Items), convey.ShouldEqual, 2)
		convey.So(searchResults.Limit, convey.ShouldEqual, 5)
		convey.So(searchResults.Offset, convey.ShouldEqual, 20)
		convey.So(searchResults.TotalCount, convey.ShouldEqual, 22)
	})

	convey.Convey("Given the search query satisfies the search index when limit parameter is set beyond the maximum then return a status 200", t, func() {
		testres := setupTest(testOpts{
			url:        "http://localhost:23100/dimension-search/datasets/123/editions/2017/versions/1/dimensions/aggregate?q=term&limit=20000",
			maxResults: defaultMaxResults,
			reqHasAuth: true,
		})
		convey.So(testres.w.Code, convey.ShouldEqual, http.StatusOK)

		// Check response json
		searchResults := getSearchResults(testres.w.Body)

		convey.So(searchResults.Count, convey.ShouldEqual, 2)
		convey.So(len(searchResults.Items), convey.ShouldEqual, 2)
		convey.So(searchResults.Limit, convey.ShouldEqual, defaultMaxResults)
		convey.So(searchResults.Offset, convey.ShouldEqual, 0)
		convey.So(searchResults.TotalCount, convey.ShouldEqual, 22)
	})
}

func TestGetSearchFailureScenarios(t *testing.T) {
	t.Parallel()
	convey.Convey("Given search API fails to connect to the dataset API return status 500 (internal service error)", t, func() {
		testres := setupTest(testOpts{
			url:                   "http://localhost:23100/dimension-search/datasets/123/editions/2017/versions/1/dimensions/aggregate?q=term",
			dsInternalServerError: true,
		})
		convey.So(testres.w.Code, convey.ShouldEqual, http.StatusInternalServerError)
		convey.So(testres.w.Body.String(), convey.ShouldContainSubstring, "internal server error")
	})

	convey.Convey("Given the version document was not found via the dataset API return status 404 (not found)", t, func() {
		testres := setupTest(testOpts{
			url:               "http://localhost:23100/dimension-search/datasets/123/editions/2017/versions/1/dimensions/aggregate?q=term",
			dsVersionNotFound: true,
		})
		convey.So(testres.w.Code, convey.ShouldEqual, http.StatusNotFound)
		convey.So(testres.w.Body.String(), convey.ShouldContainSubstring, errs.ErrVersionNotFound.Error())
		convey.So(testres.datasetAPIMock.Calls, convey.ShouldEqual, 1)
		convey.So(testres.datasetAPIMock.IsAuthenticated, convey.ShouldEqual, false)
	})

	convey.Convey("Given the limit parameter in request is not a number return status 400 (bad request)", t, func() {
		testres := setupTest(testOpts{
			url: "http://localhost:23100/dimension-search/datasets/123/editions/2017/versions/1/dimensions/aggregate?q=term&limit=four",
		})
		convey.So(testres.w.Code, convey.ShouldEqual, http.StatusBadRequest)
		convey.So(testres.w.Body.String(), convey.ShouldContainSubstring, errs.ErrParsingQueryParameters.Error())
	})

	convey.Convey("Given the offset parameter in request is not a number return status 400 (bad request)", t, func() {
		testres := setupTest(testOpts{
			url: "http://localhost:23100/dimension-search/datasets/123/editions/2017/versions/1/dimensions/aggregate?q=term&offset=fifty",
		})
		convey.So(testres.w.Code, convey.ShouldEqual, http.StatusBadRequest)
		convey.So(testres.w.Body.String(), convey.ShouldContainSubstring, errs.ErrParsingQueryParameters.Error())
	})

	convey.Convey("Given the query parameter, q does not exist in request return status 400 (bad request)", t, func() {
		testres := setupTest(testOpts{
			url: "http://localhost:23100/dimension-search/datasets/123/editions/2017/versions/1/dimensions/aggregate",
		})
		convey.So(testres.w.Code, convey.ShouldEqual, http.StatusBadRequest)
		convey.So(testres.w.Body.String(), convey.ShouldContainSubstring, errs.ErrEmptySearchTerm.Error())
	})

	convey.Convey("Given the offset parameter exceeds the default maximum results return status 400 (bad request)", t, func() {
		testres := setupTest(testOpts{
			url: "http://localhost:23100/dimension-search/datasets/123/editions/2017/versions/1/dimensions/aggregate?q=term&offset=500",
		})
		convey.So(testres.w.Code, convey.ShouldEqual, http.StatusBadRequest)
		convey.So(testres.w.Body.String(), convey.ShouldEqual, "the maximum offset has been reached, the offset cannot be more than "+strconv.Itoa(defaultMaxResults)+"\n")
	})

	convey.Convey("Given search API fails to connect to elastic search cluster return status 500 (internal service error)", t, func() {
		testres := setupTest(testOpts{
			url:                   "http://localhost:23100/dimension-search/datasets/123/editions/2017/versions/1/dimensions/aggregate?q=term",
			esInternalServerError: true,
		})
		convey.So(testres.w.Code, convey.ShouldEqual, http.StatusInternalServerError)
		convey.So(testres.w.Body.String(), convey.ShouldEqual, "internal server error\n")
	})

	convey.Convey("Given the search index does not exist but the version resource does then return status 500 (internal server error)", t, func() {
		testres := setupTest(testOpts{
			url:             "http://localhost:23100/dimension-search/datasets/123/editions/2017/versions/1/dimensions/aggregate?q=term",
			esIndexNotFound: true,
		})
		convey.So(testres.w.Code, convey.ShouldEqual, http.StatusInternalServerError)
		convey.So(testres.w.Body.String(), convey.ShouldContainSubstring, errs.ErrInternalServer.Error())
	})
}

// ensure no authentication is sent to the dataset API from public
func TestPublicSubnetUsersCannotSeeUnpublished(t *testing.T) {
	convey.Convey("Given public subnet, when an authenticated GET is made, then the dataset api should not see authentication and returns not found, so we return status 404 (not found)", t, func() {
		testres := setupTest(testOpts{
			url:               "http://localhost:23100/dimension-search/datasets/123/editions/2017/versions/1/dimensions/aggregate?q=term",
			dsRequireNoAuth:   true,
			dsVersionNotFound: true,
		})
		convey.So(testres.w.Code, convey.ShouldEqual, http.StatusNotFound)
		convey.So(testres.w.Body.String(), convey.ShouldContainSubstring, errs.ErrVersionNotFound.Error())
		convey.So(testres.datasetAPIMock.Calls, convey.ShouldEqual, 1)
		convey.So(testres.datasetAPIMock.IsAuthenticated, convey.ShouldEqual, false)
	})

	convey.Convey("Given public subnet, when an unauthenticated GET is made, then the dataset api should not see authentication and returns not found, so we return status 404 (not found)", t, func() {
		testres := setupTest(testOpts{
			url:               "http://localhost:23100/dimension-search/datasets/123/editions/2017/versions/1/dimensions/aggregate?q=term",
			dsRequireNoAuth:   true,
			dsVersionNotFound: true,
		})
		convey.So(testres.w.Code, convey.ShouldEqual, http.StatusNotFound)
		convey.So(testres.w.Body.String(), convey.ShouldContainSubstring, errs.ErrVersionNotFound.Error())
		convey.So(testres.datasetAPIMock.Calls, convey.ShouldEqual, 1)
		convey.So(testres.datasetAPIMock.IsAuthenticated, convey.ShouldEqual, false)
	})
}

// ensure authentication is sent to the dataset API appropriately (only when client is authenticated)
func TestPrivateSubnetMightSeeUnpublished(t *testing.T) {
	convey.Convey("Given private subnet, when an authenticated GET is made, then the dataset api should see authentication and return ok, so we return OK", t, func() {
		testres := setupTest(testOpts{
			url:           "http://localhost:23100/dimension-search/datasets/123/editions/2017/versions/1/dimensions/aggregate?q=term",
			dsRequireAuth: true,
			reqHasAuth:    true,
			privateSubnet: true,
		})
		convey.So(testres.w.Body.String(), convey.ShouldStartWith, "{")
		convey.So(testres.w.Code, convey.ShouldEqual, http.StatusOK)
		convey.So(testres.datasetAPIMock.Calls, convey.ShouldEqual, 1)
		convey.So(testres.datasetAPIMock.IsAuthenticated, convey.ShouldEqual, true)
	})

	convey.Convey("Given private subnet, when an authenticated GET is made, force the dataset api to return 404 if authenticated, so we return 404", t, func() {
		testres := setupTest(testOpts{
			url:               "http://localhost:23100/dimension-search/datasets/123/editions/2017/versions/1/dimensions/aggregate?q=term",
			reqHasAuth:        true,
			dsRequireAuth:     true,
			dsVersionNotFound: true,
			privateSubnet:     true,
		})
		convey.So(testres.w.Code, convey.ShouldEqual, http.StatusNotFound)
		convey.So(testres.w.Body.String(), convey.ShouldContainSubstring, errs.ErrVersionNotFound.Error())
	})

	convey.Convey("Given private subnet, when an authenticated GET is made, force the dataset api to return 500 if authenticated, so we return 500", t, func() {
		testres := setupTest(testOpts{
			url:                   "http://localhost:23100/dimension-search/datasets/123/editions/2017/versions/1/dimensions/aggregate?q=term",
			dsInternalServerError: true,
			dsRequireAuth:         true,
			reqHasAuth:            true,
			privateSubnet:         true,
		})
		convey.So(testres.w.Code, convey.ShouldEqual, http.StatusInternalServerError)
	})

	convey.Convey("Given private subnet, when an unauthenticated GET is made, then the dataset api should see no authentication and return not found, so we return status 404 (not found)", t, func() {
		testres := setupTest(testOpts{
			url:               "http://localhost:23100/dimension-search/datasets/123/editions/2017/versions/1/dimensions/aggregate?q=term",
			dsVersionNotFound: true,
			dsRequireNoAuth:   true,
			privateSubnet:     true,
		})
		convey.So(testres.w.Code, convey.ShouldEqual, http.StatusNotFound)
		convey.So(testres.datasetAPIMock.Calls, convey.ShouldEqual, 1)
		convey.So(testres.datasetAPIMock.IsAuthenticated, convey.ShouldEqual, false)
		convey.So(testres.w.Body.String(), convey.ShouldContainSubstring, errs.ErrVersionNotFound.Error())
	})

	convey.Convey("Given private subnet, when a badly-authenticated GET is made, then the dataset api should see no authentication and returns not found, so we return server error", t, func() {
		testres := setupTest(testOpts{
			url:               "http://localhost:23100/dimension-search/datasets/123/editions/2017/versions/1/dimensions/aggregate?q=term",
			dsRequireNoAuth:   true,
			dsVersionNotFound: true,
			privateSubnet:     true,
		})
		convey.So(testres.w.Code, convey.ShouldEqual, http.StatusNotFound)
		convey.So(testres.w.Body.String(), convey.ShouldContainSubstring, errs.ErrVersionNotFound.Error())

		convey.So(testres.datasetAPIMock.Calls, convey.ShouldEqual, 1)
		convey.So(testres.datasetAPIMock.IsAuthenticated, convey.ShouldEqual, false)
	})

	convey.Convey("Given private subnet, when an unauthenticated GET is made, then the dataset api should see no authentication and return not found, so we return status 404 (not found)", t, func() {
		testres := setupTest(testOpts{
			url:               "http://localhost:23100/dimension-search/datasets/123/editions/2017/versions/1/dimensions/aggregate?q=term",
			dsRequireNoAuth:   true,
			dsVersionNotFound: true,
			privateSubnet:     true,
		})
		convey.So(testres.w.Code, convey.ShouldEqual, http.StatusNotFound)
		convey.So(testres.w.Body.String(), convey.ShouldContainSubstring, errs.ErrVersionNotFound.Error())
	})
}

func TestCreateSearchIndexReturnsOK(t *testing.T) {
	convey.Convey("Given instance and dimension exist return a status 200 (ok)", t, func() {
		testres := setupTest(testOpts{
			method:        "PUT",
			url:           "http://localhost:23100/dimension-search/instances/123/dimensions/aggregate",
			dsRequireAuth: true,
			reqHasAuth:    true,
			privateSubnet: true,
		})
		convey.So(testres.w.Code, convey.ShouldEqual, http.StatusOK)
	})
}

func TestFailToCreateSearchIndex(t *testing.T) {
	convey.Convey("Given a request to create search index but no auth header is set return a status 401 (unauthorized)", t, func() {
		testres := setupTest(testOpts{
			method:          "PUT",
			url:             "http://localhost:23100/dimension-search/instances/123/dimensions/aggregate",
			dsRequireNoAuth: true,
			privateSubnet:   true,
		})
		convey.So(testres.w.Code, convey.ShouldEqual, http.StatusUnauthorized)
	})

	convey.Convey("Given a request to create search index but the auth header is wrong return a status 401 (unauthorized)", t, func() {
		testres := setupTest(testOpts{
			method:          "PUT",
			url:             "http://localhost:23100/dimension-search/instances/123/dimensions/aggregate",
			dsRequireNoAuth: true,
			privateSubnet:   true,
		})
		convey.So(testres.w.Code, convey.ShouldEqual, http.StatusUnauthorized)
	})

	convey.Convey("Given a request to create search index but unable to connect to kafka broker return a status 500 (internal service error)", t, func() {
		testres := setupTest(testOpts{
			method:            "PUT",
			url:               "http://localhost:23100/dimension-search/instances/123/dimensions/aggregate",
			dsRequireAuth:     true,
			searchReturnError: true,
			reqHasAuth:        true,
			privateSubnet:     true,
		})
		convey.So(testres.w.Code, convey.ShouldEqual, http.StatusInternalServerError)
		convey.So(testres.w.Body.String(), convey.ShouldContainSubstring, errs.ErrInternalServer.Error())
	})
}

func TestDeleteSearchIndexReturnsOK(t *testing.T) {
	convey.Convey("Given a search index exists return a status 200 (ok)", t, func() {
		testres := setupTest(testOpts{
			method:        "DELETE",
			url:           "http://localhost:23100/dimension-search/instances/123/dimensions/aggregate",
			dsRequireAuth: true,
			reqHasAuth:    true,
			privateSubnet: true,
		})
		convey.So(testres.w.Code, convey.ShouldEqual, http.StatusOK)
	})
}

func TestFailToDeleteSearchIndex(t *testing.T) {
	convey.Convey("Given a search index exists but no auth header set return a status 401 (unauthorized)", t, func() {
		testres := setupTest(testOpts{
			method:          "DELETE",
			url:             "http://localhost:23100/dimension-search/instances/123/dimensions/aggregate",
			dsRequireNoAuth: true,
			privateSubnet:   true,
		})
		convey.So(testres.w.Code, convey.ShouldEqual, http.StatusUnauthorized)
	})

	convey.Convey("Given a search index exists but auth header is wrong return a status 401 (unauthorized)", t, func() {
		testres := setupTest(testOpts{
			method:        "DELETE",
			url:           "http://localhost:23100/dimension-search/instances/123/dimensions/aggregate",
			dsRequireAuth: true,
			privateSubnet: true,
		})
		convey.So(testres.w.Code, convey.ShouldEqual, http.StatusUnauthorized)
	})

	convey.Convey("Given a search index exists but unable to connect to elasticsearch cluster return a status 500 (internal service error)", t, func() {
		testres := setupTest(testOpts{
			method:                "DELETE",
			url:                   "http://localhost:23100/dimension-search/instances/123/dimensions/aggregate",
			dsRequireAuth:         true,
			esInternalServerError: true,
			privateSubnet:         true,
			reqHasAuth:            true,
		})
		convey.So(testres.w.Code, convey.ShouldEqual, http.StatusInternalServerError)
		convey.So(testres.w.Body.String(), convey.ShouldContainSubstring, errs.ErrInternalServer.Error())
	})

	convey.Convey("Given a search index does not exists return a status 404 (not found)", t, func() {
		testres := setupTest(testOpts{
			method:          "DELETE",
			url:             "http://localhost:23100/dimension-search/instances/123/dimensions/aggregate",
			dsRequireAuth:   true,
			esIndexNotFound: true,
			reqHasAuth:      true,
			privateSubnet:   true,
		})
		convey.So(testres.w.Code, convey.ShouldEqual, http.StatusNotFound)
		convey.So(testres.w.Body.String(), convey.ShouldContainSubstring, errs.ErrDeleteIndexNotFound.Error())
	})
}

func TestCheckhighlights(t *testing.T) {
	convey.Convey("Given the elasticsearch results contain highlights then the correct snippet pairs are returned", t, func() {
		result := models.HitList{
			Highlight: models.Highlight{
				Code:  []string{"\u0001Sstrangeness\u0001E"},
				Label: []string{"04 \u0001SHousing\u0001E, water, \u0001Selectricity\u0001E, gas and other fuels"},
			},
		}
		result = getSnippets(context.Background(), result)
		convey.So(len(result.Source.Matches.Code), convey.ShouldEqual, 1)
		convey.So(result.Source.Matches.Code[0].Start, convey.ShouldEqual, 1)
		convey.So(result.Source.Matches.Code[0].End, convey.ShouldEqual, 11)
		convey.So(len(result.Source.Matches.Label), convey.ShouldEqual, 2)
		convey.So(result.Source.Matches.Label[0].Start, convey.ShouldEqual, 4)
		convey.So(result.Source.Matches.Label[0].End, convey.ShouldEqual, 10)
		convey.So(result.Source.Matches.Label[1].Start, convey.ShouldEqual, 20)
		convey.So(result.Source.Matches.Label[1].End, convey.ShouldEqual, 30)
	})
}

func TestDeleteEndpointInWebReturnsNotFound(t *testing.T) {
	convey.Convey("Given a search index exists and credentials are correct, return a status 404 (not found)", t, func() {
		testres := setupTest(testOpts{
			method:        "DELETE",
			url:           "http://localhost:23100/dimension-search/instances/123/dimensions/aggregate",
			dsRequireAuth: true,
			reqHasAuth:    true,
		})
		convey.So(testres.w.Code, convey.ShouldEqual, http.StatusNotFound)
		convey.So(testres.w.Body.String(), convey.ShouldEqual, "404 page not found\n")
	})

	convey.Convey("Given a search index exists and credentials are incorrect, return a status 404 (not found)", t, func() {
		testres := setupTest(testOpts{
			method:          "DELETE",
			url:             "http://localhost:23100/dimension-search/instances/123/dimensions/aggregate",
			dsRequireNoAuth: true,
		})
		convey.So(testres.w.Code, convey.ShouldEqual, http.StatusNotFound)
		convey.So(testres.w.Body.String(), convey.ShouldContainSubstring, "404 page not found")
	})
}

func TestCreateSearchIndexEndpointInWebReturnsNotFound(t *testing.T) {
	convey.Convey("Given instance and dimension exist and has valid auth return a status 404 (not found)", t, func() {
		testres := setupTest(testOpts{
			method:          "PUT",
			url:             "http://localhost:23100/dimension-search/instances/123/dimensions/aggregate",
			dsRequireNoAuth: true,
			reqHasAuth:      true,
		})
		convey.So(testres.w.Code, convey.ShouldEqual, http.StatusNotFound)
		convey.So(testres.w.Body.String(), convey.ShouldContainSubstring, "404 page not found")
	})

	convey.Convey("Given a request to create search index and no private endpoints when a bad auth header is used, return a status 404 (not found)", t, func() {
		testres := setupTest(testOpts{
			method:          "PUT",
			url:             "http://localhost:23100/dimension-search/instances/123/dimensions/aggregate",
			dsRequireNoAuth: true,
		})
		convey.So(testres.w.Code, convey.ShouldEqual, http.StatusNotFound)
		convey.So(testres.w.Body.String(), convey.ShouldContainSubstring, "404 page not found")
	})
}

func getSearchResults(body *bytes.Buffer) *models.SearchResults {
	jsonBody, err := io.ReadAll(body)
	if err != nil {
		os.Exit(1)
	}

	searchResults := &models.SearchResults{}
	if err := json.Unmarshal(jsonBody, searchResults); err != nil {
		os.Exit(1)
	}

	return searchResults
}
