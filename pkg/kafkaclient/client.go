package kafka

import (
	"fmt"

	"github.com/segmentio/kafka-go"

	"github.com/calgorr/sms-gateway/config"
)

func NewProducer(cfg config.Kafka) (*kafka.Writer, error) {
	if len(cfg.Brokers) == 0 {
		return nil, fmt.Errorf("at least one broker is required")
	}

	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Balancer:     &kafka.Hash{},
		RequiredAcks: kafka.RequireAll,
		Async:        false,
	}

	return writer, nil
}

func NewConsumer(cfg config.Kafka, topic string) (*kafka.Reader, error) {
	if len(cfg.Brokers) == 0 {
		return nil, fmt.Errorf("at least one broker is required")
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cfg.Brokers,
		GroupID:  cfg.GroupID,
		Topic:    topic,
		MaxBytes: 10e6,
	})

	return reader, nil
}
