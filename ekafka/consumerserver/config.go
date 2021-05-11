package consumerserver

import "github.com/gotomicro/ego-component/ekafka"

type config struct {
	Debug             bool   `json:"debug" toml:"debug"`
	ConsumerName      string `json:"consumerName" toml:"consumerName"`
	ConsumerGroupName string `json:"consumerGroupName" toml:"consumerGroupName"`
	ekafkaComponent   *ekafka.Component
}

// DefaultConfig returns a default config.
func DefaultConfig() *config {
	return &config{
		Debug:             true,
		ConsumerName:      "default",
		ConsumerGroupName: "default",
	}
}
