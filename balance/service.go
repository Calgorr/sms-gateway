package balance

import "context"

type CheckBalanceService interface {
	CheckAndDeduct(ctx context.Context, customerID string) (int64, error)
	AddOrSetBalance(ctx context.Context, customerID string, amount int64) error
	SetBalance(ctx context.Context, customerID string, amount int64) error
}
