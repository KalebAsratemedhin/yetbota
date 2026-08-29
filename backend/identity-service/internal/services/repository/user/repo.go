package user

import (
	"context"
	"errors"
	"fmt"
	"time"

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
			return domain.User{}, domain.ErrNotFound
		}
		return domain.User{}, err
	}
	return toDomain(row)
}

func (r *Repo) ByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	row, err := r.q.GetUserByID(ctx, pgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.ErrNotFound
		}
		return domain.User{}, err
	}
	return toDomain(row)
}

func (r *Repo) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	return r.q.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		ID:           pgUUID(id),
		PasswordHash: passwordHash,
	})
}

func (r *Repo) UpdateProfile(ctx context.Context, id uuid.UUID, patch domain.ProfileUpdate) (domain.User, error) {
	row, err := r.q.UpdateUserProfile(ctx, db.UpdateUserProfileParams{
		ID:       pgUUID(id),
		FullName: textPtr(patch.FullName),
		Phone:    textPtr(patch.Phone),
		Address:  textPtr(patch.Address),
		Bio:      textPtr(patch.Bio),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.ErrNotFound
		}
		return domain.User{}, err
	}
	return toDomain(row)
}

func (r *Repo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteUser(ctx, pgUUID(id))
}

func (r *Repo) RevokeAccess(ctx context.Context, jti string, expiresAt time.Time) error {
	return r.q.RevokeToken(ctx, db.RevokeTokenParams{
		Jti:       jti,
		ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})
}

func (r *Repo) IsAccessRevoked(ctx context.Context, jti string) (bool, error) {
	return r.q.IsTokenRevoked(ctx, jti)
}

func (r *Repo) SaveRefresh(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	_, err := r.q.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
		ID:        pgUUID(uuid.New()),
		UserID:    pgUUID(userID),
		TokenHash: tokenHash,
		ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})
	return err
}

func (r *Repo) GetRefreshByHash(ctx context.Context, tokenHash string) (uuid.UUID, time.Time, error) {
	row, err := r.q.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, time.Time{}, domain.ErrUnauthorized
		}
		return uuid.Nil, time.Time{}, err
	}
	uid, err := uuidFromPg(row.UserID)
	if err != nil {
		return uuid.Nil, time.Time{}, err
	}
	return uid, row.ExpiresAt.Time, nil
}

func (r *Repo) DeleteRefreshByHash(ctx context.Context, tokenHash string) error {
	return r.q.DeleteRefreshTokenByHash(ctx, tokenHash)
}

func (r *Repo) DeleteRefreshByUserID(ctx context.Context, userID uuid.UUID) error {
	return r.q.DeleteRefreshTokensByUserID(ctx, pgUUID(userID))
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
		FullName:     row.FullName,
		Phone:        row.Phone,
		Address:      row.Address,
		Bio:          row.Bio,
	}, nil
}

func textPtr(v *string) pgtype.Text {
	if v == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *v, Valid: true}
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
