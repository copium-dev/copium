// configuration file for worker pools
package config

import (
	"github.com/algolia/algoliasearch-client-go/v4/algolia/search"
)

type Config struct {
	NumWorkers 	int32
	AlgoliaClient *search.APIClient
}

func NewConfig(numWorkers int32, algoliaClient *search.APIClient) *Config {
	return &Config{
		NumWorkers: numWorkers,
		AlgoliaClient: algoliaClient,
	}
}