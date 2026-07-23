package kafka

import (
	"fmt"

	"github.com/IBM/sarama"

	"github.com/calgorr/sms-gateway/config"
)

func NewProducer(cfg config.Config) (sarama.SyncProducer, error) {
	if len(cfg.Kafka.Brokers) == 0 {
		return nil, fmt.Errorf("at least one broker is required")
	}

	scfg := sarama.NewConfig()

	scfg.ClientID = cfg.Kafka.ClientID
	scfg.Version = sarama.V3_6_0_0

	scfg.Producer.RequiredAcks = sarama.WaitForAll
	scfg.Producer.Return.Successes = true
	scfg.Producer.Return.Errors = true

	return sarama.NewSyncProducer(cfg.Kafka.Brokers, scfg)
}

func NewConsumer(cfg config.Config) (sarama.Consumer, error) {
	if len(cfg.Kafka.Brokers) == 0 {
		return nil, fmt.Errorf("at least one broker is required")
	}

	scfg := sarama.NewConfig()
	scfg.ClientID = cfg.Kafka.ClientID
	scfg.Version = sarama.V3_6_0_0

	return sarama.NewConsumer(cfg.Kafka.Brokers, scfg)
}
