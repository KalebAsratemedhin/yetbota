package auth

import (
	"context"
	"time"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	"github.com/golang-jwt/jwt"
)

type SessionDetails struct {
	AccessToken  string
	RefreshToken string
	AccessUuid   string
	RefreshUuid  string
	AccessTtl    time.Duration
	RefreshTtl   time.Duration
	Algo         jwt.SigningMethod
}

type UserSession struct {
	SessionID string
	Username  string
	Exp       float64
	UserID    string
	Role      string
}

type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

type TokenInfo struct {
	TokenType TokenType
	Token     string
}

type SessionInfo struct {
	Username   string        `validate:"required"`
	UserID     string        `validate:"required,uuid"`
	Role       string        `validate:"required"`
	RefreshTTL time.Duration `validate:"required"`
	AccessTTL  time.Duration `validate:"required"`
}

func (s *SessionInfo) Validate() error {
	if err := validator.Validate.Struct(s); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type SessionManager interface {
	NewSessionDetails(ctx context.Context, sessInfo *SessionInfo) (*SessionDetails, error)
	DeleteSession(ctx context.Context, userSess *UserSession) (int64, error)
	SaveSession(ctx context.Context, td *SessionDetails) error
	ExtractUserSession(ctx context.Context, token *TokenInfo) (*UserSession, error)
}

// GuestTokenSigner signs and verifies guest device tokens (long-lived, device-scoped).
type GuestTokenSigner interface {
	SignGuestToken(ctx context.Context, deviceID string) (string, error)
	ExtractGuestDeviceID(token string) (string, error)
}
