package models

import (
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var p = PageVariables{DefaultMaxResults: 1000, Limit: 0, Offset: 0}

func TestValidateQueryParameters(t *testing.T) {
	Convey("Given the query term is NOT empty return without an error", t, func() {
		err := p.ValidateQueryParameters("term")
		So(err, ShouldBeNil)
		So(p.Limit, ShouldEqual, 0)
		So(p.Offset, ShouldEqual, 0)
	})

	Convey("Given the query term is NOT empty and the combined sum of the offset and limit does not exceed the maximum number of results return without an error", t, func() {
		p.Limit = 30
		p.Offset = 60
		err := p.ValidateQueryParameters("term")
		So(err, ShouldBeNil)
		So(p.Limit, ShouldEqual, 30)
		So(p.Offset, ShouldEqual, 60)
	})

	Convey("Given the query term is NOT empty and the offset does not exceed the maximum number of results return without an error", t, func() {
		p.Limit = 30
		p.Offset = 985
		err := p.ValidateQueryParameters("term")
		So(err, ShouldBeNil)
		So(p.Limit, ShouldEqual, 15) // Limit should be reduced as the combined limit and offset should not exceed the default maximum results
		So(p.Offset, ShouldEqual, 985)
	})

	Convey("Given the query term is empty return with an error", t, func() {
		err := p.ValidateQueryParameters("")
		So(err, ShouldNotBeEmpty)
		So(err, ShouldResemble, errors.New("empty search term"))
	})

	Convey("Given the query term is NOT empty and the offset exceeds the maximum number of results return with an error", t, func() {
		p.Limit = 30
		p.Offset = 1200
		err := p.ValidateQueryParameters("term")
		So(err, ShouldResemble, errors.New("the maximum offset has been reached, the offset cannot be more than 1000"))
	})
}
