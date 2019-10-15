package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	errs "github.com/ONSdigital/dp-search-api/apierrors"
	"github.com/ONSdigital/dp-search-api/mocks"
	"github.com/ONSdigital/dp-search-api/models"
	"github.com/ONSdigital/go-ns/audit"
	"github.com/ONSdigital/go-ns/common"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	defaultMaxResults = 200
)

type testOpts struct {
	method                string
	url                   string
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
}
type testRes struct {
	w          *httptest.ResponseRecorder
	dsWithAuth *mocks.DatasetAPI
	dsNoAuth   *mocks.DatasetAPI
	audit      *audit.AuditorServiceMock
}

func setupTest(opts testOpts) testRes {
	if opts.method == "" {
		opts.method = "GET"
	}
	r := httptest.NewRequest(opts.method, opts.url, nil)
	w := httptest.NewRecorder()

	if opts.maxResults == 0 {
		opts.maxResults = defaultMaxResults
	}

	datasetWithAuth := &mocks.DatasetAPI{InternalServerError: opts.dsInternalServerError, VersionNotFound: opts.dsVersionNotFound, RequireNoAuth: opts.dsRequireNoAuth, RequireAuth: opts.dsRequireAuth, SvcAuth: "AuthMe!"}
	datasetNoAuth := &mocks.DatasetAPI{InternalServerError: opts.dsInternalServerError, VersionNotFound: opts.dsVersionNotFound, RequireNoAuth: opts.dsRequireNoAuth, RequireAuth: opts.dsRequireAuth}

	mockAuditor := getMockAuditor()

	api := routes(
		"host", mux.NewRouter(),
		&mocks.BuildSearch{ReturnError: opts.searchReturnError},
		datasetWithAuth, datasetNoAuth,
		&mocks.Elasticsearch{InternalServerError: opts.esInternalServerError, IndexNotFound: opts.esIndexNotFound},
		opts.maxResults,
		opts.privateSubnet,
		mockAuditor,
	)

	// fake the auth wrapper by adding user,caller to r.Context() before ServeHTTP() is called
	if opts.reqHasAuth {
		r = r.WithContext(common.SetUser(r.Context(), "coffee@test"))
		r = r.WithContext(common.SetCaller(r.Context(), "APIAmWhoAPIAm"))
	}

	api.router.ServeHTTP(w, r)

	return testRes{w: w, dsWithAuth: datasetWithAuth, dsNoAuth: datasetNoAuth, audit: mockAuditor}
}

func TestGetSearchPublishedWithoutAuthReturnsOK(t *testing.T) {
	t.Parallel()
	Convey("Given the search query satisfies the published search index then return OK", t, func() {
		testres := setupTest(testOpts{url: "http://localhost:23100/search/datasets/123/editions/2017/versions/1/dimensions/aggregate?q=term"})
		So(testres.w.Code, ShouldEqual, http.StatusOK)
		So(testres.dsWithAuth.Calls, ShouldEqual, 0)
		So(testres.dsNoAuth.Calls, ShouldEqual, 1)
		expectedAuditOutcome(models.AuditTaskGetSearch, models.Scenario_attemptAndSucceed, testres)
	})
}

