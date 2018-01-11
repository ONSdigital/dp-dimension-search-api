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
				So(cfg.BindAddr, ShouldEqual, ":23100")
				So(cfg.ElasticSearchAPIURL, ShouldEqual, "http://localhost:9200")
				So(cfg.GracefulShutdownTimeout, ShouldEqual, 5*time.Second)
				So(cfg.HealthcheckInterval, ShouldEqual, 60*time.Second)
				So(cfg.HealthcheckTimeout, ShouldEqual, 2*time.Second)
				So(cfg.MaxRetries, ShouldEqual, 3)
			})
		})
	})
}
