package elasticsearch_test

import (
	"context"
	"github.com/ONSdigital/dp-dimension-search-api/elasticsearch"
	"github.com/ONSdigital/log.go/v2/log"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTransformResponse_ES6(t *testing.T) {
	Convey("Given an ES6 response with int total", t, func() {
		body := []byte(`{"hits":{"total":6}}`)

		Convey("When the response is transformed", func() {

			response, err := elasticsearch.TransformResponse(context.Background(), body, log.Data{})

			Convey("There should be no error returned", func() {
				So(err, ShouldBeNil)
				So(response.Hits.Total, ShouldEqual, 6)
			})
		})
	})
}

func TestTransformResponse_ES7(t *testing.T) {
	Convey("Given an ES7 response with object total", t, func() {
		body := []byte(`{"hits":{"total": {"value": 7,"relation": "eq"}}}`)

		Convey("When the response is transformed", func() {

			response, err := elasticsearch.TransformResponse(context.Background(), body, log.Data{})

			Convey("There should be no error returned", func() {
				So(err, ShouldBeNil)
				So(response.Hits.Total, ShouldEqual, 7)
			})
		})
	})
}
