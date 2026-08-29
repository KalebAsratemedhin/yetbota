package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	driverauth "github.com/yetbota/identity-service/drivers/auth"
	"github.com/yetbota/identity-service/internal/domain/apperr"
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
	FullName string // required for SignUp; ignored for SignIn
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
	username, password, fullName, err := normalizeSignUp(in)
	if err != nil {
		return AuthResult{}, err
	}
	hash, err := s.hasher.Hash(password)
	if err != nil {
		return AuthResult{}, apperr.Wrap(err, apperr.CodeInternal, "hash password")
	}
	u, err := s.users.Create(ctx, username, hash, fullName)
	if err != nil {
		if isUniqueViolation(err) {
			return AuthResult{}, apperr.ErrUsernameTaken
		}
		return AuthResult{}, err
	}
	return s.issuePair(ctx, u)
}

func (s *Service) SignIn(ctx context.Context, in Credentials) (AuthResult, error) {
	username, password, err := normalizeSignIn(in)
	if err != nil {
		return AuthResult{}, err
	}
	u, err := s.users.ByUsername(ctx, username)
	if err != nil {
		return AuthResult{}, apperr.ErrInvalidCredentials
	}
	if err := s.hasher.Compare(u.PasswordHash, password); err != nil {
		return AuthResult{}, apperr.ErrInvalidCredentials
	}
	return s.issuePair(ctx, u)
}

func (s *Service) Me(ctx context.Context, userID uuid.UUID) (UserInfo, error) {
	u, err := s.users.ByID(ctx, userID)
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			return UserInfo{}, apperr.ErrUnauthorized
		}
		return UserInfo{}, err
	}
	return toUserInfo(u), nil
}

func (s *Service) UpdateProfile(ctx context.Context, userID uuid.UUID, patch domain.ProfileUpdate) (UserInfo, error) {
	if patch.FullName == nil && patch.Phone == nil && patch.Address == nil && patch.Bio == nil {
		return UserInfo{}, apperr.Validation(apperr.Field{
			Field: "body", Code: "EMPTY", Message: "at least one profile field is required",
		})
	}
	u, err := s.users.UpdateProfile(ctx, userID, patch)
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			return UserInfo{}, apperr.ErrUnauthorized
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
		return apperr.Wrap(err, apperr.CodeInternal, "revoke access token")
	}
	raw := strings.TrimSpace(in.RefreshToken)
	if raw == "" {
		return apperr.Validation(apperr.Required("refresh_token"))
	}
	if err := s.refreshStore.DeleteRefreshByHash(ctx, driverauth.HashRefresh(raw)); err != nil {
		return apperr.Wrap(err, apperr.CodeInternal, "revoke refresh token")
	}
	return nil
}

func (s *Service) ChangePassword(ctx context.Context, in ChangePasswordInput) error {
	var details []apperr.Field
	if strings.TrimSpace(in.CurrentPassword) == "" {
		details = append(details, apperr.Required("current_password"))
	}
	if in.NewPassword == "" {
		details = append(details, apperr.Required("new_password"))
	}
	if len(details) > 0 {
		return apperr.Validation(details...)
	}
	u, err := s.users.ByID(ctx, in.UserID)
	if err != nil {
		return apperr.ErrUnauthorized
	}
	if err := s.hasher.Compare(u.PasswordHash, in.CurrentPassword); err != nil {
		return apperr.ErrInvalidCredentials
	}
	hash, err := s.hasher.Hash(in.NewPassword)
	if err != nil {
		return apperr.Wrap(err, apperr.CodeInternal, "hash password")
	}
	if err := s.users.UpdatePassword(ctx, in.UserID, hash); err != nil {
		return apperr.Wrap(err, apperr.CodeInternal, "update password")
	}
	return nil
}

