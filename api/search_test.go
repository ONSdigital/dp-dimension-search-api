package api

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/ONSdigital/dp-search-api/mocks"
	"github.com/ONSdigital/dp-search-api/models"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	host                = "8080"
	secretKey           = "coffee"
	datasetAPISecretKey = "tea"
	subnet				= models.SubnetWeb   // will need to be switched to models.SubnetPublishing if auth is expected to pass.
	defaultMaxResults   = 20
	brokers             = []string{"localhost:9092"}
	topic               = "testing"
)

func TestGetSearchReturnsOK(t *testing.T) {
	t.Parallel()
	Convey("Given the search query satisfies the search index then return a status 200", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:23100/search/datasets/123/editions/2017/versions/1/dimensions/aggregate?q=term", nil)
		w := httptest.NewRecorder()

		api := routes(host, secretKey, datasetAPISecretKey, subnet, mux.NewRouter(), &mocks.DatasetAPI{}, &mocks.Elasticsearch{}, defaultMaxResults)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)

		// Check response json
		searchResults := getSearchResults(w.Body)

		So(searchResults.Count, ShouldEqual, 2)
		So(len(searchResults.Items), ShouldEqual, 2)
		So(searchResults.Limit, ShouldEqual, 20)
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
	})

	Convey("Given the search query satisfies the search index when limit and offset parameters are set then return a status 200", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:23100/search/datasets/123/editions/2017/versions/1/dimensions/aggregate?q=term&limit=5&offset=20", nil)
		w := httptest.NewRecorder()

		api := routes(host, secretKey, datasetAPISecretKey, subnet, mux.NewRouter(), &mocks.DatasetAPI{}, &mocks.Elasticsearch{}, 40)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)

		// Check response json
		searchResults := getSearchResults(w.Body)

		So(searchResults.Count, ShouldEqual, 2)
		So(len(searchResults.Items), ShouldEqual, 2)
		So(searchResults.Limit, ShouldEqual, 5)
		So(searchResults.Offset, ShouldEqual, 20)
	})
}

func TestGetSearchFailureScenarios(t *testing.T) {
	t.Parallel()
	Convey("Given search API fails to connect to the dataset API return status 500 (internal service error)", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:23100/search/datasets/123/editions/2017/versions/1/dimensions/aggregate?q=term", nil)
		w := httptest.NewRecorder()

		api := routes(host, secretKey, datasetAPISecretKey, subnet, mux.NewRouter(), &mocks.DatasetAPI{InternalServerError: true}, &mocks.Elasticsearch{}, defaultMaxResults)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Given the version document was not found via the dataset API return status 404 (not found)", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:23100/search/datasets/123/editions/2017/versions/1/dimensions/aggregate?q=term", nil)
		w := httptest.NewRecorder()

		api := routes(host, secretKey, datasetAPISecretKey, subnet, mux.NewRouter(), &mocks.DatasetAPI{VersionNotFound: true}, &mocks.Elasticsearch{}, defaultMaxResults)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
	})

	Convey("Given the limit parameter in request is not a number return status 400 (bad request)", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:23100/search/datasets/123/editions/2017/versions/1/dimensions/aggregate?q=term&limit=four", nil)
		w := httptest.NewRecorder()

		api := routes(host, secretKey, datasetAPISecretKey, subnet, mux.NewRouter(), &mocks.DatasetAPI{}, &mocks.Elasticsearch{}, defaultMaxResults)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Given the offset parameter in request is not a number return status 400 (bad request)", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:23100/search/datasets/123/editions/2017/versions/1/dimensions/aggregate?q=term&offset=fifty", nil)
		w := httptest.NewRecorder()

		api := routes(host, secretKey, datasetAPISecretKey, subnet, mux.NewRouter(), &mocks.DatasetAPI{}, &mocks.Elasticsearch{}, defaultMaxResults)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Given the query parameter, q does not exist in request return status 400 (bad request)", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:23100/search/datasets/123/editions/2017/versions/1/dimensions/aggregate", nil)
		w := httptest.NewRecorder()

		api := routes(host, secretKey, datasetAPISecretKey, subnet, mux.NewRouter(), &mocks.DatasetAPI{}, &mocks.Elasticsearch{}, defaultMaxResults)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Given the offset parameter exceeds the default maximum results return status 400 (bad request)", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:23100/search/datasets/123/editions/2017/versions/1/dimensions/aggregate?q=term&offset=50", nil)
		w := httptest.NewRecorder()

		api := routes(host, secretKey, datasetAPISecretKey, subnet, mux.NewRouter(), &mocks.DatasetAPI{}, &mocks.Elasticsearch{}, defaultMaxResults)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Given search API fails to connect to elastic search cluster return status 500 (internal service error)", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:23100/search/datasets/123/editions/2017/versions/1/dimensions/aggregate?q=term", nil)
		w := httptest.NewRecorder()

		api := routes(host, secretKey, datasetAPISecretKey, subnet, mux.NewRouter(), &mocks.DatasetAPI{}, &mocks.Elasticsearch{InternalServerError: true}, defaultMaxResults)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Given the search index does not exist return status 404 (not found)", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:23100/search/datasets/123/editions/2017/versions/1/dimensions/aggregate?q=term", nil)
		w := httptest.NewRecorder()

		api := routes(host, secretKey, datasetAPISecretKey, subnet, mux.NewRouter(), &mocks.DatasetAPI{}, &mocks.Elasticsearch{IndexNotFound: true}, defaultMaxResults)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
	})
}

