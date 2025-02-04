package searchoutputqueue

import (
	"context"

	"github.com/ONSdigital/dp-import/events"
	kafka "github.com/ONSdigital/dp-kafka/v4"
)

// Output is an object containing the search output queue channel
type Output struct {
	searchOutputQueue chan kafka.BytesMessage
}

// Search is an object containing the unique values to create a search index
type Search struct {
	Dimension  string
	InstanceID string
}

// CreateOutputQueue returns an object containing a channel for queueing filter outputs
func CreateOutputQueue(queue chan kafka.BytesMessage) Output {
	return Output{searchOutputQueue: queue}
}

// Queue represents a mechanism to add messages to the filter jobs queue
func (search *Output) Queue(ctx context.Context, outputSearch *Search) error {
	message := &events.HierarchyBuilt{
		DimensionName: outputSearch.Dimension,
		InstanceID:    outputSearch.InstanceID,
	}

	bytes, err := events.HierarchyBuiltSchema.Marshal(message)
	if err != nil {
		return err
	}

	search.searchOutputQueue <- kafka.BytesMessage{Value: bytes, Context: ctx}

	return nil
}
