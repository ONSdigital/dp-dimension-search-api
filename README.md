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
| ELASTIC_SEARCH_URL         | http://localhost:9200                | The host name for elasticsearch
| GRACEFUL_SHUTDOWN_TIMEOUT  | 5s                                   | The graceful shutdown timeout
| HEALTHCHECK_INTERVAL       | 1m                                   | The time between calling the healthcheck endpoint for check subsystems
| HEALTHCHECK_TIMEOUT        | 2s                                   | The timeout that the healthcheck allows for checked subsystems
| REQUEST_MAX_RETRIES        | 3                                    | The maximum number of attempts for a single http request due to external service failure


### Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

### License

Copyright © 2016-2017, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details