func TestDeleteSearchIndex(t *testing.T) {
	Convey("Given a search index exists return a status 200 (ok)", t, func() {
		r := httptest.NewRequest("DELETE", "http://localhost:23100/search/instances/123/dimensions/aggregate", nil)
		w := httptest.NewRecorder()
		r.Header.Add("internal-token", secretKey)

		api := routes(host, secretKey, datasetAPISecretKey, subnet, mux.NewRouter(), &mocks.DatasetAPI{}, &mocks.Elasticsearch{}, defaultMaxResults)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
	})
}

func TestFailToDeleteSearchIndex(t *testing.T) {
	Convey("Given a search index exists but no auth header set return a status 401 (unauthorised)", t, func() {
		r := httptest.NewRequest("DELETE", "http://localhost:23100/search/instances/123/dimensions/aggregate", nil)
		w := httptest.NewRecorder()

		api := routes(host, secretKey, datasetAPISecretKey, subnet, mux.NewRouter(), &mocks.DatasetAPI{}, &mocks.Elasticsearch{}, defaultMaxResults)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusUnauthorized)
	})

	Convey("Given a search index exists but auth header is wrong return a status 401 (unauthorised)", t, func() {
		r := httptest.NewRequest("DELETE", "http://localhost:23100/search/instances/123/dimensions/aggregate", nil)
		w := httptest.NewRecorder()
		r.Header.Add("internal-token", "abcdef")

		api := routes(host, secretKey, datasetAPISecretKey, subnet, mux.NewRouter(), &mocks.DatasetAPI{}, &mocks.Elasticsearch{}, defaultMaxResults)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusUnauthorized)
	})

	Convey("Given a search index exists but unable to connect to elasticsearch cluster return a status 500 (internal service error)", t, func() {
		r := httptest.NewRequest("DELETE", "http://localhost:23100/search/instances/123/dimensions/aggregate", nil)
		w := httptest.NewRecorder()
		r.Header.Add("internal-token", secretKey)

		api := routes(host, secretKey, datasetAPISecretKey, subnet, mux.NewRouter(), &mocks.DatasetAPI{}, &mocks.Elasticsearch{InternalServerError: true}, defaultMaxResults)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Given a search index does not exists return a status 404 (not found)", t, func() {
		r := httptest.NewRequest("DELETE", "http://localhost:23100/search/instances/123/dimensions/aggregate", nil)
		w := httptest.NewRecorder()
		r.Header.Add("internal-token", secretKey)

		api := routes(host, secretKey, datasetAPISecretKey, subnet, mux.NewRouter(), &mocks.DatasetAPI{}, &mocks.Elasticsearch{IndexNotFound: true}, defaultMaxResults)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
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
		result = getSnippets(result)
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
