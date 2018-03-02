dp-search-api
==================

A ONS API used to search information against datasets which are published.

Requirements
-----------------
In order to run the service locally you will need the following:
- [Go](https://golang.org/doc/install)
- [Git](https://git-scm.com/downloads)
- [ElasticSearch](https://www.elastic.co/guide/en/elasticsearch/reference/5.4/index.html)

### Getting started

* Clone the repo `go get github.com/ONSdigital/dp-search-api`
* Run elasticsearch
* Run the application `make debug`

### Healthcheck

The endpoint `/healthcheck` checks the connection to elasticsearch and returns
one of:

- success (200, JSON "status": "OK")
- failure (500, JSON "status": "error").

### Configuration

| Environment variable       | Default                              | Description
| -------------------------- | -------------------------------------| -----------
| BIND_ADDR                  | :23100                               | The host and port to bind to
| DATASET_API_URL            | http://localhost:22000               | The host name for the dataset API
| DATASET_API_SECRET_KEY     | FD0108EA-825D-411C-9B1D-41EF7727F465 | The dataset APi secret key used for authentication
| ELASTIC_SEARCH_URL         | http://localhost:9200                | The host name for elasticsearch
| GRACEFUL_SHUTDOWN_TIMEOUT  | 5s                                   | The graceful shutdown timeout
| HEALTHCHECK_INTERVAL       | 1m                                   | The time between calling the healthcheck endpoint for check subsystems
| HEALTHCHECK_TIMEOUT        | 2s                                   | The timeout that the healthcheck allows for checked subsystems
| HIERARCHY_BUILT_TOPIC      | hierarchy-built                      | The kafka topic to write messages to
| KAFKA_ADDR                 | localhost:9092                       | The list of kafka hosts
| KAFKA_MAX_BYTES            | 2000000                              | The maximum permitted size of a message. Should be set equal to or smaller than the broker's `message.max.bytes`
| MAX_SEARCH_RESULTS_OFFSET  | 1000                                 | The maximum offset for the number of results returned by search query
| REQUEST_MAX_RETRIES        | 3                                    | The maximum number of attempts for a single http request due to external service failure
| SEARCH_API_URL             | http://localhost:23100               | The host name for this service, search API
| SECRET_KEY                 | SD0108EA-825D-411C-45J3-41EF7727F123 | A secret key used for authentication


### Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

### License

Copyright © 2016-2017, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details
