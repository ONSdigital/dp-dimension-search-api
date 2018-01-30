package api

import (
	"testing"

	"github.com/ONSdigital/dp-search-api/models"
	. "github.com/smartystreets/goconvey/convey"
)

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
