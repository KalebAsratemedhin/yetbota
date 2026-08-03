package follower

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	domainFollower "github.com/beka-birhanu/yetbota/content-service/internal/domain/follower"
	repoGraphRead "github.com/beka-birhanu/yetbota/content-service/internal/services/repository/graphread"
)

type repository struct {
	graph repoGraphRead.FollowRepository
	db    *sql.DB
}

type Config struct {
	Graph repoGraphRead.FollowRepository `validate:"required"`
	DB    *sql.DB                        `validate:"required"`
}

func (c *Config) Validate() error {
	return validator.Validate.Struct(c)
}

func NewRepo(c *Config) (domainFollower.Repository, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return &repository{graph: c.Graph, db: c.DB}, nil
}

func (r *repository) FollowerTree(ctx context.Context, authorID string, maxDepth int) ([]domainFollower.UserWithDepth, error) {
	return r.graph.FollowerTree(ctx, authorID, maxDepth)
}

// CountFollowers reads the pre-maintained followers counter from Postgres — O(1) index lookup.
func (r *repository) CountFollowers(ctx context.Context, authorID string) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx, `SELECT followers FROM users WHERE id = $1`, authorID).Scan(&count)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "something went wrong",
			ServiceMessage:    fmt.Sprintf("follower repo: count followers failed: %v", err),
		}
	}
	return count, nil
}

// ListCelebrityFollows returns IDs of accounts userID follows whose follower count exceeds threshold.
func (r *repository) ListCelebrityFollows(ctx context.Context, userID string, threshold int64) ([]string, error) {
	followedIDs, err := r.graph.ListFollowedIDs(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(followedIDs) == 0 {
		return nil, nil
	}

	placeholders := make([]string, len(followedIDs))
	args := make([]any, len(followedIDs)+1)
	args[0] = threshold
	for i, id := range followedIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args[i+1] = id
	}
	query := fmt.Sprintf(
		`SELECT id FROM users WHERE followers > $1 AND id IN (%s)`,
		strings.Join(placeholders, ","),
	)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "something went wrong",
			ServiceMessage:    fmt.Sprintf("follower repo: filter celebrities failed: %v", err),
		}
	}
	defer func() { _ = rows.Close() }()

	var celebIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, &toddlerr.Error{
				PublicStatusCode:  status.ServerError,
				ServiceStatusCode: status.ServerError,
				PublicMessage:     "something went wrong",
				ServiceMessage:    fmt.Sprintf("follower repo: scan celebrity id failed: %v", err),
			}
		}
		celebIDs = append(celebIDs, id)
	}
	return celebIDs, rows.Err()
}

func (r *repository) IsFollowing(ctx context.Context, followerID string, userIDs []string) (map[string]bool, error) {
	return r.graph.IsFollowing(ctx, followerID, userIDs)
}

func (r *repository) FollowersOf(ctx context.Context, userIDs []string) (map[string][]string, error) {
	return r.graph.FollowersOf(ctx, userIDs)
}
