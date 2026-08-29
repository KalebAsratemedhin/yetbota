package surreal

import (
	"context"
	"fmt"

	surrealdb "github.com/surrealdb/surrealdb.go"

	"github.com/yetbota/identity-service/drivers/config"
)

// Connect opens a SurrealDB client (shared graph store for follows).
// Sign in as root first, then select namespace/database.
func Connect(ctx context.Context, cfg config.SurrealConfig) (*surrealdb.DB, error) {
	db, err := surrealdb.FromEndpointURLString(ctx, cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("surreal connect: %w", err)
	}
	if _, err := db.SignIn(ctx, surrealdb.Auth{
		Username: cfg.Username,
		Password: cfg.Password,
	}); err != nil {
		_ = db.Close(ctx)
		return nil, fmt.Errorf("surreal signin: %w", err)
	}
	if err := db.Use(ctx, cfg.Namespace, cfg.Database); err != nil {
		_ = db.Close(ctx)
		return nil, fmt.Errorf("surreal use ns/db: %w", err)
	}
	return db, nil
}

// EnsureSchema creates graph tables if missing.
// Cloud Surreal does not auto-create tables on SELECT; without this, empty
// follower lists fail with "table 'follows' does not exist".
func EnsureSchema(ctx context.Context, db *surrealdb.DB) error {
	res, err := surrealdb.Query[any](ctx, db, `
		DEFINE TABLE IF NOT EXISTS user TYPE NORMAL SCHEMALESS;
		DEFINE TABLE IF NOT EXISTS follows TYPE RELATION IN user OUT user SCHEMALESS;
		DEFINE FIELD IF NOT EXISTS created_at ON follows TYPE option<datetime>;
	`, nil)
	if err != nil {
		return fmt.Errorf("surreal ensure schema: %w", err)
	}
	if res == nil {
		return nil
	}
	for _, qr := range *res {
		if qr.Error != nil {
			return fmt.Errorf("surreal ensure schema: %w", qr.Error)
		}
		if qr.Status != "" && qr.Status != "OK" {
			return fmt.Errorf("surreal ensure schema status %s", qr.Status)
		}
	}
	return nil
}
