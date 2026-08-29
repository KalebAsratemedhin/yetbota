package user

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrUsernameTaken      = errors.New("username already taken")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrValidation         = errors.New("validation error")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrNotFound           = errors.New("not found")
)

type User struct {
	ID           uuid.UUID
	Username     string
	PasswordHash string
	CreatedAt    time.Time
	FullName     string
	Phone        string
	Address      string
	Bio          string
}

type ProfileUpdate struct {
	FullName *string
	Phone    *string
	Address  *string
	Bio      *string
}

type Repository interface {
	Create(ctx context.Context, username, passwordHash string) (User, error)
	ByUsername(ctx context.Context, username string) (User, error)
	ByID(ctx context.Context, id uuid.UUID) (User, error)
	UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error
	UpdateProfile(ctx context.Context, id uuid.UUID, patch ProfileUpdate) (User, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hash, password string) error
}

type TokenClaims struct {
	UserID    uuid.UUID
	Username  string
	JTI       string
	ExpiresAt time.Time
}

type TokenService interface {
	IssueAccess(userID uuid.UUID, username string) (token string, claims TokenClaims, err error)
	ParseAccess(token string) (TokenClaims, error)
	NewRefreshToken() (raw string, hash string, err error)
}

type TokenRevoker interface {
	RevokeAccess(ctx context.Context, jti string, expiresAt time.Time) error
	IsAccessRevoked(ctx context.Context, jti string) (bool, error)
}

type RefreshStore interface {
	SaveRefresh(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error
	GetRefreshByHash(ctx context.Context, tokenHash string) (userID uuid.UUID, expiresAt time.Time, err error)
	DeleteRefreshByHash(ctx context.Context, tokenHash string) error
	DeleteRefreshByUserID(ctx context.Context, userID uuid.UUID) error
}
