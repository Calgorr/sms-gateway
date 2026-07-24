package message

import (
	"context"
	"fmt"
	"time"

	"github.com/gocql/gocql"
)

type MessageRepository interface {
	Upsert(ctx context.Context, msg *Message) error
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

func (r *repository) Report(
	ctx context.Context,
	customerID int64,
	from, to time.Time,
) (*MessageReport, error) {

	query := `
	SELECT
		customer_id,
		id,
		to_number,
		text,
		priority,
		status,
		created_at,
		sent_at
	FROM messages
	WHERE customer_id = ?
	  AND sent_at >= ?
	  AND sent_at <= ?
	ALLOW FILTERING
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
		&m.CustomerID,
		&m.ID,
		&m.To,
		&m.Text,
		&m.Priority,
		&m.Status,
		&m.CreatedAt,
		&m.SentAt,
	) {
		report.Messages = append(report.Messages, m)

		report.TotalMessages++

		switch m.Status {
		case StatusSent:
			report.SuccessCount++

		case StatusFailed:
			report.FailedCount++

		case StatusQueued:
			report.PendingCount++
		}
	}

	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("generate report: %w", err)
	}

	return report, nil
}

// Upsert inserts a new message or updates an existing one if the ID already exists.
func (r *repository) Upsert(ctx context.Context, msg *Message) error {
	query := `
	UPDATE messages SET
		to_number = ?,
		text = ?,
		priority = ?,
		status = ?,
		created_at = ?,
		sent_at = ?
	WHERE customer_id = ? AND id = ?
	`

	if err := r.session.Query(
		query,
		msg.To,
		msg.Text,
		string(msg.Priority),
		string(msg.Status),
		msg.CreatedAt,
		msg.SentAt,
		msg.CustomerID,
		msg.ID,
	).WithContext(ctx).Exec(); err != nil {
		return fmt.Errorf("upsert message: %w", err)
	}

	return nil
}