func TestGetSearchWithAuthReturnsOK(t *testing.T) {
	t.Parallel()

	Convey("Given the search query satisfies the search index then return a status 200", t, func() {
		testres := setupTest(testOpts{
			url:        "http://localhost:23100/search/datasets/123/editions/2017/versions/1/dimensions/aggregate?q=term",
			reqHasAuth: true,
		})
		So(testres.w.Code, ShouldEqual, http.StatusOK)

		// Check response json
		searchResults := getSearchResults(testres.w.Body)

		So(searchResults.Count, ShouldEqual, 2)
		So(len(searchResults.Items), ShouldEqual, 2)
		So(searchResults.Limit, ShouldEqual, 50)
		So(searchResults.Offset, ShouldEqual, 0)
		So(searchResults.Items[0].Code, ShouldEqual, "frs34g5t98hdd")
		So(searchResults.Items[0].DimensionOptionURL, ShouldEqual, "http://localhost:8080/testing/1")
		So(searchResults.Items[0].HasData, ShouldEqual, true)
		So(searchResults.Items[0].Label, ShouldEqual, "something and someone")
		So(searchResults.Items[0].NumberOfChildren, ShouldEqual, 3)
		So(len(searchResults.Items[0].Matches.Code), ShouldEqual, 1)
		So(searchResults.Items[0].Matches.Code[0].Start, ShouldEqual, 1)
		So(searchResults.Items[0].Matches.Code[0].End, ShouldEqual, 13)
		So(len(searchResults.Items[0].Matches.Label), ShouldEqual, 2)
		So(searchResults.Items[0].Matches.Label[0].Start, ShouldEqual, 1)
		So(searchResults.Items[0].Matches.Label[0].End, ShouldEqual, 9)
		So(searchResults.Items[0].Matches.Label[1].Start, ShouldEqual, 13)
		So(searchResults.Items[0].Matches.Label[1].End, ShouldEqual, 19)
		So(searchResults.Items[1].Code, ShouldEqual, "gt534g5t98hs1")
		So(searchResults.Items[1].DimensionOptionURL, ShouldEqual, "http://localhost:8080/testing/2")
		So(searchResults.Items[1].HasData, ShouldEqual, false)
		So(searchResults.Items[1].Label, ShouldEqual, "something else and someone else")
		So(searchResults.Items[1].NumberOfChildren, ShouldEqual, 10)
		So(len(searchResults.Items[1].Matches.Code), ShouldEqual, 0)
		So(len(searchResults.Items[1].Matches.Label), ShouldEqual, 2)
		So(searchResults.Items[1].Matches.Label[0].Start, ShouldEqual, 1)
		So(searchResults.Items[1].Matches.Label[0].End, ShouldEqual, 9)
		So(searchResults.Items[1].Matches.Label[1].Start, ShouldEqual, 19)
		So(searchResults.Items[1].Matches.Label[1].End, ShouldEqual, 25)
		So(searchResults.Items[1].Matches, ShouldResemble, models.Matches{Code: []models.Snippet(nil), Label: []models.Snippet{models.Snippet{Start: 1, End: 9}, models.Snippet{Start: 19, End: 25}}})
		expectedAuditOutcome(models.AuditTaskGetSearch, models.Scenario_attemptAndSucceed, testres)

	})

	Convey("Given the search query satisfies the search index when limit and offset parameters are set then return a status 200", t, func() {
		testres := setupTest(testOpts{
			url:        "http://localhost:23100/search/datasets/123/editions/2017/versions/1/dimensions/aggregate?q=term&limit=5&offset=20",
			maxResults: 40,
			reqHasAuth: true,
		})
		So(testres.w.Code, ShouldEqual, http.StatusOK)

		// Check response json
		searchResults := getSearchResults(testres.w.Body)

		So(searchResults.Count, ShouldEqual, 2)
		So(len(searchResults.Items), ShouldEqual, 2)
		So(searchResults.Limit, ShouldEqual, 5)
		So(searchResults.Offset, ShouldEqual, 20)
		expectedAuditOutcome(models.AuditTaskGetSearch, models.Scenario_attemptAndSucceed, testres)
	})

	Convey("Given the search query satisfies the search index when limit parameter is set beyond the maximum then return a status 200", t, func() {
		testres := setupTest(testOpts{
			url:        "http://localhost:23100/search/datasets/123/editions/2017/versions/1/dimensions/aggregate?q=term&limit=20000",
			maxResults: defaultMaxResults,
			reqHasAuth: true,
		})
		So(testres.w.Code, ShouldEqual, http.StatusOK)

		// Check response json
		searchResults := getSearchResults(testres.w.Body)

		So(searchResults.Count, ShouldEqual, 2)
		So(len(searchResults.Items), ShouldEqual, 2)
		So(searchResults.Limit, ShouldEqual, defaultMaxResults)
		So(searchResults.Offset, ShouldEqual, 0)
		expectedAuditOutcome(models.AuditTaskGetSearch, models.Scenario_attemptAndSucceed, testres)
	})
}

