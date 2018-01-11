package main

import (
	"os"

	"github.com/ONSdigital/dp-search-api/config"
	"github.com/ONSdigital/go-ns/log"
)

func main() {
	cfg, err := config.Get()
	if err != nil {
		log.Error(err, nil)
		os.Exit(1)
	}

	log.Namespace = "dp-search-builder"

	log.Info("config on startup", log.Data{"config": cfg})
}
