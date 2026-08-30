package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Separate from identity-service's goose_db_version to avoid version collisions
// on the shared Postgres database.
const versionTable = "goose_db_version_content"

func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("pgx pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return pool, nil
}

func Migrate(databaseURL, migrationsDir string) error {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return fmt.Errorf("open db for migrate: %w", err)
	}
	defer db.Close()

	if _, err := db.Exec(fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id SERIAL PRIMARY KEY,
			version_id BIGINT NOT NULL,
			is_applied BOOLEAN NOT NULL,
			tstamp TIMESTAMP DEFAULT NOW()
		)`, versionTable)); err != nil {
		return fmt.Errorf("ensure %s: %w", versionTable, err)
	}

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	for _, name := range files {
		version, err := parseGooseVersion(name)
		if err != nil {
			return err
		}
		var applied bool
		err = db.QueryRow(
			fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM %s WHERE version_id = $1 AND is_applied = true)`, versionTable),
			version,
		).Scan(&applied)
		if err != nil {
			return fmt.Errorf("check migration %s: %w", name, err)
		}
		if applied {
			continue
		}
		body, err := os.ReadFile(filepath.Join(migrationsDir, name))
		if err != nil {
			return fmt.Errorf("read %s: %w", name, err)
		}
		upSQL, err := extractGooseUp(string(body))
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		if _, err := tx.Exec(upSQL); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("apply %s: %w", name, err)
		}
		if _, err := tx.Exec(
			fmt.Sprintf(`INSERT INTO %s (version_id, is_applied) VALUES ($1, true)`, versionTable),
			version,
		); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("record %s: %w", name, err)
		}
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}

func parseGooseVersion(filename string) (int64, error) {
	base := strings.TrimSuffix(filename, ".sql")
	parts := strings.SplitN(base, "_", 2)
	var version int64
	if _, err := fmt.Sscanf(parts[0], "%d", &version); err != nil {
		return 0, fmt.Errorf("bad migration version in %q: %w", filename, err)
	}
	return version, nil
}

func extractGooseUp(content string) (string, error) {
	const up = "-- +goose Up"
	const down = "-- +goose Down"
	start := strings.Index(content, up)
	if start < 0 {
		return "", fmt.Errorf("missing %s", up)
	}
	rest := content[start+len(up):]
	if i := strings.Index(rest, down); i >= 0 {
		rest = rest[:i]
	}
	sql := strings.TrimSpace(rest)
	if sql == "" {
		return "", fmt.Errorf("empty Up section")
	}
	return sql, nil
}
