package dto

import "fmt"

// Priority values accepted for a send request.
const (
	PriorityNormal  = "normal"
	PriorityExpress = "express"
)

// SendSMSRequest is the body for POST /sms.
type SendSMSRequest struct {
	CustomerID int64  `json:"customer_id"`
	To         string `json:"to"`
	Text       string `json:"text"`
	Priority   string `json:"priority"`
}

func (r *SendSMSRequest) Validate() error {
	if r.CustomerID == 0 {
		return fmt.Errorf("customer_id is required")
	}
	if r.To == "" {
		return fmt.Errorf("to is required")
	}
	if r.Text == "" {
		return fmt.Errorf("text is required")
	}
	if len(r.Text) > 160 {
		return fmt.Errorf("text exceeds single-page SMS length (160 chars)")
	}
	switch r.Priority {
	case "", PriorityNormal:
		r.Priority = PriorityNormal
	case PriorityExpress:
	default:
		return fmt.Errorf("priority must be %q or %q, got %q", PriorityNormal, PriorityExpress, r.Priority)
	}
	return nil
}
