// configuration file for worker pools
package config

type Config struct {
	NumWorkers 	int32
	QueueName 	string
}

func NewConfig(numWorkers int32, queueName string) *Config {
	return &Config{
		NumWorkers: numWorkers,
		QueueName: queueName,
	}
}