func TestGetSearchFailureScenarios(t *testing.T) {
	t.Parallel()
	Convey("Given search API fails to connect to the dataset API return status 500 (internal service error)", t, func() {
		testres := setupTest(testOpts{
			url:                   "http://localhost:23100/search/datasets/123/editions/2017/versions/1/dimensions/aggregate?q=term",
			dsInternalServerError: true,
		})
		So(testres.w.Code, ShouldEqual, http.StatusInternalServerError)
		So(testres.w.Body.String(), ShouldContainSubstring, "internal server error")
		expectedAuditOutcome(models.AuditTaskGetSearch, models.Scenario_attemptAndFail, testres)
	})

	Convey("Given the version document was not found via the dataset API return status 404 (not found)", t, func() {
		testres := setupTest(testOpts{
			url:               "http://localhost:23100/search/datasets/123/editions/2017/versions/1/dimensions/aggregate?q=term",
			dsVersionNotFound: true,
		})
		So(testres.w.Code, ShouldEqual, http.StatusNotFound)
		So(testres.w.Body.String(), ShouldContainSubstring, errs.ErrVersionNotFound.Error())
		So(testres.dsNoAuth.Calls, ShouldEqual, 1)
		expectedAuditOutcome(models.AuditTaskGetSearch, models.Scenario_attemptAndFail, testres)
	})

	Convey("Given the limit parameter in request is not a number return status 400 (bad request)", t, func() {
		testres := setupTest(testOpts{
			url: "http://localhost:23100/search/datasets/123/editions/2017/versions/1/dimensions/aggregate?q=term&limit=four",
		})
		So(testres.w.Code, ShouldEqual, http.StatusBadRequest)
		So(testres.w.Body.String(), ShouldContainSubstring, errs.ErrParsingQueryParameters.Error())
		expectedAuditOutcome(models.AuditTaskGetSearch, models.Scenario_attemptOnly, testres)
	})

	Convey("Given the offset parameter in request is not a number return status 400 (bad request)", t, func() {
		testres := setupTest(testOpts{
			url: "http://localhost:23100/search/datasets/123/editions/2017/versions/1/dimensions/aggregate?q=term&offset=fifty",
		})
		So(testres.w.Code, ShouldEqual, http.StatusBadRequest)
		So(testres.w.Body.String(), ShouldContainSubstring, errs.ErrParsingQueryParameters.Error())
		expectedAuditOutcome(models.AuditTaskGetSearch, models.Scenario_attemptOnly, testres)
	})

	Convey("Given the query parameter, q does not exist in request return status 400 (bad request)", t, func() {
		testres := setupTest(testOpts{
			url: "http://localhost:23100/search/datasets/123/editions/2017/versions/1/dimensions/aggregate",
		})
		So(testres.w.Code, ShouldEqual, http.StatusBadRequest)
		So(testres.w.Body.String(), ShouldContainSubstring, errs.ErrEmptySearchTerm.Error())
		expectedAuditOutcome(models.AuditTaskGetSearch, models.Scenario_attemptOnly, testres)
	})

	Convey("Given the offset parameter exceeds the default maximum results return status 400 (bad request)", t, func() {
		testres := setupTest(testOpts{
			url: "http://localhost:23100/search/datasets/123/editions/2017/versions/1/dimensions/aggregate?q=term&offset=500",
		})
		So(testres.w.Code, ShouldEqual, http.StatusBadRequest)
		So(testres.w.Body.String(), ShouldEqual, "the maximum offset has been reached, the offset cannot be more than "+strconv.Itoa(defaultMaxResults)+"\n")
		expectedAuditOutcome(models.AuditTaskGetSearch, models.Scenario_attemptOnly, testres)
	})

	Convey("Given search API fails to connect to elastic search cluster return status 500 (internal service error)", t, func() {
		testres := setupTest(testOpts{
			url:                   "http://localhost:23100/search/datasets/123/editions/2017/versions/1/dimensions/aggregate?q=term",
			esInternalServerError: true,
		})
		So(testres.w.Code, ShouldEqual, http.StatusInternalServerError)
		So(testres.w.Body.String(), ShouldEqual, "internal server error\n")
		expectedAuditOutcome(models.AuditTaskGetSearch, models.Scenario_attemptOnly, testres)
	})

	Convey("Given the search index does not exist but the version resource does then return status 500 (internal server error)", t, func() {
		testres := setupTest(testOpts{
			url:             "http://localhost:23100/search/datasets/123/editions/2017/versions/1/dimensions/aggregate?q=term",
			esIndexNotFound: true,
		})
		So(testres.w.Code, ShouldEqual, http.StatusInternalServerError)
		So(testres.w.Body.String(), ShouldContainSubstring, errs.ErrInternalServer.Error())
		expectedAuditOutcome(models.AuditTaskGetSearch, models.Scenario_attemptOnly, testres)
	})
}

