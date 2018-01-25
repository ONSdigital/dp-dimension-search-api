package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/ONSdigital/dp-search-api/models"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/rchttp"
)

// ErrorUnexpectedStatusCode represents the error message to be returned when
// the status received from elastic is not as expected
var ErrorUnexpectedStatusCode = errors.New("unexpected status code from api")

// API aggregates a client and URL and other common data for accessing the API
type API struct {
	client *rchttp.Client
	url    string
}

// NewElasticSearchAPI creates an ElasticSearchAPI object
func NewElasticSearchAPI(client *rchttp.Client, elasticSearchAPIURL string) *API {
	return &API{
		client: client,
		url:    elasticSearchAPIURL,
	}
}

// DeleteSearchIndex removes an index from elastic search
func (api *API) DeleteSearchIndex(ctx context.Context, instanceID, dimension string) (int, error) {
	path := api.url + "/" + instanceID + "_" + dimension

	_, status, err := api.CallElastic(ctx, path, "DELETE", nil)
	if err != nil {
		return status, err
	}

	return status, nil
}

// QuerySearchIndex builds query to call elasticsearch index with
func (api *API) QuerySearchIndex(ctx context.Context, instanceID, dimension, term string, limit, offset int) (*models.SearchResponse, int, error) {
	response := &models.SearchResponse{}

	logData := log.Data{"term": term}

	log.Info("searching index", logData)

	path := api.url + "/" + instanceID + "_" + dimension + "/_search"

	logData["path"] = path

	body := buildSearchQuery(term, limit, offset)

	bytes, err := json.Marshal(body)
	if err != nil {
		log.Error(err, logData)
		return nil, 0, err
	}

	logData["request_body"] = string(bytes)

	log.Info("path is", logData)

	responseBody, status, err := api.CallElastic(ctx, path, "GET", bytes)
	logData["status"] = status
	if err != nil {
		log.ErrorC("failed to call elasticsearch", err, logData)
		return nil, status, err
	}

	logData["response_body"] = string(responseBody)

	if err = json.Unmarshal(responseBody, response); err != nil {
		log.ErrorC("unable to parse json body", err, logData)
		return nil, status, errors.New("Failed to parse json body")
	}

	log.Info("search results", log.Data{"results": response})

	return response, status, nil
}

// CallElastic builds a request to elastic search based on the method, path and payload
func (api *API) CallElastic(ctx context.Context, path, method string, payload interface{}) ([]byte, int, error) {
	logData := log.Data{"url": path, "method": method}

	URL, err := url.Parse(path)
	if err != nil {
		log.ErrorC("failed to create url for elastic call", err, logData)
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
		log.ErrorC("failed to create request for call to elastic", err, logData)
		return nil, 0, err
	}

	resp, err := api.client.Do(ctx, req)
	if err != nil {
		log.ErrorC("failed to call elastic", err, logData)
		return nil, 0, err
	}
	defer resp.Body.Close()

	logData["http_code"] = resp.StatusCode

	jsonBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.ErrorC("failed to read response body from call to elastic", err, logData)
		return nil, resp.StatusCode, err
	}
	logData["json_body"] = string(jsonBody)
	logData["status_code"] = resp.StatusCode

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= 300 {
		log.ErrorC("failed", ErrorUnexpectedStatusCode, logData)
		return nil, resp.StatusCode, ErrorUnexpectedStatusCode
	}

	return jsonBody, resp.StatusCode, nil
}

type Body struct {
	From      int        `json:"from"`
	Size      int        `json:"size"`
	Highlight *Highlight `json:"highlight,omitempty"`
	Query     Query      `json:"query"`
	Sort      []Scores   `json:"sort"`
}

type Highlight struct {
	PreTags  []string `json:"pre_tags,omitempty"`
	PostTags []string `json:"post_tags,omitempty"`
	Fields   []string `json:"fields,omitempty"` // guessing type for now
	Order    string   `json:"score,omitempty"`
}

type Query struct {
	Bool Bool `json:"bool"`
}

type Bool struct {
	Must   []Match `json:"must,omitempty"`
	Should []Match `json:"should,omitempty"`
}

type Match struct {
	Match map[string]string `json:"match,omitempty"`
}

type Scores struct {
	Score Score `json:"_score"`
}

type Score struct {
	Order string `json:"order"`
}

func buildSearchQuery(term string, limit, offset int) *Body {

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
