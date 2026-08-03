package jwt

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/beka-birhanu/yetbota/identity-service/drivers/validator"
	"github.com/beka-birhanu/yetbota/identity-service/internal/domain/auth"
	"github.com/go-redis/redis/v8"

	toddlerErr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

const guestTokenTTL = 365 * 24 * time.Hour

var claimKeys = []string{
	"session_id",
	"username",
	"exp",
	"user_id",
	"role",
}

type SessionManager struct {
	accessKey     string
	refreshKey    string
	guestKey      string
	accessTTL     time.Duration
	refreshTTL    time.Duration
	signingMethod jwt.SigningMethod
	redisConn     *redis.Client
}

type Config struct {
	AccessKey  string        `validate:"required"`
	RefreshKey string        `validate:"required"`
	GuestKey   string        // optional; defaults to AccessKey+"_guest"
	AccessTTL  time.Duration `validate:"required"`
	RefreshTTL time.Duration `validate:"required"`
	Algo       string        `validate:"required"`
	RedisConn  *redis.Client `validate:"required"`
}

func (c *Config) Validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return toddlerErr.FromValidationErrors(err)
	}
	return nil
}

func NewSessionManager(cfg *Config) (*SessionManager, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	guestKey := cfg.GuestKey
	if guestKey == "" {
		guestKey = cfg.AccessKey + "_guest"
	}

	return &SessionManager{
		accessKey:     cfg.AccessKey,
		refreshKey:    cfg.RefreshKey,
		guestKey:      guestKey,
		accessTTL:     cfg.AccessTTL,
		refreshTTL:    cfg.RefreshTTL,
		signingMethod: jwt.GetSigningMethod(cfg.Algo),
		redisConn:     cfg.RedisConn,
	}, nil
}

func (s *SessionManager) NewSessionDetails(ctx context.Context, sessInfo *auth.SessionInfo) (*auth.SessionDetails, error) {
	accessUuid := uuid.NewString()
	refreshUuid := accessUuid + "++" + sessInfo.Username

	td := &auth.SessionDetails{
		AccessTtl:   s.accessTTL,
		AccessUuid:  accessUuid,
		RefreshTtl:  s.refreshTTL,
		RefreshUuid: refreshUuid,
	}

	// Creating Access Token
	atClaims := jwt.MapClaims{
		"session_id": accessUuid,
		"username":   sessInfo.Username,
		"user_id":    sessInfo.UserID,
		"role":       sessInfo.Role,
		"exp":        time.Now().Add(s.accessTTL).Unix(),
	}
	at := jwt.NewWithClaims(s.signingMethod, atClaims)
	var err error
	td.AccessToken, err = at.SignedString([]byte(s.accessKey))
	if err != nil {
		return nil, &toddlerErr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "Something went wrong",
			ServiceMessage:    "JWT signing failed while creating Access Token: " + err.Error(),
		}
	}

	// Creating Refresh Token
	rtClaims := jwt.MapClaims{
		"session_id": td.RefreshUuid,
		"username":   sessInfo.Username,
		"user_id":    sessInfo.UserID,
		"role":       sessInfo.Role,
		"exp":        time.Now().Add(s.refreshTTL).Unix(),
	}
	rt := jwt.NewWithClaims(s.signingMethod, rtClaims)
	td.RefreshToken, err = rt.SignedString([]byte(s.refreshKey))
	if err != nil {
		return nil, &toddlerErr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "Something went wrong",
			ServiceMessage:    "JWT signing failed while creating Refresh Token: " + err.Error(),
		}
	}

	return td, nil
}

// DeleteSession removes auth sessions from Redis.
// Returns delete count. Enforcing session deletion check in on the caller.
func (s *SessionManager) DeleteSession(ctx context.Context, userSess *auth.UserSession) (int64, error) {
	deleted, err := s.redisConn.Del(ctx, userSess.SessionID).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to delete session: %w", err)
	}

	return deleted, nil
}

func (s *SessionManager) SaveSession(ctx context.Context, td *auth.SessionDetails) error {
	errAccess := s.redisConn.Set(ctx, td.AccessUuid, true, td.AccessTtl).Err()
	if errAccess != nil {
		return &toddlerErr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "Something went wrong",
			ServiceMessage:    "Redis error while storing access token: " + errAccess.Error(),
		}
	}

	errRefresh := s.redisConn.Set(ctx, td.RefreshUuid, true, td.RefreshTtl).Err()
	if errRefresh != nil {
		return &toddlerErr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "Something went wrong",
			ServiceMessage:    "Redis error while storing refresh token: " + errRefresh.Error(),
		}
	}

	return nil
}

func (s *SessionManager) ExtractUserSession(ctx context.Context, tokenInfo *auth.TokenInfo) (*auth.UserSession, error) {
	var key string
	switch tokenInfo.TokenType {
	case auth.AccessToken:
		key = s.accessKey
	case auth.RefreshToken:
		key = s.refreshKey
	}

	token, err := s.parseToken(tokenInfo.Token, key)
	if err != nil {
		return nil, err
	}

	claims, err := validateToken(token)
	if err != nil {
		return nil, err
	}

	sessionID, ok1 := claims["session_id"].(string)
	username, ok2 := claims["username"].(string)
	exp, ok3 := claims["exp"].(float64)
	userID, ok4 := claims["user_id"].(string)
	role, ok5 := claims["role"].(string)
	if !ok1 || !ok2 || !ok3 || !ok4 || !ok5 {
		return nil, &toddlerErr.Error{
			PublicStatusCode:  status.BadRequest,
			ServiceStatusCode: status.BadRequest,
			PublicMessage:     "Invalid token",
			ServiceMessage:    "unexpected claim types in token",
		}
	}

	exists, err := s.redisConn.Exists(ctx, sessionID).Result()
	if err != nil {
		return nil, &toddlerErr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "Something went wrong",
			ServiceMessage:    "Redis error while verifying session: " + err.Error(),
		}
	}
	if exists == 0 {
		return nil, &toddlerErr.Error{
			PublicStatusCode:  status.Unauthorized,
			ServiceStatusCode: status.Unauthorized,
			PublicMessage:     "Session expired or logged out",
			ServiceMessage:    "session not found in Redis",
		}
	}

	return &auth.UserSession{
		SessionID: sessionID,
		Username:  username,
		Exp:       exp,
		UserID:    userID,
		Role:      role,
	}, nil
}

