package searchOutputQueue

import "github.com/ONSdigital/dp-import/events"

// Output is an object containing the search output queue channel
type Output struct {
	SearchOutputQueue chan []byte
}

// Search is an object containing the unique values to create a search index
type Search struct {
	Dimension  string
	InstanceID string
}

// CreateOutputQueue returns an object containing a channel for queueing filter outputs
func CreateOutputQueue(searchOutputQueue chan []byte) Output {
	return Output{SearchOutputQueue: searchOutputQueue}
}

// Queue represents a mechanism to add messages to the filter jobs queue
func (search *Output) Queue(outputSearch *Search) error {
	message := &events.HierarchyBuilt{
		DimensionName: outputSearch.Dimension,
		InstanceID:    outputSearch.InstanceID,
	}

	bytes, err := events.HierarchyBuiltSchema.Marshal(message)
	if err != nil {
		return err
	}

	search.SearchOutputQueue <- bytes

	return nil
}
