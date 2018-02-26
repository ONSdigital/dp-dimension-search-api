package searchOutputQueue

import (
	"testing"

	"github.com/ONSdigital/dp-import/events"
	. "github.com/smartystreets/goconvey/convey"
)

func TestFilterOuputQueue(t *testing.T) {
	Convey("When a search output is created, a message is sent to kafka", t, func() {
		searchOutputQueue := make(chan []byte, 1)
		outputQueue := CreateOutputQueue(searchOutputQueue)
		search := Search{InstanceID: "12345678", Dimension: "aggregate"}
		err := outputQueue.Queue(&search)
		So(err, ShouldBeNil)

		bytes := <-searchOutputQueue
		var searchMessage events.HierarchyBuilt
		events.HierarchyBuiltSchema.Unmarshal(bytes, &searchMessage)
		So(searchMessage.InstanceID, ShouldEqual, search.InstanceID)
		So(searchMessage.DimensionName, ShouldEqual, search.Dimension)
	})
}