func (s *SessionManager) parseToken(tokenStr, key string) (*jwt.Token, error) {
	parts := strings.SplitN(tokenStr, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, &toddlerErr.Error{
			PublicStatusCode:  status.BadRequest,
			ServiceStatusCode: status.BadRequestInvalidFormat,
			PublicMessage:     "Invalid token format",
			ServiceMessage:    "Bearer token structure incorrect",
		}
	}

	token, err := jwt.Parse(parts[1], func(token *jwt.Token) (any, error) {
		if s.signingMethod != token.Method {
			return nil, &toddlerErr.Error{
				PublicStatusCode:  status.BadRequest,
				ServiceStatusCode: status.BadRequest,
				PublicMessage:     "Invalid token",
				ServiceMessage:    "unexpected JWT signing method",
			}
		}
		return []byte(key), nil
	})
	if err != nil {
		// Treat expiration as an authentication failure (401) rather than a generic bad request.
		if ve, ok := err.(*jwt.ValidationError); ok && (ve.Errors&jwt.ValidationErrorExpired) != 0 {
			return nil, &toddlerErr.Error{
				PublicStatusCode:  status.Unauthorized,
				ServiceStatusCode: status.Unauthorized,
				PublicMessage:     "Session expired or logged out",
				ServiceMessage:    fmt.Sprintf("jwt.Parse expired: %v", err),
			}
		}
		return nil, &toddlerErr.Error{
			PublicStatusCode:  status.BadRequest,
			ServiceStatusCode: status.BadRequest,
			PublicMessage:     "Invalid token",
			ServiceMessage:    fmt.Sprintf("jwt.Parse error: %v", err),
		}
	}

	return token, nil
}

// SignGuestToken creates a JWT embedding the guest device_id, stores it in Redis
// with a 1-year TTL, and returns the signed token string.
func (s *SessionManager) SignGuestToken(ctx context.Context, deviceID string) (string, error) {
	claims := jwt.MapClaims{
		"device_id": deviceID,
		"exp":       time.Now().Add(guestTokenTTL).Unix(),
	}
	t := jwt.NewWithClaims(s.signingMethod, claims)
	signed, err := t.SignedString([]byte(s.guestKey))
	if err != nil {
		return "", &toddlerErr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "Something went wrong",
			ServiceMessage:    "JWT signing failed while creating guest token: " + err.Error(),
		}
	}

	if err := s.redisConn.Set(ctx, "guest:"+deviceID, true, guestTokenTTL).Err(); err != nil {
		return "", &toddlerErr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "Something went wrong",
			ServiceMessage:    "Redis error while storing guest token: " + err.Error(),
		}
	}

	return signed, nil
}

// ExtractGuestDeviceID parses and verifies a guest token, returning the embedded device_id.
func (s *SessionManager) ExtractGuestDeviceID(token string) (string, error) {
	parsed, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		if s.signingMethod != t.Method {
			return nil, &toddlerErr.Error{
				PublicStatusCode:  status.BadRequest,
				ServiceStatusCode: status.BadRequest,
				PublicMessage:     "Invalid token",
				ServiceMessage:    "unexpected JWT signing method",
			}
		}
		return []byte(s.guestKey), nil
	})
	if err != nil {
		return "", &toddlerErr.Error{
			PublicStatusCode:  status.BadRequest,
			ServiceStatusCode: status.BadRequest,
			PublicMessage:     "Invalid token",
			ServiceMessage:    fmt.Sprintf("guest token parse error: %v", err),
		}
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok || !parsed.Valid {
		return "", &toddlerErr.Error{
			PublicStatusCode:  status.BadRequest,
			ServiceStatusCode: status.BadRequest,
			PublicMessage:     "Invalid token",
			ServiceMessage:    "invalid or malformed guest token claims",
		}
	}

	deviceID, ok := claims["device_id"].(string)
	if !ok || deviceID == "" {
		return "", &toddlerErr.Error{
			PublicStatusCode:  status.BadRequest,
			ServiceStatusCode: status.BadRequest,
			PublicMessage:     "Invalid token",
			ServiceMessage:    "guest token missing device_id claim",
		}
	}

	return deviceID, nil
}

func validateToken(token *jwt.Token) (jwt.MapClaims, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, &toddlerErr.Error{
			PublicStatusCode:  status.BadRequest,
			ServiceStatusCode: status.BadRequest,
			PublicMessage:     "Invalid token",
			ServiceMessage:    "invalid or malformed token claims",
		}
	}

	for _, key := range claimKeys {
		if _, ok := claims[key]; !ok {
			return nil, &toddlerErr.Error{
				PublicStatusCode:  status.BadRequest,
				ServiceStatusCode: status.BadRequest,
				PublicMessage:     "Invalid token",
				ServiceMessage:    fmt.Sprintf("missing claim: %s", key),
			}
		}
	}

	return claims, nil
}
