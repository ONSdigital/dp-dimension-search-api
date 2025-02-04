package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/ONSdigital/dp-dimension-search-api/config"
	"github.com/smartystreets/goconvey/convey"
)

func TestSpec(t *testing.T) {
	convey.Convey("Given an environment with no environment variables set", t, func() {
		os.Clearenv()
		cfg, err := config.Get()

		convey.Convey("When the config values are retrieved", func() {
			convey.Convey("There should be no error returned", func() {
				convey.So(err, convey.ShouldBeNil)
			})

			convey.Convey("The values should be set to the expected defaults", func() {
				convey.So(cfg.AuthAPIURL, convey.ShouldEqual, "http://localhost:8082")
				convey.So(cfg.AwsRegion, convey.ShouldEqual, "eu-west-1")
				convey.So(cfg.AwsService, convey.ShouldEqual, "es")
				convey.So(cfg.BindAddr, convey.ShouldEqual, ":23100")
				convey.So(cfg.Brokers, convey.ShouldResemble, []string{"localhost:9092", "localhost:9093", "localhost:9094"})
				convey.So(cfg.DatasetAPIURL, convey.ShouldEqual, "http://localhost:22000")
				convey.So(cfg.ElasticSearchAPIURL, convey.ShouldEqual, "http://localhost:10200")
				convey.So(cfg.GracefulShutdownTimeout, convey.ShouldEqual, 5*time.Second)
				convey.So(cfg.HealthCheckInterval, convey.ShouldEqual, 30*time.Second)
				convey.So(cfg.HealthCheckCriticalTimeout, convey.ShouldEqual, 90*time.Second)
				convey.So(cfg.HierarchyBuiltTopic, convey.ShouldEqual, "hierarchy-built")
				convey.So(cfg.KafkaMaxBytes, convey.ShouldEqual, 2000000)
				convey.So(cfg.MaxRetries, convey.ShouldEqual, 3)
				convey.So(cfg.KafkaVersion, convey.ShouldEqual, "1.0.2")
				convey.So(cfg.KafkaSecProtocol, convey.ShouldEqual, "")
				convey.So(cfg.MaxSearchResultsOffset, convey.ShouldEqual, 1000)
				convey.So(cfg.SearchAPIURL, convey.ShouldEqual, "http://localhost:23100")
				convey.So(cfg.ServiceAuthToken, convey.ShouldEqual, "a507f722-f25a-4889-9653-23a2655b925c")
				convey.So(cfg.EnableURLRewriting, convey.ShouldEqual, false)
			})
		})
	})
}
