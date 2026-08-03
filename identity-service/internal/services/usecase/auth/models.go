package auth

import (
	"strings"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/validator"
)

type LoginRequest struct {
	Username string `validate:"required"`
	Password string `validate:"required" mask:"true"`
	Site     string
}

func (r *LoginRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

func (r *LoginRequest) normalize() error {
	r.Username = strings.ToLower(r.Username)
	return nil
}

type LoginResponse struct {
	AccessToken     string
	AccessTokenTTL  int64
	RefreshToken    string
	RefreshTokenTTL int64
}

type RefreshRequest struct {
	RefreshToken string `validate:"required" mask:"true"`
	Username     string `validate:"required"`
}

func (r *RefreshRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type RefreshResponse struct {
	AccessToken     string
	AccessTokenTTL  int64
	RefreshToken    string
	RefreshTokenTTL int64
}

type LogoutRequest struct {
	RefreshToken string `validate:"required" mask:"true"`
	Username     string `validate:"required"`
	DeviceID     string
}

func (r *LogoutRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type LogoutResponse struct{}

type AuthorizationRequest struct {
	Resource string `validate:"required"`
	Action   string `validate:"required"`
}

func (r *AuthorizationRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type AuthorizationResponse struct{}

type ChangePasswordRequest struct {
	CurrentPassword string `validate:"required" mask:"true"`
	NewPassword     string `validate:"required,min=8" mask:"true"`
}

func (r *ChangePasswordRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type ChangePasswordResponse struct{}
