// configuration file for worker pools
package config

import (
	"cloud.google.com/go/bigquery"
)

type Config struct {
	NumWorkers 	int32
	BigQueryClient *bigquery.Client
}

func NewConfig(numWorkers int32, algoliaClient *bigquery.Client) *Config {
	return &Config{
		NumWorkers: numWorkers,
		BigQueryClient: algoliaClient,
	}
}