// ensure no authentication is sent to the dataset API from public
func TestPublicSubnetUsersCannotSeeUnpublished(t *testing.T) {
	Convey("Given public subnet, when an authenticated GET is made, then the dataset api should not see authentication and returns not found, so we return status 404 (not found)", t, func() {
		testres := setupTest(testOpts{
			url:               "http://localhost:23100/search/datasets/123/editions/2017/versions/1/dimensions/aggregate?q=term",
			dsRequireNoAuth:   true,
			dsVersionNotFound: true,
		})
		So(testres.w.Code, ShouldEqual, http.StatusNotFound)
		So(testres.w.Body.String(), ShouldContainSubstring, errs.ErrVersionNotFound.Error())
		So(testres.dsWithAuth.Calls, ShouldEqual, 0)
		So(testres.dsNoAuth.Calls, ShouldEqual, 1)
		expectedAuditOutcome(models.AuditTaskGetSearch, models.Scenario_attemptAndFail, testres)
	})

	Convey("Given public subnet, when an unauthenticated GET is made, then the dataset api should not see authentication and returns not found, so we return status 404 (not found)", t, func() {
		testres := setupTest(testOpts{
			url:               "http://localhost:23100/search/datasets/123/editions/2017/versions/1/dimensions/aggregate?q=term",
			dsRequireNoAuth:   true,
			dsVersionNotFound: true,
		})
		So(testres.w.Code, ShouldEqual, http.StatusNotFound)
		So(testres.w.Body.String(), ShouldContainSubstring, errs.ErrVersionNotFound.Error())
		So(testres.dsWithAuth.Calls, ShouldEqual, 0)
		So(testres.dsNoAuth.Calls, ShouldEqual, 1)
		expectedAuditOutcome(models.AuditTaskGetSearch, models.Scenario_attemptAndFail, testres)
	})
}

