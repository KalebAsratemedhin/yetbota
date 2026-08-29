package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	driverauth "github.com/yetbota/identity-service/drivers/auth"
	domain "github.com/yetbota/identity-service/internal/domain/user"
)

type Service struct {
	users        domain.Repository
	hasher       domain.PasswordHasher
	tokens       domain.TokenService
	accessRevoke domain.TokenRevoker
	refreshStore domain.RefreshStore
	refreshTTL   time.Duration
}

func NewService(
	users domain.Repository,
	hasher domain.PasswordHasher,
	tokens domain.TokenService,
	accessRevoke domain.TokenRevoker,
	refreshStore domain.RefreshStore,
	refreshTTLDays int,
) *Service {
	if refreshTTLDays <= 0 {
		refreshTTLDays = 30
	}
	return &Service{
		users:        users,
		hasher:       hasher,
		tokens:       tokens,
		accessRevoke: accessRevoke,
		refreshStore: refreshStore,
		refreshTTL:   time.Duration(refreshTTLDays) * 24 * time.Hour,
	}
}

type Credentials struct {
	Username string
	Password string
}

type AuthResult struct {
	UserID       string
	Username     string
	AccessToken  string
	RefreshToken string
}

type UserInfo struct {
	UserID   string
	Username string
	FullName string
	Phone    string
	Address  string
	Bio      string
}

type IntrospectResult struct {
	Active   bool
	UserID   string
	Username string
}

type ChangePasswordInput struct {
	UserID          uuid.UUID
	CurrentPassword string
	NewPassword     string
}

type LogoutInput struct {
	Claims       domain.TokenClaims
	RefreshToken string
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
	return s.issuePair(ctx, u)
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
	return s.issuePair(ctx, u)
}

func (s *Service) Me(ctx context.Context, userID uuid.UUID) (UserInfo, error) {
	u, err := s.users.ByID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return UserInfo{}, domain.ErrUnauthorized
		}
		return UserInfo{}, err
	}
	return toUserInfo(u), nil
}

func (s *Service) UpdateProfile(ctx context.Context, userID uuid.UUID, patch domain.ProfileUpdate) (UserInfo, error) {
	if patch.FullName == nil && patch.Phone == nil && patch.Address == nil && patch.Bio == nil {
		return UserInfo{}, domain.ErrValidation
	}
	u, err := s.users.UpdateProfile(ctx, userID, patch)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return UserInfo{}, domain.ErrUnauthorized
		}
		return UserInfo{}, err
	}
	return toUserInfo(u), nil
}

func (s *Service) Introspect(ctx context.Context, token string) (IntrospectResult, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return IntrospectResult{Active: false}, nil
	}
	claims, err := s.validateAccess(ctx, token)
	if err != nil {
		return IntrospectResult{Active: false}, nil
	}
	if _, err := s.users.ByID(ctx, claims.UserID); err != nil {
		return IntrospectResult{Active: false}, nil
	}
	return IntrospectResult{
		Active:   true,
		UserID:   claims.UserID.String(),
		Username: claims.Username,
	}, nil
}

func (s *Service) Logout(ctx context.Context, in LogoutInput) error {
	if err := s.accessRevoke.RevokeAccess(ctx, in.Claims.JTI, in.Claims.ExpiresAt); err != nil {
		return err
	}
	raw := strings.TrimSpace(in.RefreshToken)
	if raw == "" {
		return domain.ErrValidation
	}
	return s.refreshStore.DeleteRefreshByHash(ctx, driverauth.HashRefresh(raw))
}

func (s *Service) ChangePassword(ctx context.Context, in ChangePasswordInput) error {
	if strings.TrimSpace(in.CurrentPassword) == "" || in.NewPassword == "" {
		return domain.ErrValidation
	}
	u, err := s.users.ByID(ctx, in.UserID)
	if err != nil {
		return domain.ErrUnauthorized
	}
	if err := s.hasher.Compare(u.PasswordHash, in.CurrentPassword); err != nil {
		return domain.ErrInvalidCredentials
	}
	hash, err := s.hasher.Hash(in.NewPassword)
	if err != nil {
		return err
	}
	return s.users.UpdatePassword(ctx, in.UserID, hash)
}

func (s *Service) Refresh(ctx context.Context, rawRefresh string) (AuthResult, error) {
	rawRefresh = strings.TrimSpace(rawRefresh)
	if rawRefresh == "" {
		return AuthResult{}, domain.ErrValidation
	}
	hash := driverauth.HashRefresh(rawRefresh)
	userID, expiresAt, err := s.refreshStore.GetRefreshByHash(ctx, hash)
	if err != nil {
		return AuthResult{}, domain.ErrUnauthorized
	}
	if time.Now().After(expiresAt) {
		_ = s.refreshStore.DeleteRefreshByHash(ctx, hash)
		return AuthResult{}, domain.ErrUnauthorized
	}
	u, err := s.users.ByID(ctx, userID)
	if err != nil {
		return AuthResult{}, domain.ErrUnauthorized
	}
	if err := s.refreshStore.DeleteRefreshByHash(ctx, hash); err != nil {
		return AuthResult{}, err
	}
	return s.issuePair(ctx, u)
}

func (s *Service) DeleteAccount(ctx context.Context, claims domain.TokenClaims) error {
	_ = s.refreshStore.DeleteRefreshByUserID(ctx, claims.UserID)
	_ = s.accessRevoke.RevokeAccess(ctx, claims.JTI, claims.ExpiresAt)
	return s.users.Delete(ctx, claims.UserID)
}

func (s *Service) AuthenticateToken(ctx context.Context, token string) (domain.TokenClaims, error) {
	return s.validateAccess(ctx, token)
}

func (s *Service) validateAccess(ctx context.Context, token string) (domain.TokenClaims, error) {
	claims, err := s.tokens.ParseAccess(token)
	if err != nil {
		return domain.TokenClaims{}, domain.ErrUnauthorized
	}
	revoked, err := s.accessRevoke.IsAccessRevoked(ctx, claims.JTI)
	if err != nil {
		return domain.TokenClaims{}, err
	}
	if revoked {
		return domain.TokenClaims{}, domain.ErrUnauthorized
	}
	return claims, nil
}

func (s *Service) issuePair(ctx context.Context, u domain.User) (AuthResult, error) {
	access, _, err := s.tokens.IssueAccess(u.ID, u.Username)
	if err != nil {
		return AuthResult{}, err
	}
	rawRefresh, hash, err := s.tokens.NewRefreshToken()
	if err != nil {
		return AuthResult{}, err
	}
	exp := time.Now().Add(s.refreshTTL)
	if err := s.refreshStore.SaveRefresh(ctx, u.ID, hash, exp); err != nil {
		return AuthResult{}, err
	}
	return AuthResult{
		UserID:       u.ID.String(),
		Username:     u.Username,
		AccessToken:  access,
		RefreshToken: rawRefresh,
	}, nil
}

func toUserInfo(u domain.User) UserInfo {
	return UserInfo{
		UserID:   u.ID.String(),
		Username: u.Username,
		FullName: u.FullName,
		Phone:    u.Phone,
		Address:  u.Address,
		Bio:      u.Bio,
	}
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
