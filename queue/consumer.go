package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/gocql/gocql"
	"github.com/segmentio/kafka-go"

	"github.com/calgorr/sms-gateway/ledger"
	"github.com/calgorr/sms-gateway/message"
	"github.com/calgorr/sms-gateway/sms"
)

type Consumer struct {
	reader      *kafka.Reader
	ledgerRepo  ledger.LedgerRepository
	messageRepo message.MessageRepository
	workers     int
}

func NewConsumer(
	reader *kafka.Reader,
	ledgerRepo ledger.LedgerRepository,
	messageRepo message.MessageRepository,
	workers int,
) *Consumer {
	return &Consumer{
		reader:      reader,
		ledgerRepo:  ledgerRepo,
		messageRepo: messageRepo,
		workers:     workers,
	}
}

func (c *Consumer) Start(ctx context.Context) error {
	defer c.reader.Close()

	jobs := make(chan kafka.Message, c.workers*2)

	var wg sync.WaitGroup

	// Start workers.
	for i := 0; i < c.workers; i++ {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return

				case msg, ok := <-jobs:
					if !ok {
						return
					}

					if err := c.process(ctx, msg); err != nil {
						log.Printf("worker %d: %v", id, err)
						continue
					}

					if err := c.reader.CommitMessages(ctx, msg); err != nil {
						log.Printf("worker %d: commit failed: %v", id, err)
					}
				}
			}
		}(i)
	}

	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				break
			}
			return fmt.Errorf("fetch message: %w", err)
		}

		select {
		case <-ctx.Done():
			break

		case jobs <- msg:
		}
	}

	close(jobs)
	wg.Wait()

	return nil
}

func (c *Consumer) process(ctx context.Context, msg kafka.Message) error {
	var event sms.SMSEvent

	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	messageID, err := gocql.ParseUUID(event.MessageID)
	if err != nil {
		log.Printf("invalid message id %s: %v", event.MessageID, err)
		return err
	}

	if err := c.messageRepo.Upsert(ctx, &message.Message{
		ID:         messageID,
		CustomerID: event.CustomerID,
		To:         event.To,
		Text:       event.Text,
		Priority:   message.Priority(event.Priority),
		Status:     message.StatusQueued,
		CreatedAt:  event.CreatedAt,
	}); err != nil {
		return fmt.Errorf("create message: %w", err)
	}

	if err := c.ledgerRepo.InsertDebit(ctx, event.CustomerID, 1, event.MessageID); err != nil {
		return fmt.Errorf("create ledger: %w", err)
	}

	return nil
}
