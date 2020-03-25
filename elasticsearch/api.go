package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	rchttp "github.com/ONSdigital/dp-rchttp"
	errs "github.com/ONSdigital/dp-search-api/apierrors"
	"github.com/ONSdigital/dp-search-api/models"
	"github.com/ONSdigital/log.go/log"
	"github.com/smartystreets/go-aws-auth"
)

// API aggregates a client and URL and other common data for accessing the API
type API struct {
	client       rchttp.Clienter
	url          string
	signRequests bool
}

// NewElasticSearchAPI creates an API object
func NewElasticSearchAPI(client rchttp.Clienter, elasticSearchAPIURL string, signRequests bool) *API {
	return &API{
		client:       client,
		url:          elasticSearchAPIURL,
		signRequests: signRequests,
	}
}

// DeleteSearchIndex removes an index from elasticsearch
func (api *API) DeleteSearchIndex(ctx context.Context, instanceID, dimension string) (int, error) {
	path := api.url + "/" + instanceID + "_" + dimension

	_, status, err := api.CallElastic(ctx, path, "DELETE", nil)
	if err != nil {
		if status == http.StatusNotFound {
			return status, errs.ErrDeleteIndexNotFound
		}

		return status, err
	}

	return status, nil
}

// QuerySearchIndex builds query as a json body to call an elasticsearch index with
func (api *API) QuerySearchIndex(ctx context.Context, instanceID, dimension, term string, limit, offset int) (*models.SearchResponse, int, error) {
	response := &models.SearchResponse{}

	path := api.url + "/" + instanceID + "_" + dimension + "/_search"

	logData := log.Data{"term": term, "path": path}

	log.Event(ctx, "searching index", log.INFO, logData)

	body := buildSearchQuery(term, limit, offset)

	bytes, err := json.Marshal(body)
	if err != nil {
		log.Event(ctx, "unable to marshal elastic search query to bytes", log.ERROR, log.Error(err), logData)
		return nil, 0, errs.ErrMarshallingQuery
	}

	logData["request_body"] = string(bytes)

	responseBody, status, err := api.CallElastic(ctx, path, "GET", bytes)
	logData["status"] = status
	if err != nil {
		log.Event(ctx, "failed to call elasticsearch", log.ERROR, log.Error(err), logData)
		return nil, status, errs.ErrIndexNotFound
	}

	logData["response_body"] = string(responseBody)

	if err = json.Unmarshal(responseBody, response); err != nil {
		log.Event(ctx, "unable to unmarshal json body", log.ERROR, log.Error(err), logData)
		return nil, status, errs.ErrUnmarshallingJSON
	}

	log.Event(ctx, "search results", log.INFO, logData)

	return response, status, nil
}

// CallElastic builds a request to elastic search based on the method, path and payload
func (api *API) CallElastic(ctx context.Context, path, method string, payload interface{}) ([]byte, int, error) {
	logData := log.Data{"url": path, "method": method}

	URL, err := url.Parse(path)
	if err != nil {
		log.Event(ctx, "failed to create url for elastic call", log.ERROR, log.Error(err), logData)
		return nil, 0, err
	}
	path = URL.String()
	logData["url"] = path

	var req *http.Request

	if payload != nil {
		req, err = http.NewRequest(method, path, bytes.NewReader(payload.([]byte)))
		req.Header.Add("Content-type", "application/json")
		logData["payload"] = string(payload.([]byte))
	} else {
		req, err = http.NewRequest(method, path, nil)
	}
	// check req, above, didn't error
	if err != nil {
		log.Event(ctx, "failed to create request for call to elastic", log.ERROR, log.Error(err), logData)
		return nil, 0, err
	}

	if api.signRequests {
		awsauth.Sign(req)
	}

	resp, err := api.client.Do(ctx, req)
	if err != nil {
		log.Event(ctx, "failed to call elastic", log.ERROR, log.Error(err), logData)
		return nil, 0, err
	}
	defer resp.Body.Close()

	logData["status_code"] = resp.StatusCode

	jsonBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Event(ctx, "failed to read response body from call to elastic", log.ERROR, log.Error(err), logData)
		return nil, resp.StatusCode, err
	}
	logData["json_body"] = string(jsonBody)

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= 300 {
		log.Event(ctx, errs.ErrUnexpectedStatusCode.Error(), log.ERROR, log.Error(errs.ErrUnexpectedStatusCode), logData)
		return nil, resp.StatusCode, errs.ErrUnexpectedStatusCode
	}

	return jsonBody, resp.StatusCode, nil
}

// Body represents the request body to elasticsearch
type Body struct {
	From      int        `json:"from"`
	Size      int        `json:"size"`
	Highlight *Highlight `json:"highlight,omitempty"`
	Query     Query      `json:"query"`
	Sort      []Scores   `json:"sort"`
}

// Highlight represents parts of the fields that matched
type Highlight struct {
	PreTags  []string          `json:"pre_tags,omitempty"`
	PostTags []string          `json:"post_tags,omitempty"`
	Fields   map[string]Object `json:"fields,omitempty"`
	Order    string            `json:"score,omitempty"`
}

// Object represents an empty object (as expected by elasticsearch)
type Object struct{}

// Query represents the request query details
type Query struct {
	Bool Bool `json:"bool"`
}

// Bool represents the desirable goals for query
type Bool struct {
	Must   []Match `json:"must,omitempty"`
	Should []Match `json:"should,omitempty"`
}

// Match represents the fields that the term should or must match within query
type Match struct {
	Match map[string]string `json:"match,omitempty"`
}

// Scores represents a list of scoring, e.g. scoring on relevance, but can add in secondary
// score such as alphabetical order if relevance is the same for two search results
type Scores struct {
	Score Score `json:"_score"`
}

// Score contains the ordering of the score (ascending or descending)
type Score struct {
	Order string `json:"order"`
}

func buildSearchQuery(term string, limit, offset int) *Body {
	var object Object
	highlight := make(map[string]Object)

	highlight["label"] = object
	highlight["code"] = object

	label := make(map[string]string)
	code := make(map[string]string)
	label["label"] = term
	code["code"] = term

	labelMatch := Match{
		Match: label,
	}

	codeMatch := Match{
		Match: code,
	}

	scores := Scores{
		Score: Score{
			Order: "desc",
		},
	}

	listOfScores := []Scores{}
	listOfScores = append(listOfScores, scores)

	query := &Body{
		From: offset,
		Size: limit,
		Highlight: &Highlight{
			PreTags:  []string{"\u0001S"},
			PostTags: []string{"\u0001E"},
			Fields:   highlight,
		},
		Query: Query{
			Bool: Bool{
				Should: []Match{
					labelMatch,
					codeMatch,
				},
			},
		},
		Sort: listOfScores,
	}

	return query
}
