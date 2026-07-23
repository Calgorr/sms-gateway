package message

import (
	"time"

	"github.com/gocql/gocql"
)

type Priority string

const (
	PriorityNormal  Priority = "normal"
	PriorityExpress Priority = "express"
)

type Status string

const (
	StatusPending Status = "pending"
	StatusQueued  Status = "queued"
	StatusSent    Status = "sent"
	StatusFailed  Status = "failed"
)

type Message struct {
	ID           gocql.UUID `cql:"id"`
	CustomerID   int64      `cql:"customer_id"`
	Text         string     `cql:"text"`
	Priority     Priority   `cql:"priority"`
	Status       Status     `cql:"status"`
	Attempts     int32      `cql:"attempts"`
	CreatedAt    time.Time  `cql:"created_at"`
	SentAt       time.Time  `cql:"sent_at"`
	FailedReason string     `cql:"failed_reason"`
}

type MessageReport struct {
	CustomerID int64
	From       time.Time
	To         time.Time

	TotalMessages int
	TotalCost     int64
	SuccessCount  int
	FailedCount   int
	PendingCount  int

	Messages []Message
}
