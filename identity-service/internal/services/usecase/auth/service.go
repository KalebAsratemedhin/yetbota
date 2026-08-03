package auth

import (
	"context"
	"time"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/validator"
	domainAuth "github.com/beka-birhanu/yetbota/identity-service/internal/domain/auth"
	contextYB "github.com/beka-birhanu/yetbota/identity-service/internal/domain/context"
	domainUser "github.com/beka-birhanu/yetbota/identity-service/internal/domain/user"
	domainDevice "github.com/beka-birhanu/yetbota/identity-service/internal/domain/userdevice"
)

type Service interface {
	Login(ctx context.Context, ctxSess *contextYB.Context, req *LoginRequest) (*LoginResponse, error)
	Refresh(ctx context.Context, ctxSess *contextYB.Context, req *RefreshRequest) (*RefreshResponse, error)
	Logout(ctx context.Context, ctxSess *contextYB.Context, req *LogoutRequest) (*LogoutResponse, error)
	Authorization(ctx context.Context, ctxSess *contextYB.Context, req *AuthorizationRequest) (*AuthorizationResponse, error)
	ChangePassword(ctx context.Context, ctxSess *contextYB.Context, req *ChangePasswordRequest) (*ChangePasswordResponse, error)
}

type Config struct {
	UserRepo       domainUser.Repository     `validate:"required"`
	SessionManager domainAuth.SessionManager `validate:"required"`
	Hasher         domainAuth.Hasher         `validate:"required"`
	AccessTTL      time.Duration             `validate:"required"`
	RefreshTTL     time.Duration             `validate:"required"`
	DeviceRepo     domainDevice.Repository
}

func (c *Config) Validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type service struct {
	userRepo       domainUser.Repository
	sessionManager domainAuth.SessionManager
	hasher         domainAuth.Hasher
	accessTTL      time.Duration
	refreshTTL     time.Duration
	deviceRepo     domainDevice.Repository
}

func NewService(cfg *Config) (Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &service{
		userRepo:       cfg.UserRepo,
		sessionManager: cfg.SessionManager,
		hasher:         cfg.Hasher,
		accessTTL:      cfg.AccessTTL,
		refreshTTL:     cfg.RefreshTTL,
		deviceRepo:     cfg.DeviceRepo,
	}, nil
}