// ensure authentication is sent to the dataset API appropriately (only when client is authenticated)
func TestPrivateSubnetMightSeeUnpublished(t *testing.T) {
	Convey("Given private subnet, when an authenticated GET is made, then the dataset api should see authentication and return ok, so we return OK", t, func() {
		testres := setupTest(testOpts{
			url:           "http://localhost:23100/search/datasets/123/editions/2017/versions/1/dimensions/aggregate?q=term",
			dsRequireAuth: true,
			reqHasAuth:    true,
			privateSubnet: true,
		})
		So(testres.w.Body.String(), ShouldStartWith, "{")
		So(testres.w.Code, ShouldEqual, http.StatusOK)
		So(testres.dsWithAuth.Calls, ShouldEqual, 1)
		So(testres.dsNoAuth.Calls, ShouldEqual, 0)
		expectedAuditOutcome(models.AuditTaskGetSearch, models.Scenario_attemptAndSucceed, testres)
	})

	Convey("Given private subnet, when an authenticated GET is made, force the dataset api to return 404 if authenticated, so we return 404", t, func() {
		testres := setupTest(testOpts{
			url:               "http://localhost:23100/search/datasets/123/editions/2017/versions/1/dimensions/aggregate?q=term",
			reqHasAuth:        true,
			dsRequireAuth:     true,
			dsVersionNotFound: true,
			privateSubnet:     true,
		})
		So(testres.w.Code, ShouldEqual, http.StatusNotFound)
		So(testres.w.Body.String(), ShouldContainSubstring, errs.ErrVersionNotFound.Error())
		expectedAuditOutcome(models.AuditTaskGetSearch, models.Scenario_attemptAndFail, testres)
	})
	Convey("Given private subnet, when an authenticated GET is made, force the dataset api to return 500 if authenticated, so we return 500", t, func() {
		testres := setupTest(testOpts{
			url:                   "http://localhost:23100/search/datasets/123/editions/2017/versions/1/dimensions/aggregate?q=term",
			dsInternalServerError: true,
			dsRequireAuth:         true,
			reqHasAuth:            true,
			privateSubnet:         true,
		})
		So(testres.w.Code, ShouldEqual, http.StatusInternalServerError)
		expectedAuditOutcome(models.AuditTaskGetSearch, models.Scenario_attemptAndFail, testres)
	})

	Convey("Given private subnet, when an unauthenticated GET is made, then the dataset api should see no authentication and return not found, so we return status 404 (not found)", t, func() {
		testres := setupTest(testOpts{
			url:               "http://localhost:23100/search/datasets/123/editions/2017/versions/1/dimensions/aggregate?q=term",
			dsVersionNotFound: true,
			dsRequireNoAuth:   true,
			privateSubnet:     true,
		})
		So(testres.w.Code, ShouldEqual, http.StatusNotFound)
		So(testres.dsNoAuth.Calls, ShouldEqual, 1)
		So(testres.w.Body.String(), ShouldContainSubstring, errs.ErrVersionNotFound.Error())

		So(testres.dsWithAuth.Calls, ShouldEqual, 0)
		So(testres.dsNoAuth.Calls, ShouldEqual, 1)
		expectedAuditOutcome(models.AuditTaskGetSearch, models.Scenario_attemptAndFail, testres)
	})

	Convey("Given private subnet, when a badly-authenticated GET is made, then the dataset api should see no authentication and returns not found, so we return server error", t, func() {
		testres := setupTest(testOpts{
			url:               "http://localhost:23100/search/datasets/123/editions/2017/versions/1/dimensions/aggregate?q=term",
			dsRequireNoAuth:   true,
			dsVersionNotFound: true,
			privateSubnet:     true,
		})
		So(testres.w.Code, ShouldEqual, http.StatusNotFound)
		So(testres.w.Body.String(), ShouldContainSubstring, errs.ErrVersionNotFound.Error())

		So(testres.dsWithAuth.Calls, ShouldEqual, 0)
		So(testres.dsNoAuth.Calls, ShouldEqual, 1)
		expectedAuditOutcome(models.AuditTaskGetSearch, models.Scenario_attemptAndFail, testres)
	})

	Convey("Given private subnet, when an unauthenticated GET is made, then the dataset api should see no authentication and return not found, so we return status 404 (not found)", t, func() {
		testres := setupTest(testOpts{
			url:               "http://localhost:23100/search/datasets/123/editions/2017/versions/1/dimensions/aggregate?q=term",
			dsRequireNoAuth:   true,
			dsVersionNotFound: true,
			privateSubnet:     true,
		})
		So(testres.w.Code, ShouldEqual, http.StatusNotFound)
		So(testres.w.Body.String(), ShouldContainSubstring, errs.ErrVersionNotFound.Error())
		expectedAuditOutcome(models.AuditTaskGetSearch, models.Scenario_attemptAndFail, testres)
	})
}

