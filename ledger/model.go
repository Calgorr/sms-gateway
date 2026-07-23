package ledger

import (
	"database/sql"
	"time"
)

type Entry struct {
	ID         int64          `db:"id"`
	CustomerID string         `db:"customer_id"`
	Amount     int64          `db:"amount"`
	MessageID  sql.NullString `db:"message_id"`
	CreatedAt  time.Time      `db:"created_at"`
}
