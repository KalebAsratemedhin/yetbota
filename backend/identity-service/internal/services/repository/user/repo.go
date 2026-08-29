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
	"github.com/yetbota/identity-service/internal/domain/apperr"
	domain "github.com/yetbota/identity-service/internal/domain/user"
)

type Repo struct {
	q *db.Queries
}

func New(pool *pgxpool.Pool) *Repo {
	return &Repo{q: db.New(pool)}
}

func (r *Repo) Create(ctx context.Context, username, passwordHash, fullName string) (domain.User, error) {
	id := uuid.New()
	row, err := r.q.CreateUser(ctx, db.CreateUserParams{
		ID:           pgUUID(id),
		Username:     username,
		PasswordHash: passwordHash,
		FullName:     fullName,
	})
	if err != nil {
		return domain.User{}, apperr.Wrap(err, apperr.CodeInternal, "create user")
	}
	return toDomain(row)
}

func (r *Repo) ByUsername(ctx context.Context, username string) (domain.User, error) {
	row, err := r.q.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, apperr.ErrNotFound
		}
		return domain.User{}, apperr.Wrap(err, apperr.CodeInternal, "get user by username")
	}
	return toDomain(row)
}

func (r *Repo) ByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	row, err := r.q.GetUserByID(ctx, pgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, apperr.ErrNotFound
		}
		return domain.User{}, apperr.Wrap(err, apperr.CodeInternal, "get user by id")
	}
	return toDomain(row)
}

func (r *Repo) ByIDs(ctx context.Context, ids []uuid.UUID) ([]domain.User, error) {
	if len(ids) == 0 {
		return []domain.User{}, nil
	}
	pgIDs := make([]pgtype.UUID, len(ids))
	for i, id := range ids {
		pgIDs[i] = pgUUID(id)
	}
	rows, err := r.q.GetUsersByIDs(ctx, pgIDs)
	if err != nil {
		return nil, apperr.Wrap(err, apperr.CodeInternal, "get users by ids")
	}
	out := make([]domain.User, 0, len(rows))
	for _, row := range rows {
		u, err := toDomain(row)
		if err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, nil
}

func (r *Repo) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	if err := r.q.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		ID:           pgUUID(id),
		PasswordHash: passwordHash,
	}); err != nil {
		return apperr.Wrap(err, apperr.CodeInternal, "update password")
	}
	return nil
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
			return domain.User{}, apperr.ErrNotFound
		}
		return domain.User{}, apperr.Wrap(err, apperr.CodeInternal, "update profile")
	}
	return toDomain(row)
}

func (r *Repo) UpdateAvatar(ctx context.Context, id uuid.UUID, avatarURL, publicID string) (domain.User, error) {
	row, err := r.q.UpdateUserAvatar(ctx, db.UpdateUserAvatarParams{
		ID:             pgUUID(id),
		AvatarUrl:      avatarURL,
		AvatarPublicID: publicID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, apperr.ErrNotFound
		}
		return domain.User{}, apperr.Wrap(err, apperr.CodeInternal, "update avatar")
	}
	return toDomain(row)
}

func (r *Repo) Search(ctx context.Context, query string, excludeID uuid.UUID, limit int32) ([]domain.User, error) {
	rows, err := r.q.SearchUsers(ctx, db.SearchUsersParams{
		Query:       query,
		ExcludeID:   pgUUID(excludeID),
		ResultLimit: limit,
	})
	if err != nil {
		return nil, apperr.Wrap(err, apperr.CodeInternal, "search users")
	}
	out := make([]domain.User, 0, len(rows))
	for _, row := range rows {
		u, err := toDomain(row)
		if err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, nil
}

func (r *Repo) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.q.DeleteUser(ctx, pgUUID(id)); err != nil {
		return apperr.Wrap(err, apperr.CodeInternal, "delete user")
	}
	return nil
}

func (r *Repo) RevokeAccess(ctx context.Context, jti string, expiresAt time.Time) error {
	if err := r.q.RevokeToken(ctx, db.RevokeTokenParams{
		Jti:       jti,
		ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
	}); err != nil {
		return apperr.Wrap(err, apperr.CodeInternal, "revoke access jti")
	}
	return nil
}

func (r *Repo) IsAccessRevoked(ctx context.Context, jti string) (bool, error) {
	revoked, err := r.q.IsTokenRevoked(ctx, jti)
	if err != nil {
		return false, apperr.Wrap(err, apperr.CodeInternal, "check revoked jti")
	}
	return revoked, nil
}

func (r *Repo) SaveRefresh(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	_, err := r.q.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
		ID:        pgUUID(uuid.New()),
		UserID:    pgUUID(userID),
		TokenHash: tokenHash,
		ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})
	if err != nil {
		return apperr.Wrap(err, apperr.CodeInternal, "save refresh token")
	}
	return nil
}

func (r *Repo) GetRefreshByHash(ctx context.Context, tokenHash string) (uuid.UUID, time.Time, error) {
	row, err := r.q.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, time.Time{}, apperr.ErrUnauthorized
		}
		return uuid.Nil, time.Time{}, apperr.Wrap(err, apperr.CodeInternal, "get refresh token")
	}
	uid, err := uuidFromPg(row.UserID)
	if err != nil {
		return uuid.Nil, time.Time{}, apperr.Wrap(err, apperr.CodeInternal, "parse refresh user id")
	}
	return uid, row.ExpiresAt.Time, nil
}

func (r *Repo) DeleteRefreshByHash(ctx context.Context, tokenHash string) error {
	if err := r.q.DeleteRefreshTokenByHash(ctx, tokenHash); err != nil {
		return apperr.Wrap(err, apperr.CodeInternal, "delete refresh token")
	}
	return nil
}

func (r *Repo) DeleteRefreshByUserID(ctx context.Context, userID uuid.UUID) error {
	if err := r.q.DeleteRefreshTokensByUserID(ctx, pgUUID(userID)); err != nil {
		return apperr.Wrap(err, apperr.CodeInternal, "delete user refresh tokens")
	}
	return nil
}

func toDomain(row db.User) (domain.User, error) {
	id, err := uuidFromPg(row.ID)
	if err != nil {
		return domain.User{}, err
	}
	return domain.User{
		ID:             id,
		Username:       row.Username,
		PasswordHash:   row.PasswordHash,
		CreatedAt:      row.CreatedAt.Time,
		FullName:       row.FullName,
		Phone:          row.Phone,
		Address:        row.Address,
		Bio:            row.Bio,
		AvatarURL:      row.AvatarUrl,
		AvatarPublicID: row.AvatarPublicID,
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
