package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/ONSdigital/dp-dimension-search-api/config"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSpec(t *testing.T) {
	Convey("Given an environment with no environment variables set", t, func() {
		os.Clearenv()
		cfg, err := config.Get()

		Convey("When the config values are retrieved", func() {

			Convey("There should be no error returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("The values should be set to the expected defaults", func() {
				So(cfg.AuthAPIURL, ShouldEqual, "http://localhost:8082")
				So(cfg.AwsRegion, ShouldEqual, "eu-west-1")
				So(cfg.AwsService, ShouldEqual, "es")
				So(cfg.AwsSDKSigner, ShouldEqual, false)
				So(cfg.BindAddr, ShouldEqual, ":23100")
				So(cfg.Brokers, ShouldResemble, []string{"localhost:9092"})
				So(cfg.DatasetAPIURL, ShouldEqual, "http://localhost:22000")
				So(cfg.ElasticSearchAPIURL, ShouldEqual, "http://localhost:10200")
				So(cfg.GracefulShutdownTimeout, ShouldEqual, 5*time.Second)
				So(cfg.HealthCheckInterval, ShouldEqual, 30*time.Second)
				So(cfg.HealthCheckCriticalTimeout, ShouldEqual, 90*time.Second)
				So(cfg.HierarchyBuiltTopic, ShouldEqual, "hierarchy-built")
				So(cfg.KafkaMaxBytes, ShouldEqual, 2000000)
				So(cfg.MaxRetries, ShouldEqual, 3)
				So(cfg.MaxSearchResultsOffset, ShouldEqual, 1000)
				So(cfg.SearchAPIURL, ShouldEqual, "http://localhost:23100")
				So(cfg.ServiceAuthToken, ShouldEqual, "a507f722-f25a-4889-9653-23a2655b925c")
			})
		})
	})
}
