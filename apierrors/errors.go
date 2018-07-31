package apierrors

import "errors"

// A list of error messages for Search API
var (
	ErrDatasetNotFound        = errors.New("dataset not found")
	ErrDeleteIndexNotFound    = errors.New("search index not found")
	ErrEditionNotFound        = errors.New("edition not found")
	ErrEmptySearchTerm        = errors.New("empty search term")
	ErrIndexNotFound          = errors.New("search index not found")
	ErrInternalServer         = errors.New("internal server error")
	ErrMarshallingQuery       = errors.New("failed to marshal query to bytes for request body to send to elastic")
	ErrParsingQueryParameters = errors.New("failed to parse query parameters, values must be an integer")
	ErrUnauthenticatedRequest = errors.New("unauthenticated request")
	ErrUnmarshallingJSON      = errors.New("failed to parse json body")
	ErrUnexpectedStatusCode   = errors.New("unexpected status code from elastic api")
	ErrVersionNotFound        = errors.New("version not found")

	NotFoundMap = map[error]bool{
		ErrDatasetNotFound:     true,
		ErrDeleteIndexNotFound: true,
		ErrEditionNotFound:     true,
		ErrVersionNotFound:     true,
	}

	BadRequestMap = map[error]bool{
		ErrEmptySearchTerm:        true,
		ErrParsingQueryParameters: true,
	}
)
