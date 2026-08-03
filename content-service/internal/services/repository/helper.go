package repository

import (
	"context"
	"database/sql"

	"github.com/aarondl/sqlboiler/v4/boil"
	toddlerr "github.com/beka-birhanu/toddler/error"
)

func BeginNewTx(ctx context.Context) (*sql.Tx, error) {
	tx, err := boil.BeginTx(ctx, nil)
	if err != nil {
		return nil, toddlerr.FromDBError(err, "transaction")
	}
	return tx, nil
}

func CommitTx(tx *sql.Tx) error {
	if err := tx.Commit(); err != nil {
		return toddlerr.FromDBError(err, "transaction")
	}
	return nil
}
