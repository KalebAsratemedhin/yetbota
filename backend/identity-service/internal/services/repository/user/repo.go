package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yetbota/identity-service/internal/db"
	domain "github.com/yetbota/identity-service/internal/domain/user"
)

type Repo struct {
	q *db.Queries
}

func New(pool *pgxpool.Pool) *Repo {
	return &Repo{q: db.New(pool)}
}

func (r *Repo) Create(ctx context.Context, username, passwordHash string) (domain.User, error) {
	id := uuid.New()
	row, err := r.q.CreateUser(ctx, db.CreateUserParams{
		ID:           pgUUID(id),
		Username:     username,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return domain.User{}, err
	}
	return toDomain(row)
}

func (r *Repo) ByUsername(ctx context.Context, username string) (domain.User, error) {
	row, err := r.q.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, fmt.Errorf("user not found: %w", err)
		}
		return domain.User{}, err
	}
	return toDomain(row)
}

func toDomain(row db.User) (domain.User, error) {
	id, err := uuidFromPg(row.ID)
	if err != nil {
		return domain.User{}, err
	}
	return domain.User{
		ID:           id,
		Username:     row.Username,
		PasswordHash: row.PasswordHash,
		CreatedAt:    row.CreatedAt.Time,
	}, nil
}

func pgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func uuidFromPg(v pgtype.UUID) (uuid.UUID, error) {
	if !v.Valid {
		return uuid.Nil, fmt.Errorf("invalid uuid")
	}
	return uuid.UUID(v.Bytes), nil
}
