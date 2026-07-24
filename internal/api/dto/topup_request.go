package dto

import (
	"fmt"

	"github.com/calgorr/sms-gateway/config"
)

type TopUpRequest struct {
	CustomerID int64 `json:"customer_id"`
	Amount     int64 `json:"amount"`
}

func (r *TopUpRequest) Validate() error {
	if r.CustomerID == 0 {
		return fmt.Errorf("customer_id is required")
	}
	if r.Amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}

	if r.Amount%config.C.Opts.CostPerSms != 0 {
		return fmt.Errorf("amount must be a multiple of cost per SMS (%d)", config.C.Opts.CostPerSms)
	}
	return nil
}
