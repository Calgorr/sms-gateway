package message

import (
	"context"
	"fmt"
	"time"

	"github.com/gocql/gocql"
)

type MessageRepository interface {
	Insert(ctx context.Context, msg *Message) error
	Report(ctx context.Context, customerID int64, from, to time.Time) (*MessageReport, error)
}

type repository struct {
	session *gocql.Session
}

func NewRepository(session *gocql.Session) MessageRepository {
	return &repository{
		session: session,
	}
}

func (r *repository) Insert(ctx context.Context, msg *Message) error {
	query := `
	INSERT INTO messages (
		id,
		customer_id,
		request_id,
		to_number,
		text,
		cost,
		priority,
		status,
		operator_id,
		attempts,
		created_at,
		sent_at,
		delivered_at,
		failed_reason
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	if err := r.session.Query(
		query,
		msg.ID,
		msg.CustomerID,
		msg.Text,
		string(msg.Priority),
		string(msg.Status),
		msg.Attempts,
		msg.CreatedAt,
		msg.SentAt,
		msg.FailedReason,
	).WithContext(ctx).Exec(); err != nil {
		return fmt.Errorf("insert message: %w", err)
	}

	return nil
}

func (r *repository) Report(
	ctx context.Context,
	customerID int64,
	from, to time.Time,
) (*MessageReport, error) {

	query := `
	SELECT
		id,
		customer_id,
		request_id,
		to_number,
		text,
		cost,
		priority,
		status,
		operator_id,
		attempts,
		created_at,
		sent_at,
		delivered_at,
		failed_reason
	FROM messages
	WHERE customer_id = ?
	  AND sent_at >= ?
	  AND sent_at <= ?
	`

	iter := r.session.Query(
		query,
		customerID,
		from,
		to,
	).WithContext(ctx).Iter()

	report := &MessageReport{
		CustomerID: customerID,
		From:       from,
		To:         to,
	}

	var m Message

	for iter.Scan(
		&m.ID,
		&m.CustomerID,
		&m.Text,
		&m.Priority,
		&m.Status,
		&m.Attempts,
		&m.CreatedAt,
		&m.SentAt,
		&m.FailedReason,
	) {
		report.Messages = append(report.Messages, m)

		report.TotalMessages++
		report.TotalCost += 1

		switch m.Status {
		case StatusSent:
			report.SuccessCount++

		case StatusFailed:
			report.FailedCount++

		case StatusPending, StatusQueued:
			report.PendingCount++
		}
	}

	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("generate report: %w", err)
	}

	return report, nil
}
