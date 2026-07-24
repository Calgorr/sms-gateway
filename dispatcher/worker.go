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
	name     string
	reader   *kafka.Reader
	operator operator.Client
	messages message.MessageRepository
	workers  int
}

func NewWorker(name string, reader *kafka.Reader, op operator.Client, messages message.MessageRepository, workers int) *Worker {
	return &Worker{name: name, reader: reader, operator: op, messages: messages, workers: workers}
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
				log.Printf("%s: fetch error: %v", w.name, err)
				continue
			}
			msgChan <- m
		}
	}
}

func (w *Worker) processLoop(ctx context.Context, msgChan <-chan kafka.Message) {
	for m := range msgChan {
		if err := w.handle(ctx, m); err != nil {
			log.Printf("%s: handle failed, leaving uncommitted for retry: %v", w.name, err)
			continue // offset not committed -> redelivered on rebalance/restart
		}
		if err := w.reader.CommitMessages(ctx, m); err != nil {
			log.Printf("%s: commit failed: %v", w.name, err)
		}
	}
}

func (w *Worker) handle(ctx context.Context, m kafka.Message) error {
	var event sms.SMSEvent
	if err := json.Unmarshal(m.Value, &event); err != nil {
		log.Printf("%s: bad event payload, dropping: %v", w.name, err)
		return nil
	}

	sendErr := w.operator.Send(ctx, event.To, event.Text)

	status := message.StatusSent
	if sendErr != nil {
		status = message.StatusFailed
		log.Printf("%s: operator send failed for %s: %v", w.name, event.MessageID, sendErr)
	}
	messageID, err := gocql.ParseUUID(event.MessageID)
	if err != nil {
		log.Printf("%s: invalid message id %s: %v", w.name, event.MessageID, err)
		return err
	}

	msg := &message.Message{
		ID:         messageID,
		CustomerID: event.CustomerID,
		To:         event.To,
		Text:       event.Text,
		Priority:   message.Priority(event.Priority),
		Status:     status,
		CreatedAt:  event.CreatedAt,
		SentAt:     time.Now(),
	}
	if err := w.messages.Upsert(ctx, msg); err != nil {
		log.Printf("%s: failed to update message status for %s: %v", w.name, messageID, err)
		//TODO: consider retrying the update or sending to a dead-letter queue
	}

	return nil
}
