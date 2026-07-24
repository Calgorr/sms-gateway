package ledger

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type LedgerRepository interface {
	InsertDebit(ctx context.Context, customerID, amount int64, messageID string) error
	InsertTopup(ctx context.Context, customerID, amount int64) error
	FetchByCustomer(ctx context.Context, customerID int64, limit int) ([]Entry, error)
	SumByCustomer(ctx context.Context, customerID int64) (int64, error)
}

type ledgerRepository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) LedgerRepository {
	return &ledgerRepository{db: db}
}

// InsertDebit records a negative ledger entry tied to a specific message.
func (r *ledgerRepository) InsertDebit(ctx context.Context, customerID, amount int64, messageID string) error {
	if amount > 0 {
		return fmt.Errorf("debit amount must be negative or zero, got %d", amount)
	}

	return r.insert(ctx, customerID, amount, sql.NullString{
		String: messageID,
		Valid:  true,
	})
}

// InsertTopup records a positive ledger entry.
func (r *ledgerRepository) InsertTopup(ctx context.Context, customerID, amount int64) error {
	if amount <= 0 {
		return fmt.Errorf("topup amount must be positive, got %d", amount)
	}

	return r.insert(ctx, customerID, amount, sql.NullString{})
}

func (r *ledgerRepository) insert(
	ctx context.Context,
	customerID int64,
	amount int64,
	messageID sql.NullString,
) error {
	query := `
		INSERT INTO ledger (
			customer_id,
			amount,
			message_id,
			created_at
		)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		customerID,
		amount,
		messageID,
		time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("insert ledger entry: %w", err)
	}

	return nil
}

// FetchByCustomer returns the most recent entries.
// If limit <= 0, all entries are returned.
func (r *ledgerRepository) FetchByCustomer(
	ctx context.Context,
	customerID int64,
	limit int,
) ([]Entry, error) {
	query := `
		SELECT
			id,
			customer_id,
			amount,
			message_id,
			created_at
		FROM ledger
		WHERE customer_id = $1
		ORDER BY created_at DESC
	`

	var (
		entries []Entry
		err     error
	)

	if limit > 0 {
		query += " LIMIT $2"
		err = r.db.SelectContext(ctx, &entries, query, customerID, limit)
	} else {
		err = r.db.SelectContext(ctx, &entries, query, customerID)
	}

	if err != nil {
		return nil, fmt.Errorf("fetch ledger entries: %w", err)
	}

	return entries, nil
}

func (r *ledgerRepository) SumByCustomer(
	ctx context.Context,
	customerID int64,
) (int64, error) {
	query := `
		SELECT COALESCE(SUM(amount), 0)
		FROM ledger
		WHERE customer_id = $1
	`

	var total int64
	if err := r.db.GetContext(ctx, &total, query, customerID); err != nil {
		return 0, fmt.Errorf("sum ledger: %w", err)
	}

	return total, nil
}
