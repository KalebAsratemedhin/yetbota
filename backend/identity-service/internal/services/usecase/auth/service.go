package auth

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"

	domain "github.com/yetbota/identity-service/internal/domain/user"
)

type Service struct {
	users  domain.Repository
	hasher domain.PasswordHasher
	tokens domain.TokenIssuer
}

func NewService(users domain.Repository, hasher domain.PasswordHasher, tokens domain.TokenIssuer) *Service {
	return &Service{users: users, hasher: hasher, tokens: tokens}
}

type Credentials struct {
	Username string
	Password string
}

type AuthResult struct {
	UserID   string
	Username string
	Token    string
}

func (s *Service) SignUp(ctx context.Context, in Credentials) (AuthResult, error) {
	username, password, err := normalize(in)
	if err != nil {
		return AuthResult{}, err
	}
	hash, err := s.hasher.Hash(password)
	if err != nil {
		return AuthResult{}, err
	}
	u, err := s.users.Create(ctx, username, hash)
	if err != nil {
		if isUniqueViolation(err) {
			return AuthResult{}, domain.ErrUsernameTaken
		}
		return AuthResult{}, err
	}
	token, err := s.tokens.Issue(u.ID, u.Username)
	if err != nil {
		return AuthResult{}, err
	}
	return AuthResult{UserID: u.ID.String(), Username: u.Username, Token: token}, nil
}

func (s *Service) SignIn(ctx context.Context, in Credentials) (AuthResult, error) {
	username, password, err := normalize(in)
	if err != nil {
		return AuthResult{}, err
	}
	u, err := s.users.ByUsername(ctx, username)
	if err != nil {
		return AuthResult{}, domain.ErrInvalidCredentials
	}
	if err := s.hasher.Compare(u.PasswordHash, password); err != nil {
		return AuthResult{}, domain.ErrInvalidCredentials
	}
	token, err := s.tokens.Issue(u.ID, u.Username)
	if err != nil {
		return AuthResult{}, err
	}
	return AuthResult{UserID: u.ID.String(), Username: u.Username, Token: token}, nil
}

func normalize(in Credentials) (string, string, error) {
	username := strings.TrimSpace(in.Username)
	password := in.Password
	if username == "" || password == "" {
		return "", "", domain.ErrValidation
	}
	return username, password, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
