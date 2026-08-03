package repository

import (
	"context"
	"database/sql"

	"github.com/aarondl/sqlboiler/v4/boil"
)

// BeginNewTx starts a new database transaction
func BeginNewTx(ctx context.Context) (*sql.Tx, error) {
	return boil.BeginTx(ctx, nil)
}
