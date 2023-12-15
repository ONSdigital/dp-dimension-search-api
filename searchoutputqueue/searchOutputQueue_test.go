package searchoutputqueue

import (
	"context"
	"testing"

	"github.com/ONSdigital/dp-import/events"
	kafka "github.com/ONSdigital/dp-kafka/v4"
	. "github.com/smartystreets/goconvey/convey"
)

func TestFilterOuputQueue(t *testing.T) {
	Convey("When a search output is created, a message is sent to kafka", t, func() {
		searchOutputQueue := make(chan kafka.BytesMessage, 1)
		outputQueue := CreateOutputQueue(searchOutputQueue)
		search := Search{InstanceID: "12345678", Dimension: "aggregate"}
		err := outputQueue.Queue(context.Background(), &search)
		So(err, ShouldBeNil)

		message := <-searchOutputQueue
		So(message, ShouldNotBeNil)

		var searchMessage events.HierarchyBuilt
		events.HierarchyBuiltSchema.Unmarshal(message.Value, &searchMessage)

		So(searchMessage.InstanceID, ShouldEqual, search.InstanceID)
		So(searchMessage.DimensionName, ShouldEqual, search.Dimension)
	})
}
