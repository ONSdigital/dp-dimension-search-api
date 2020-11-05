package mocks

import (
	"fmt"

	"github.com/ONSdigital/dp-dimension-search-api/searchoutputqueue"
)

// BuildSearch contains a flag indicating whether the message failed to go on queue
type BuildSearch struct {
	ReturnError bool
}

// MessageData contains the unique identifiers for search message
type MessageData struct {
	Dimension  string
	InstanceID string
}

// Queue checks whether the filter job has errored
func (bs *BuildSearch) Queue(search *searchoutputqueue.Search) error {
	if bs.ReturnError {
		return fmt.Errorf("No message produced for hierarchy built")
	}
	return nil
}
