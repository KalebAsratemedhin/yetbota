package feed

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	"github.com/beka-birhanu/yetbota/content-service/internal/domain/feed"
)

type seenRepo struct {
	db *sql.DB
}

type SeenConfig struct {
	DB *sql.DB `validate:"required"`
}

func (c *SeenConfig) Validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

func NewSeenRepo(cfg *SeenConfig) (feed.SeenRepository, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &seenRepo{db: cfg.DB}, nil
}

func (r *seenRepo) AddBulk(ctx context.Context, userID string, postIDs []string) error {
	if len(postIDs) == 0 {
		return nil
	}

	placeholders := make([]string, len(postIDs))
	args := make([]any, 0, len(postIDs)*2)
	for i, id := range postIDs {
		placeholders[i] = fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2)
		args = append(args, userID, id)
	}

	query := fmt.Sprintf(
		"INSERT INTO user_seen_posts (user_id, post_id) VALUES %s ON CONFLICT DO NOTHING",
		strings.Join(placeholders, ", "),
	)

	if _, err := r.db.ExecContext(ctx, query, args...); err != nil {
		return &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "Something went wrong",
			ServiceMessage:    fmt.Sprintf("failed to mark posts seen: %v", err),
		}
	}
	return nil
}
