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
	StatusQueued Status = "queued"
	StatusSent   Status = "sent"
	StatusFailed Status = "failed"
)

type Message struct {
	CustomerID int64      `cql:"customer_id"`
	ID         gocql.UUID `cql:"id"`
	To         string     `cql:"to_number"`
	Text       string     `cql:"text"`
	Priority   Priority   `cql:"priority"`
	Status     Status     `cql:"status"`
	CreatedAt  time.Time  `cql:"created_at"`
	SentAt     time.Time  `cql:"sent_at"`
}

type MessageReport struct {
	CustomerID int64
	From       time.Time
	To         time.Time

	TotalMessages int
	SuccessCount  int
	FailedCount   int
	PendingCount  int

	Messages []Message
}
