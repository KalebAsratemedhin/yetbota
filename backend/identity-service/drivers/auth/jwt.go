package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	domain "github.com/yetbota/identity-service/internal/domain/user"
)

type JWTService struct {
	secret    []byte
	accessTTL time.Duration
}

func NewJWTService(secret string, accessTTLMinutes int) *JWTService {
	if accessTTLMinutes <= 0 {
		accessTTLMinutes = 15
	}
	return &JWTService{
		secret:    []byte(secret),
		accessTTL: time.Duration(accessTTLMinutes) * time.Minute,
	}
}

// NewJWTIssuer kept for call-site compatibility during rename.
func NewJWTIssuer(secret string, accessTTLMinutes int) *JWTService {
	return NewJWTService(secret, accessTTLMinutes)
}

type claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func (i *JWTService) IssueAccess(userID uuid.UUID, username string) (string, domain.TokenClaims, error) {
	now := time.Now()
	jti := uuid.NewString()
	exp := now.Add(i.accessTTL)
	c := claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			ID:        jti,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	signed, err := t.SignedString(i.secret)
	if err != nil {
		return "", domain.TokenClaims{}, fmt.Errorf("sign jwt: %w", err)
	}
	return signed, domain.TokenClaims{
		UserID:    userID,
		Username:  username,
		JTI:       jti,
		ExpiresAt: exp,
	}, nil
}

func (i *JWTService) ParseAccess(token string) (domain.TokenClaims, error) {
	parsed, err := jwt.ParseWithClaims(token, &claims{}, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return i.secret, nil
	})
	if err != nil || !parsed.Valid {
		return domain.TokenClaims{}, domain.ErrUnauthorized
	}
	c, ok := parsed.Claims.(*claims)
	if !ok || c.ID == "" || c.Subject == "" {
		return domain.TokenClaims{}, domain.ErrUnauthorized
	}
	uid, err := uuid.Parse(c.Subject)
	if err != nil {
		return domain.TokenClaims{}, domain.ErrUnauthorized
	}
	var exp time.Time
	if c.ExpiresAt != nil {
		exp = c.ExpiresAt.Time
	}
	return domain.TokenClaims{
		UserID:    uid,
		Username:  c.Username,
		JTI:       c.ID,
		ExpiresAt: exp,
	}, nil
}

func (i *JWTService) NewRefreshToken() (raw string, hash string, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", fmt.Errorf("refresh token: %w", err)
	}
	raw = base64.RawURLEncoding.EncodeToString(b)
	sum := sha256.Sum256([]byte(raw))
	hash = hex.EncodeToString(sum[:])
	return raw, hash, nil
}

// HashRefresh returns the storage hash for a raw refresh token.
func HashRefresh(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