func TestCreateSearchIndexReturnsOK(t *testing.T) {
	Convey("Given instance and dimension exist return a status 200 (ok)", t, func() {
		testres := setupTest(testOpts{
			method:        "PUT",
			url:           "http://localhost:23100/search/instances/123/dimensions/aggregate",
			dsRequireAuth: true,
			reqHasAuth:    true,
			privateSubnet: true,
		})
		So(testres.w.Code, ShouldEqual, http.StatusOK)
		expectedAuditOutcome(models.AuditTaskCreateIndex, models.Scenario_attemptAndSucceed, testres)
	})
}

func TestFailToCreateSearchIndex(t *testing.T) {
	Convey("Given a request to create search index but no auth header is set return a status 401 (unauthorized)", t, func() {
		testres := setupTest(testOpts{
			method:          "PUT",
			url:             "http://localhost:23100/search/instances/123/dimensions/aggregate",
			dsRequireNoAuth: true,
			privateSubnet:   true,
		})
		So(testres.w.Code, ShouldEqual, http.StatusUnauthorized)
		So(testres.w.Body.String(), ShouldContainSubstring, errs.ErrUnauthenticatedRequest.Error())
	})

	Convey("Given a request to create search index but the auth header is wrong return a status 401 (unauthorized)", t, func() {
		testres := setupTest(testOpts{
			method:          "PUT",
			url:             "http://localhost:23100/search/instances/123/dimensions/aggregate",
			dsRequireNoAuth: true,
			privateSubnet:   true,
		})
		So(testres.w.Code, ShouldEqual, http.StatusUnauthorized)
		So(testres.w.Body.String(), ShouldContainSubstring, errs.ErrUnauthenticatedRequest.Error())
	})

	Convey("Given a request to create search index but unable to connect to kafka broker return a status 500 (internal service error)", t, func() {
		testres := setupTest(testOpts{
			method:            "PUT",
			url:               "http://localhost:23100/search/instances/123/dimensions/aggregate",
			dsRequireAuth:     true,
			searchReturnError: true,
			reqHasAuth:        true,
			privateSubnet:     true,
		})
		So(testres.w.Code, ShouldEqual, http.StatusInternalServerError)
		So(testres.w.Body.String(), ShouldContainSubstring, errs.ErrInternalServer.Error())
		expectedAuditOutcome(models.AuditTaskCreateIndex, models.Scenario_attemptAndFail, testres)
	})
}

func TestDeleteSearchIndexReturnsOK(t *testing.T) {
	Convey("Given a search index exists return a status 200 (ok)", t, func() {
		testres := setupTest(testOpts{
			method:        "DELETE",
			url:           "http://localhost:23100/search/instances/123/dimensions/aggregate",
			dsRequireAuth: true,
			reqHasAuth:    true,
			privateSubnet: true,
		})
		So(testres.w.Code, ShouldEqual, http.StatusOK)
		expectedAuditOutcome(models.AuditTaskDeleteIndex, models.Scenario_attemptAndSucceed, testres)
	})
}

