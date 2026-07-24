package dispatcher

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/gocql/gocql"
	"github.com/segmentio/kafka-go"

	"github.com/calgorr/sms-gateway/message"
	"github.com/calgorr/sms-gateway/operator"
	"github.com/calgorr/sms-gateway/sms"
)

type Worker struct {
	reader   *kafka.Reader
	operator operator.Client
	messages message.MessageRepository
	workers  int
}

func NewWorker(reader *kafka.Reader, op operator.Client, messages message.MessageRepository, workers int) *Worker {
	return &Worker{reader: reader, operator: op, messages: messages, workers: workers}
}

func (w *Worker) Run(ctx context.Context) error {
	msgChan := make(chan kafka.Message, w.workers*2)

	for i := 0; i < w.workers; i++ {
		go w.processLoop(ctx, msgChan)
	}

	for {
		select {
		case <-ctx.Done():
			close(msgChan)
			return ctx.Err()
		default:
			m, err := w.reader.FetchMessage(ctx)
			if err != nil {
				log.Printf("normal-worker: fetch error: %v", err)
				continue
			}
			msgChan <- m
		}
	}
}

func (w *Worker) processLoop(ctx context.Context, msgChan <-chan kafka.Message) {
	for m := range msgChan {
		if err := w.handle(ctx, m); err != nil {
			log.Printf("normal-worker: handle failed, leaving uncommitted for retry: %v", err)
			continue // offset not committed -> redelivered on rebalance/restart
		}
		if err := w.reader.CommitMessages(ctx, m); err != nil {
			log.Printf("normal-worker: commit failed: %v", err)
		}
	}
}

func (w *Worker) handle(ctx context.Context, m kafka.Message) error {
	var event sms.SMSEvent
	if err := json.Unmarshal(m.Value, &event); err != nil {
		log.Printf("normal-worker: bad event payload, dropping: %v", err)
		return nil
	}

	sendErr := w.operator.Send(ctx, event.To, event.Text)

	failReason := ""
	status := message.StatusSent
	if sendErr != nil {
		status = message.StatusFailed
		failReason = sendErr.Error()
	}
	messageID, err := gocql.ParseUUID(event.MessageID)
	if err != nil {
		log.Printf("invalid message id %s: %v", event.MessageID, err)
		return err
	}

	msg := &message.Message{
		ID:           messageID,
		CustomerID:   event.CustomerID,
		To:           event.To,
		Text:         event.Text,
		Priority:     message.Priority(event.Priority),
		Status:       status,
		CreatedAt:    event.CreatedAt,
		SentAt:       time.Now(),
		FailedReason: failReason,
	}
	if err := w.messages.Upsert(ctx, msg); err != nil {
		log.Printf("normal-worker: failed to update message status for %s: %v", messageID, err)
		//TODO: consider retrying the update or sending to a dead-letter queue
	}

	return nil
}
