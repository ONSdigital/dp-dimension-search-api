dp-dimension-search-api
==================

A ONS API used to search information against datasets which are published.

Requirements
-----------------
In order to run the service locally you will need the following:
- [Go](https://golang.org/doc/install)
- [Git](https://git-scm.com/downloads)
- [ElasticSearch](https://www.elastic.co/guide/en/elasticsearch/reference/5.4/index.html)

### Note
The only breaking change from verion 5.x to 6.x of elasticsearch is highlighting will
not work correctly but the api will stil be able to send back responses.

### Getting started

* Clone the repo `go get github.com/ONSdigital/dp-dimension-search-api`
* Run elasticsearch
* Run the application `make debug`

### Healthcheck

The endpoint `/health` checks all backing services, e.g. elasticsearch, dataset API:

- success (200)
- warning (429) - either application is starting up or a connection to a backend service 
    has been lost in the last `HEALTHCHECK_CRITICAL_TIMEOUT` value, (default set to 90 seconds, 
    see table below)but there is still time to recover.
- failure (500)


### Manually Creating and Deleting Indexes

CREATE: `curl -X PUT <HOSTNAME>/dimension-search/instances/<instanceID>/dimensions/<dimensionName> -H <AUTH HEADER>`
DELETE: `curl -X DELETE <HOSTNAME>/dimension-search/instances/<instanceID>/dimensions/<dimensionName> -H <AUTH HEADER>`

The `<AUTH HEADER>` must be either a valid `X-FLorence-Token` or a valid `Authorization` header.

### Kafka scripts

Scripts for updating and debugging Kafka can be found [here](https://github.com/ONSdigital/dp-data-tools)(dp-data-tools)


### Configuration

| Environment variable         | Default                              | Description
| ---------------------------- | -------------------------------------| -----------
| BIND_ADDR                    | :23100                               | The host and port to bind to
| DATASET_API_URL              | http://localhost:22000               | The host name and port for the dataset API
| ELASTIC_SEARCH_URL           | http://localhost:10200               | The host name and port for elasticsearch
| ENABLE_PRIVATE_ENDPOINTS     | false                                | Set true ("1","t","true") when private endpoints should be accessible
| GRACEFUL_SHUTDOWN_TIMEOUT    | 5s                                   | The graceful shutdown timeout
| HEALTHCHECK_INTERVAL         | 30s                                  | The time between calling the health check endpoint for check subsystems
| HEALTHCHECK_CRITICAL_TIMEOUT | 90s                                  | The timeout that the health check allows for checked subsystems
| HIERARCHY_BUILT_TOPIC        | hierarchy-built                      | The kafka topic to write messages to
| KAFKA_ADDR                   | localhost:9092                       | The list of kafka hosts
| KAFKA_MAX_BYTES              | 2000000                              | The maximum permitted size of a message. Should be set equal to or smaller than the broker's `message.max.bytes`
| KAFKA_VERSION                | "1.0.2"                              | The kafka version that this service expects to connect to
| MAX_SEARCH_RESULTS_OFFSET    | 1000                                 | The maximum offset for the number of results returned by search query
| REQUEST_MAX_RETRIES          | 3                                    | The maximum number of attempts for a single http request due to external service failure
| SEARCH_API_URL               | http://localhost:23100               | The host name and port for this service, dimension search API
| SERVICE_AUTH_TOKEN           | SD0108EA-825D-411C-45J3-41EF7727F123 | The token used to identify this service when authenticating
| SIGN_ELASTICSEARCH_REQUESTS  | false                                | Boolean flag to identify whether elasticsearch requests via elastic API need to be signed if elasticsearch cluster is running in aws
| ZEBEDEE_URL                  | http://localhost:8082                | The URL to zebedee, used to authenticate requests


### Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

### License

Copyright © 2016-2017, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details
