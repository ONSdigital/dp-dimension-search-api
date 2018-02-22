package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/ONSdigital/dp-search-api/models"
)

func TestMiddleWareAuthenticationReturnsForbiddenInPublishing(t *testing.T) {
	t.Parallel()
	Convey("When no access token is provide, unauthorised status code is returned", t, func() {
		auth := &Authenticator{"123", "internal-token", models.SubnetPublishing}
		r, err := http.NewRequest("POST", "http://localhost:23100/search/datasets/123/editions/2017/versions/1/dimensions/aggregate", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		auth.Check(mockHTTPHandler).ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusUnauthorized)
	})
}

func TestMiddleWareAuthenticationReturnsUnauthorisedInPublishing(t *testing.T) {
	t.Parallel()
	Convey("When a invalid access token is provide, unauthorised status code is returned", t, func() {
		auth := &Authenticator{"123", "internal-token", models.SubnetPublishing}
		r, err := http.NewRequest("POST", "http://localhost:23100/search/datasets/123/editions/2017/versions/1/dimensions/aggregate", nil)
		r.Header.Set("internal-token", "12")
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		auth.Check(mockHTTPHandler).ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusUnauthorized)
	})
}

func TestMiddleWareAuthenticationInPublishing(t *testing.T) {
	t.Parallel()
	Convey("When a valid access token is provide, OK code is returned", t, func() {
		auth := &Authenticator{"123", "internal-token",models.SubnetPublishing}
		r, err := http.NewRequest("POST", "http://localhost:23100/search/datasets/123/editions/2017/versions/1/dimensions/aggregate", nil)
		r.Header.Set("internal-token", "123")
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		auth.Check(mockHTTPHandler).ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
	})
}

func TestMiddleWareAuthenticationWithValueInPublishing(t *testing.T) {
	t.Parallel()
	Convey("When a valid access token is provide, true is passed to a http handler", t, func() {
		auth := &Authenticator{"123", "internal-token", models.SubnetPublishing}
		r, err := http.NewRequest("POST", "http://localhost:23100/search/datasets/123/editions/2017/versions/1/dimensions/aggregate", nil)
		r.Header.Set("internal-token", "123")
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		var isRequestAuthenticated bool
		auth.ManualCheck(func(w http.ResponseWriter, r *http.Request, isAuth bool) {
			isRequestAuthenticated = isAuth
		}).ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
		So(isRequestAuthenticated, ShouldEqual, true)
	})
}

func TestMiddleWareAuthenticationWithValidTokenReturnsStatusNotFoundInWeb(t *testing.T) {
	t.Parallel()
	Convey("When a valid access token is provided in web Subnet, status not found is returned", t, func() {
		auth := &Authenticator{"123", "internal-token", models.SubnetWeb}
		r, err := http.NewRequest("POST", "http://localhost:23100/search/datasets/123/editions/2017/versions/1/dimensions/aggregate", nil)
		r.Header.Set("internal-token", "123")
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		auth.Check(mockHTTPHandler).ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
	})
}

func TestMiddleWareAuthenticationWithoutValidTokenReturnsStatusNotFoundInWeb(t *testing.T) {
	t.Parallel()
	Convey("When an invalid token is provided in web Subnet, status not found is returned", t, func() {
		auth := &Authenticator{"123", "internal-token", models.SubnetWeb}
		r, err := http.NewRequest("POST", "http://localhost:23100/search/datasets/123/editions/2017/versions/1/dimensions/aggregate", nil)
		r.Header.Set("internal-token", "12")
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		auth.Check(mockHTTPHandler).ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
	})
}


// mockHTTPHandler is an empty http handler used for testing auth check function
func mockHTTPHandler(w http.ResponseWriter, r *http.Request) {}
