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
