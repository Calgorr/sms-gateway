package dto

import "time"

// ReportEntry represents a single SMS in the report.
type ReportEntry struct {
	MessageID string     `json:"message_id"`
	To        string     `json:"to"`
	Text      string     `json:"text"`
	Priority  string     `json:"priority"`
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	SentAt    *time.Time `json:"sent_at,omitempty"`
}

// ReportResponse is returned by GET /reports.
type ReportResponse struct {
	CustomerID string    `json:"customer_id"`
	From       time.Time `json:"from"`
	To         time.Time `json:"to"`

	TotalMessages int `json:"total_messages"`
	SuccessCount  int `json:"success_count"`
	FailedCount   int `json:"failed_count"`
	PendingCount  int `json:"pending_count"`

	Entries    []ReportEntry `json:"entries"`
	NextCursor string        `json:"next_cursor,omitempty"` // Cassandra paging
}
