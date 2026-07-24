package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/segmentio/kafka-go"

	"github.com/calgorr/sms-gateway/sms"
)

type KafkaProducer interface {
	SendSMS(ctx context.Context, topic string, message sms.SMSEvent) error
}

type smsProducer struct {
	writer *kafka.Writer
}

func NewKafkaProducer(writer *kafka.Writer) KafkaProducer {
	return &smsProducer{
		writer: writer,
	}
}

func (s *smsProducer) SendSMS(
	ctx context.Context,
	topic string,
	message sms.SMSEvent,
) error {
	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("marshal sms message: %w", err)
	}

	bucket := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(10) // Randomly assign a bucket number between 0 and 9 for partitioning evenly across partitions.

	err = s.writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Key:   []byte(fmt.Sprintf("%d-%d", message.CustomerID, bucket)),
		Value: payload,
	})
	if err != nil {
		return fmt.Errorf("publish sms message: %w", err)
	}

	return nil
}
