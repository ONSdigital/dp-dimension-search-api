package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dp-search-api/mocks"
	"github.com/ONSdigital/dp-search-api/models"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)


func TestDeleteSearchIndexReturnsNotFoundWithValidAuthInWeb(t *testing.T) {
	Convey("Given a search index exists and valid auth return a status 404 (not found)", t, func() {
		r := httptest.NewRequest("DELETE", "http://localhost:23100/search/instances/123/dimensions/aggregate", nil)
		w := httptest.NewRecorder()
		r.Header.Add("internal-token", secretKey)

		api := routes(host, secretKey, datasetAPISecretKey, mux.NewRouter(), &mocks.BuildSearch{}, &mocks.DatasetAPI{}, &mocks.Elasticsearch{IndexNotFound: true}, defaultMaxResults, models.DisablePrivateEndpoints)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
	})
}


func TestDeleteSearchIndexReturnsNotFoundWithoutValidAuthInWeb(t *testing.T) {
	Convey("Given a search index exists and no valid auth return a status 404 (not found)", t, func() {
		r := httptest.NewRequest("DELETE", "http://localhost:23100/search/instances/123/dimensions/aggregate", nil)
		w := httptest.NewRecorder()
		r.Header.Add("internal-token", "")

		api := routes(host, secretKey, datasetAPISecretKey, mux.NewRouter(), &mocks.BuildSearch{}, &mocks.DatasetAPI{}, &mocks.Elasticsearch{IndexNotFound: true}, defaultMaxResults, models.DisablePrivateEndpoints)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
	})
}


func TestCreateSearchIndexReturnsNotFoundWithValidAuthInWeb(t *testing.T) {
	Convey("Even if instance and dimension exist return a status 404 (not found)", t, func() {
		r := httptest.NewRequest("PUT", "http://localhost:23100/search/instances/123/dimensions/aggregate", nil)
		w := httptest.NewRecorder()
		r.Header.Add("internal-token", secretKey)

		api := routes(host, secretKey, datasetAPISecretKey, mux.NewRouter(), &mocks.BuildSearch{}, &mocks.DatasetAPI{}, &mocks.Elasticsearch{}, defaultMaxResults, models.DisablePrivateEndpoints)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
	})
}

func TestCreateSearchIndexReturnsNotFoundWithoutValidAuthInWeb(t *testing.T) {
	Convey("Even if instance and dimension exist return a status 404 (not found)", t, func() {
		r := httptest.NewRequest("PUT", "http://localhost:23100/search/instances/123/dimensions/aggregate", nil)
		w := httptest.NewRecorder()
		r.Header.Add("internal-token", "")

		api := routes(host, secretKey, datasetAPISecretKey, mux.NewRouter(), &mocks.BuildSearch{}, &mocks.DatasetAPI{}, &mocks.Elasticsearch{}, defaultMaxResults, models.DisablePrivateEndpoints)
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
	})
}

