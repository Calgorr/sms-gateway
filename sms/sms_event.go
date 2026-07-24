package sms

import "time"

// SMSEvent is the payload published to Kafka by the API handler.
type SMSEvent struct {
	MessageID  string    `json:"message_id"`
	CustomerID int64     `json:"customer_id"`
	To         string    `json:"to"`
	Text       string    `json:"text"`
	Priority   string    `json:"priority"`
	CreatedAt  time.Time `json:"created_at"`
}
