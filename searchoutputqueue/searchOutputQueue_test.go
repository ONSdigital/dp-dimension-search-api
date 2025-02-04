package searchoutputqueue

import (
	"context"
	"testing"

	"github.com/ONSdigital/dp-import/events"
	kafka "github.com/ONSdigital/dp-kafka/v4"
	"github.com/smartystreets/goconvey/convey"
)

func TestFilterOuputQueue(t *testing.T) {
	convey.Convey("When a search output is created, a message is sent to kafka", t, func() {
		searchOutputQueue := make(chan kafka.BytesMessage, 1)
		outputQueue := CreateOutputQueue(searchOutputQueue)
		search := Search{InstanceID: "12345678", Dimension: "aggregate"}
		err := outputQueue.Queue(context.Background(), &search)
		convey.So(err, convey.ShouldBeNil)

		message := <-searchOutputQueue
		convey.So(message, convey.ShouldNotBeNil)

		var searchMessage events.HierarchyBuilt
		err = events.HierarchyBuiltSchema.Unmarshal(message.Value, &searchMessage)
		convey.So(err, convey.ShouldBeNil)

		convey.So(searchMessage.InstanceID, convey.ShouldEqual, search.InstanceID)
		convey.So(searchMessage.DimensionName, convey.ShouldEqual, search.Dimension)
	})
}
