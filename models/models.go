package models

import (
	"errors"
	"strconv"
)

// A list of errors that maybe returned from model
var (
	ErrorEmptySearchTerm = errors.New("search term empty")
)

func ErrorMaximumOffsetReached(m int) error {
	err := errors.New("the maximum offset has been reached, the offset cannot be more than " + strconv.Itoa(m))
	return err
}

type SearchResponse struct {
	Hits Hits `json:"hits"`
}

type Hits struct {
	Total   int       `json:"total"`
	HitList []HitList `json:"hits"`
}

type HitList struct {
	Highlight Highlight    `json:"highlight"`
	Score     float64      `json:"_score"`
	Source    SearchResult `json:"_source"`
}

type Highlight struct {
	Code  []string `json:"code,omitempty"`
	Label []string `json:"label,omitempty"`
}

// SearchResults represents a structure for a list of returned objects
type SearchResults struct {
	Count  int            `json:"count"`
	Items  []SearchResult `json:"items"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

// SearchResult represents data on a single item of search results
type SearchResult struct {
	Code               string  `json:"code"`
	URL                string  `json:"url,omitempty"`
	DimensionOptionURL string  `json:"dimension_option_url,omitempty"`
	HasData            bool    `json:"has_data"`
	Label              string  `json:"label"`
	Matches            Matches `json:"matches,omitempty"`
	NumberOfChildren   int     `json:"number_of_children"`
}

// Matches represents a list of members and their arrays of character offsets that matched the search term
type Matches struct {
	Code  []Snippet `json:"code,omitempty"`
	Label []Snippet `json:"label,omitempty"`
}

// Snippet represents a pair of integers defining the start and end of a substring in the member that matched the search term
type Snippet struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

// PageVariables are the necessary fields to determine paging
type PageVariables struct {
	DefaultMaxResults int
	Limit             int
	Offset            int
}

// ValidateQueryParameters represents a model for validating query parameters
func (page *PageVariables) ValidateQueryParameters(term string) error {
	if term == "" {
		return ErrorEmptySearchTerm
	}

	if page.Offset >= page.DefaultMaxResults {
		return ErrorMaximumOffsetReached(page.DefaultMaxResults)
	}

	if page.Offset+page.Limit > page.DefaultMaxResults {
		page.Limit = page.DefaultMaxResults - page.Offset
	}

	return nil
}
