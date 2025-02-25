// configuration file for worker pools
package config

import (
	"github.com/algolia/algoliasearch-client-go/v4/algolia/search"
)

type Config struct {
	NumWorkers 	int32
	QueueName 	string
	AlgoliaClient *search.APIClient
}

func NewConfig(numWorkers int32, queueName string, algoliaClient *search.APIClient) *Config {
	return &Config{
		NumWorkers: numWorkers,
		QueueName: queueName,
		AlgoliaClient: algoliaClient,
	}
}