func TestFailToDeleteSearchIndex(t *testing.T) {
	Convey("Given a search index exists but no auth header set return a status 401 (unauthorized)", t, func() {
		testres := setupTest(testOpts{
			method:          "DELETE",
			url:             "http://localhost:23100/search/instances/123/dimensions/aggregate",
			dsRequireNoAuth: true,
			privateSubnet:   true,
		})
		So(testres.w.Code, ShouldEqual, http.StatusUnauthorized)
		So(testres.w.Body.String(), ShouldContainSubstring, errs.ErrUnauthenticatedRequest.Error())
	})

	Convey("Given a search index exists but auth header is wrong return a status 401 (unauthorized)", t, func() {
		testres := setupTest(testOpts{
			method:        "DELETE",
			url:           "http://localhost:23100/search/instances/123/dimensions/aggregate",
			dsRequireAuth: true,
			privateSubnet: true,
		})
		So(testres.w.Code, ShouldEqual, http.StatusUnauthorized)
		So(testres.w.Body.String(), ShouldContainSubstring, errs.ErrUnauthenticatedRequest.Error())
	})

	Convey("Given a search index exists but unable to connect to elasticsearch cluster return a status 500 (internal service error)", t, func() {
		testres := setupTest(testOpts{
			method:                "DELETE",
			url:                   "http://localhost:23100/search/instances/123/dimensions/aggregate",
			dsRequireAuth:         true,
			esInternalServerError: true,
			privateSubnet:         true,
			reqHasAuth:            true,
		})
		So(testres.w.Code, ShouldEqual, http.StatusInternalServerError)
		So(testres.w.Body.String(), ShouldContainSubstring, errs.ErrInternalServer.Error())
		expectedAuditOutcome(models.AuditTaskDeleteIndex, models.Scenario_attemptAndFail, testres)
	})

	Convey("Given a search index does not exists return a status 404 (not found)", t, func() {
		testres := setupTest(testOpts{
			method:          "DELETE",
			url:             "http://localhost:23100/search/instances/123/dimensions/aggregate",
			dsRequireAuth:   true,
			esIndexNotFound: true,
			reqHasAuth:      true,
			privateSubnet:   true,
		})
		So(testres.w.Code, ShouldEqual, http.StatusNotFound)
		So(testres.w.Body.String(), ShouldContainSubstring, errs.ErrDeleteIndexNotFound.Error())
		expectedAuditOutcome(models.AuditTaskDeleteIndex, models.Scenario_attemptAndFail, testres)
	})
}

func TestCheckhighlights(t *testing.T) {
	Convey("Given the elasticsearch results contain highlights then the correct snippet pairs are returned", t, func() {
		result := models.HitList{
			Highlight: models.Highlight{
				Code:  []string{"\u0001Sstrangeness\u0001E"},
				Label: []string{"04 \u0001SHousing\u0001E, water, \u0001Selectricity\u0001E, gas and other fuels"},
			},
		}
		result = getSnippets(context.Background(), result)
		So(len(result.Source.Matches.Code), ShouldEqual, 1)
		So(result.Source.Matches.Code[0].Start, ShouldEqual, 1)
		So(result.Source.Matches.Code[0].End, ShouldEqual, 11)
		So(len(result.Source.Matches.Label), ShouldEqual, 2)
		So(result.Source.Matches.Label[0].Start, ShouldEqual, 4)
		So(result.Source.Matches.Label[0].End, ShouldEqual, 10)
		So(result.Source.Matches.Label[1].Start, ShouldEqual, 20)
		So(result.Source.Matches.Label[1].End, ShouldEqual, 30)
	})
}

func TestDeleteEndpointInWebReturnsNotFound(t *testing.T) {
	Convey("Given a search index exists and credentials are correct, return a status 404 (not found)", t, func() {
		testres := setupTest(testOpts{
			method:        "DELETE",
			url:           "http://localhost:23100/search/instances/123/dimensions/aggregate",
			dsRequireAuth: true,
			reqHasAuth:    true,
		})
		So(testres.w.Code, ShouldEqual, http.StatusNotFound)
		So(testres.w.Body.String(), ShouldEqual, "404 page not found\n")
	})

	Convey("Given a search index exists and credentials are incorrect, return a status 404 (not found)", t, func() {
		testres := setupTest(testOpts{
			method:          "DELETE",
			url:             "http://localhost:23100/search/instances/123/dimensions/aggregate",
			dsRequireNoAuth: true,
		})
		So(testres.w.Code, ShouldEqual, http.StatusNotFound)
		So(testres.w.Body.String(), ShouldContainSubstring, "404 page not found")
	})
}

