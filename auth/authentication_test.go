package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/ONSdigital/dp-search-api/models"
)


/*
This block of tests refers to the service as run in the publishing subnet,
*/

func TestMiddleWareAuthenticationReturnsForbiddenInPublishing(t *testing.T) {
	t.Parallel()
	Convey("When no access token is provided in publishing subnet, unauthorised status code is returned", t, func() {
		auth := &Authenticator{"123", "internal-token", models.EnablePrivateEndpoints}
		r, err := http.NewRequest("POST", "http://localhost:23100/search/datasets/123/editions/2017/versions/1/dimensions/aggregate", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		auth.Check(mockHTTPHandler).ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusUnauthorized)
	})
}

func TestMiddleWareAuthenticationReturnsUnauthorisedInPublishing(t *testing.T) {
	t.Parallel()
	Convey("When a invalid access token is provided in publishing subnet, unauthorised status code is returned", t, func() {
		auth := &Authenticator{"123", "internal-token", models.EnablePrivateEndpoints}
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
	Convey("When a valid access token is provided in publishing subnet, OK code is returned", t, func() {
		auth := &Authenticator{"123", "internal-token",models.EnablePrivateEndpoints}
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
	Convey("When a valid access token is provided in publishing subnet, true is passed to a http handler", t, func() {
		auth := &Authenticator{"123", "internal-token", models.EnablePrivateEndpoints}
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


/*
This block of tests refers to the service as run in the web subnet.
*/

func TestMiddleWareAuthenticationWithValidTokenReturnsStatusNotFoundInWeb(t *testing.T) {
	t.Parallel()
	Convey("When a valid access token is provided in web Subnet, status not found is returned", t, func() {
		auth := &Authenticator{"123", "internal-token", models.DisablePrivateEndpoints}
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
		auth := &Authenticator{"123", "internal-token", models.DisablePrivateEndpoints}
		r, err := http.NewRequest("POST", "http://localhost:23100/search/datasets/123/editions/2017/versions/1/dimensions/aggregate", nil)
		r.Header.Set("internal-token", "12")
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		auth.Check(mockHTTPHandler).ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
	})
}

func TestMiddleWareAuthenticationInWeb(t *testing.T) {
	t.Parallel()
	Convey("When a valid access token is provided in web subnet, status not found is returned", t, func() {
		auth := &Authenticator{"123", "internal-token",models.DisablePrivateEndpoints}
		r, err := http.NewRequest("POST", "http://localhost:23100/search/datasets/123/editions/2017/versions/1/dimensions/aggregate", nil)
		r.Header.Set("internal-token", "123")
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		auth.Check(mockHTTPHandler).ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
	})
}

func TestMiddleWareAuthenticationWithValueInWeb(t *testing.T) {
	t.Parallel()
	Convey("When a valid access token is provided in web subnet, false is passed to a http handler", t, func() {
		auth := &Authenticator{"123", "internal-token", models.DisablePrivateEndpoints}
		r, err := http.NewRequest("POST", "http://localhost:23100/search/datasets/123/editions/2017/versions/1/dimensions/aggregate", nil)
		r.Header.Set("internal-token", "123")
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		var isRequestAuthenticated bool
		auth.ManualCheck(func(w http.ResponseWriter, r *http.Request, isAuth bool) {
			isRequestAuthenticated = isAuth
		}).ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
		So(isRequestAuthenticated, ShouldEqual, false)
	})
}

// mockHTTPHandler is an empty http handler used for testing auth check function
func mockHTTPHandler(w http.ResponseWriter, r *http.Request) {}
