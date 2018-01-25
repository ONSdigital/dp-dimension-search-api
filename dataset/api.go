package dataset

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/ONSdigital/dp-dataset-api/models"
	datasetclient "github.com/ONSdigital/go-ns/clients/dataset"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/rchttp"
)

// API aggregates a client and URL and other common data for accessing the dataset API
type API struct {
	client    *rchttp.Client
	url       string
	authToken string
}

// A list of errors that the dataset package could return
var (
	ErrorUnexpectedStatusCode = errors.New("unexpected status code from api")
	ErrorVersionNotFound      = errors.New("Version not found")
)

// NewDatasetAPI returns common data for accessing the dataset API
func NewDatasetAPI(client *rchttp.Client, url string) *API {
	return &API{
		client: client,
		url:    url,
	}
}

// GetVersion queries the Dataset API to get a version
func (api *API) GetVersion(ctx context.Context, datasetID, edition, version, authToken string) (versionDoc *models.Version, err error) {
	path := api.url + "/datasets/" + datasetID + "/editions/" + edition + "/versions/" + version
	logData := log.Data{"func": "GetVersion", "url": path, "dataset_id": datasetID, "edition": edition, "version": version}

	jsonResult, httpCode, err := api.get(ctx, path, authToken, nil)
	logData["http_code"] = httpCode
	logData["json_result"] = jsonResult
	if err != nil {
		log.ErrorC("api get", err, logData)
		return nil, ErrorVersionNotFound
	}

	versionDoc = &models.Version{}
	if err = json.Unmarshal(jsonResult, versionDoc); err != nil {
		log.ErrorC("unmarshal", err, logData)
		return
	}

	return
}

func (api *API) get(ctx context.Context, path string, authToken string, vars url.Values) ([]byte, int, error) {
	return api.callDatasetAPI(ctx, "GET", path, authToken, vars)
}

// callDatasetAPI contacts the Dataset API returns the json body (action = PUT, GET, POST, ...)
func (api *API) callDatasetAPI(ctx context.Context, method, path string, authToken string, payload interface{}) ([]byte, int, error) {
	logData := log.Data{"url": path, "method": method}

	URL, err := url.Parse(path)
	if err != nil {
		log.ErrorC("failed to create url for dataset api call", err, logData)
		return nil, 0, err
	}
	path = URL.String()
	logData["url"] = path

	var req *http.Request

	if payload != nil && method != "GET" {
		req, err = http.NewRequest(method, path, bytes.NewReader(payload.([]byte)))
		req.Header.Add("Content-type", "application/json")
		logData["payload"] = string(payload.([]byte))
	} else {
		req, err = http.NewRequest(method, path, nil)

		if payload != nil && method == "GET" {
			req.URL.RawQuery = payload.(url.Values).Encode()
			logData["payload"] = payload.(url.Values)
		}
	}
	// check req, above, didn't error
	if err != nil {
		log.ErrorC("failed to create request for dataset api", err, logData)
		return nil, 0, err
	}

	if authToken != "" {
		req.Header.Set("Internal-token", authToken)
	}

	resp, err := api.client.Do(ctx, req)
	if err != nil {
		log.ErrorC("failed to action dataset api", err, logData)
		return nil, 0, err
	}
	defer resp.Body.Close()

	logData["http_code"] = resp.StatusCode
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= 300 {
		return nil, resp.StatusCode, ErrorUnexpectedStatusCode
	}

	jsonBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.ErrorC("failed to read body from dataset api", err, logData)
		return nil, resp.StatusCode, err
	}

	return jsonBody, resp.StatusCode, nil
}

// GetHealthCheckClient returns a healthcheck-compatible client
func (api *API) GetHealthCheckClient() *datasetclient.Client {
	return datasetclient.New(api.url)
}