func (s *Service) Refresh(ctx context.Context, rawRefresh string) (AuthResult, error) {
	rawRefresh = strings.TrimSpace(rawRefresh)
	if rawRefresh == "" {
		return AuthResult{}, apperr.Validation(apperr.Required("refresh_token"))
	}
	hash := driverauth.HashRefresh(rawRefresh)
	userID, expiresAt, err := s.refreshStore.GetRefreshByHash(ctx, hash)
	if err != nil {
		return AuthResult{}, apperr.ErrUnauthorized
	}
	if time.Now().After(expiresAt) {
		_ = s.refreshStore.DeleteRefreshByHash(ctx, hash)
		return AuthResult{}, apperr.ErrUnauthorized
	}
	u, err := s.users.ByID(ctx, userID)
	if err != nil {
		return AuthResult{}, apperr.ErrUnauthorized
	}
	if err := s.refreshStore.DeleteRefreshByHash(ctx, hash); err != nil {
		return AuthResult{}, apperr.Wrap(err, apperr.CodeInternal, "rotate refresh token")
	}
	return s.issuePair(ctx, u)
}

func (s *Service) DeleteAccount(ctx context.Context, claims domain.TokenClaims) error {
	_ = s.refreshStore.DeleteRefreshByUserID(ctx, claims.UserID)
	_ = s.accessRevoke.RevokeAccess(ctx, claims.JTI, claims.ExpiresAt)
	if err := s.users.Delete(ctx, claims.UserID); err != nil {
		return apperr.Wrap(err, apperr.CodeInternal, "delete user")
	}
	return nil
}

func (s *Service) AuthenticateToken(ctx context.Context, token string) (domain.TokenClaims, error) {
	return s.validateAccess(ctx, token)
}

func (s *Service) validateAccess(ctx context.Context, token string) (domain.TokenClaims, error) {
	claims, err := s.tokens.ParseAccess(token)
	if err != nil {
		return domain.TokenClaims{}, apperr.ErrUnauthorized
	}
	revoked, err := s.accessRevoke.IsAccessRevoked(ctx, claims.JTI)
	if err != nil {
		return domain.TokenClaims{}, apperr.Wrap(err, apperr.CodeInternal, "check token revocation")
	}
	if revoked {
		return domain.TokenClaims{}, apperr.ErrUnauthorized
	}
	return claims, nil
}

func (s *Service) issuePair(ctx context.Context, u domain.User) (AuthResult, error) {
	access, _, err := s.tokens.IssueAccess(u.ID, u.Username)
	if err != nil {
		return AuthResult{}, apperr.Wrap(err, apperr.CodeInternal, "issue access token")
	}
	rawRefresh, hash, err := s.tokens.NewRefreshToken()
	if err != nil {
		return AuthResult{}, apperr.Wrap(err, apperr.CodeInternal, "issue refresh token")
	}
	exp := time.Now().Add(s.refreshTTL)
	if err := s.refreshStore.SaveRefresh(ctx, u.ID, hash, exp); err != nil {
		return AuthResult{}, apperr.Wrap(err, apperr.CodeInternal, "store refresh token")
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

func normalizeSignUp(in Credentials) (username, password, fullName string, err error) {
	username = strings.TrimSpace(in.Username)
	password = in.Password
	fullName = strings.TrimSpace(in.FullName)
	var details []apperr.Field
	if username == "" {
		details = append(details, apperr.Required("username"))
	}
	if password == "" {
		details = append(details, apperr.Required("password"))
	}
	if fullName == "" {
		details = append(details, apperr.Required("full_name"))
	}
	if len(details) > 0 {
		return "", "", "", apperr.Validation(details...)
	}
	return username, password, fullName, nil
}

func normalizeSignIn(in Credentials) (string, string, error) {
	username := strings.TrimSpace(in.Username)
	password := in.Password
	var details []apperr.Field
	if username == "" {
		details = append(details, apperr.Required("username"))
	}
	if password == "" {
		details = append(details, apperr.Required("password"))
	}
	if len(details) > 0 {
		return "", "", apperr.Validation(details...)
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
