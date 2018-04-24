package config_test

import (
	"testing"
	"time"

	"github.com/ONSdigital/dp-search-api/config"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSpec(t *testing.T) {
	Convey("Given an environment with no environment variables set", t, func() {
		cfg, err := config.Get()

		Convey("When the config values are retrieved", func() {

			Convey("There should be no error returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("The values should be set to the expected defaults", func() {
				So(cfg.AuthAPIURL, ShouldEqual, "http://localhost:8082")
				So(cfg.BindAddr, ShouldEqual, ":23100")
				So(cfg.Brokers, ShouldResemble, []string{"localhost:9092"})
				So(cfg.DatasetAPIURL, ShouldEqual, "http://localhost:22000")
				So(cfg.ElasticSearchAPIURL, ShouldEqual, "http://localhost:9200")
				So(cfg.GracefulShutdownTimeout, ShouldEqual, 5*time.Second)
				So(cfg.HealthCheckInterval, ShouldEqual, 60*time.Second)
				So(cfg.HealthCheckTimeout, ShouldEqual, 2*time.Second)
				So(cfg.HierarchyBuiltTopic, ShouldEqual, "hierarchy-built")
				So(cfg.KafkaMaxBytes, ShouldEqual, 2000000)
				So(cfg.MaxRetries, ShouldEqual, 3)
				So(cfg.MaxSearchResultsOffset, ShouldEqual, 1000)
				So(cfg.SearchAPIURL, ShouldEqual, "http://localhost:23100")
				So(cfg.ServiceAuthToken, ShouldEqual, "SD0108EA-825D-411C-45J3-41EF7727F123")
			})
		})
	})
}