func TestCreateSearchIndexEndpointInWebReturnsNotFound(t *testing.T) {
	Convey("Given instance and dimension exist and has valid auth return a status 404 (not found)", t, func() {
		testres := setupTest(testOpts{
			method:          "PUT",
			url:             "http://localhost:23100/search/instances/123/dimensions/aggregate",
			dsRequireNoAuth: true,
			reqHasAuth:      true,
		})
		So(testres.w.Code, ShouldEqual, http.StatusNotFound)
		So(testres.w.Body.String(), ShouldContainSubstring, "404 page not found")
	})

	Convey("Given a request to create search index and no private endpoints when a bad auth header is used, return a status 404 (not found)", t, func() {
		testres := setupTest(testOpts{
			method:          "PUT",
			url:             "http://localhost:23100/search/instances/123/dimensions/aggregate",
			dsRequireNoAuth: true,
		})
		So(testres.w.Code, ShouldEqual, http.StatusNotFound)
		So(testres.w.Body.String(), ShouldContainSubstring, "404 page not found")
	})
}

func getSearchResults(body *bytes.Buffer) *models.SearchResults {
	jsonBody, err := ioutil.ReadAll(body)
	if err != nil {
		os.Exit(1)
	}

	searchResults := &models.SearchResults{}
	if err := json.Unmarshal(jsonBody, searchResults); err != nil {
		os.Exit(1)
	}

	return searchResults
}

func getMockAuditor() *audit.AuditorServiceMock {
	return &audit.AuditorServiceMock{
		RecordFunc: func(ctx context.Context, action string, result string, params common.Params) error {
			return nil
		},
	}
}

func verifyAuditRecordCalls(c struct {
	Ctx    context.Context
	Action string
	Result string
	Params common.Params
}, expectedAction string, expectedResult string, expectedParams common.Params) {
	So(c.Action, ShouldEqual, expectedAction)
	So(c.Result, ShouldEqual, expectedResult)
	So(c.Params, ShouldResemble, expectedParams)
}

func expectedAuditOutcome(action, expectation string, testres testRes) {

	recCalls := testres.audit.RecordCalls()

	// Set expected params based on endpoint
	var expectedParams common.Params
	var expectedParamsOnAttempt common.Params
	if action == models.AuditTaskGetSearch {
		expectedParams = common.Params{"dataset_id": "123", "dimension": "aggregate", "edition": "2017", "version": "1"}
		expectedParamsOnAttempt = expectedParams
	} else {
		expectedParams = common.Params{"dimension": "aggregate", "instance_id": "123"}
		expectedParamsOnAttempt = common.Params{"caller_identity": "APIAmWhoAPIAm", "dimension": "aggregate", "instance_id": "123"}
	}

	// Test the relevant scenario
	switch expectation {
	case models.Scenario_attemptOnly:
		So(len(recCalls), ShouldEqual, 2)
		verifyAuditRecordCalls(recCalls[0], action, models.AuditActionAttempted, expectedParamsOnAttempt)
		verifyAuditRecordCalls(recCalls[1], action, models.AuditActionUnsuccessful, expectedParams)

	case models.Scenario_attemptAndSucceed:
		So(len(recCalls), ShouldEqual, 2)
		verifyAuditRecordCalls(recCalls[0], action, models.AuditActionAttempted, expectedParamsOnAttempt)
		verifyAuditRecordCalls(recCalls[1], action, models.AuditActionSuccessful, expectedParams)

	case models.Scenario_attemptAndFail:
		So(len(recCalls), ShouldEqual, 2)
		verifyAuditRecordCalls(recCalls[0], action, models.AuditActionAttempted, expectedParamsOnAttempt)
		verifyAuditRecordCalls(recCalls[1], action, models.AuditActionUnsuccessful, expectedParams)
	}
}
