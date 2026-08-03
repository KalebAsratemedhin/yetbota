package user

import (
	"strings"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/dbmodels"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/validator"
	domainUser "github.com/beka-birhanu/yetbota/identity-service/internal/domain/user"
)

// PhotoResolution represents the requested resolution for a profile photo.
type PhotoResolution string

const (
	PhotoResolutionUnspecified PhotoResolution = ""
	PhotoResolutionOriginal    PhotoResolution = "ORIGINAL"
	PhotoResolutionMobile      PhotoResolution = "MOBILE"
	PhotoResolutionWeb         PhotoResolution = "WEB"
)

// List
type ListRequest struct {
	Options    *domainUser.Options    `validate:"required"`
	Pagination *domainUser.Pagination `validate:"required"`
	Sort       *domainUser.SortOption `validate:"required"`
	Resolution PhotoResolution
}

func (r *ListRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type PaginationInfo struct {
	Total       int64
	Limit       int
	CurrentPage int
}

// UserWrapper wraps a user with a signed profile URL.
type UserWrapper struct {
	User       *dbmodels.User
	ProfileURL string
}

type ListResponse struct {
	Users      []*UserWrapper
	Pagination *PaginationInfo
}

// Read

type ReadRequest struct {
	ID         string `validate:"omitempty"`
	Resolution PhotoResolution
}

func (r *ReadRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type ReadResponse struct {
	*UserWrapper
}

// ReadPublic

type ReadPublicRequest struct {
	ID         string `validate:"required"`
	Resolution PhotoResolution
}

func (r *ReadPublicRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type ReadPublicResponse struct {
	*UserWrapper
}

// Update

type UpdateRequest struct {
	ID     string `validate:"required"`
	Status string
	Role   string
}

func (r *UpdateRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type UpdateResponse struct {
	User *dbmodels.User
}

// UpdateSelf

type UpdateSelfRequest struct {
	FirstName string
	LastName  string
	Username  string
}

func (r *UpdateSelfRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

func (r *UpdateSelfRequest) normalize() error {
	r.Username = strings.ToLower(r.Username)
	return nil
}

type UpdateSelfResponse struct {
	User *dbmodels.User
}

// Register

type RegisterRequest struct {
	FirstName string `validate:"required"`
	LastName  string `validate:"required"`
	Username  string `validate:"required"`
	Password  string `validate:"required" mask:"true"`
}

func (r *RegisterRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}

	return nil
}

func (r *RegisterRequest) normalize() error {
	r.Username = strings.ToLower(r.Username)
	return nil
}

type RegisterResponse struct {
	User *dbmodels.User
}

// Delete

type DeleteRequest struct {
	ID string `validate:"required"`
}

func (r *DeleteRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type DeleteResponse struct{}

// DeleteSelf

type DeleteSelfRequest struct{}

func (r *DeleteSelfRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type DeleteSelfResponse struct{}

// UploadProfile

type UploadProfileRequest struct {
	Image []byte `validate:"required" mask:"true"`
}

func (r *UploadProfileRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type UploadProfileResponse struct {
	URL string
}

// Follow

type FollowRequest struct {
	FolloweeID string `validate:"required"`
}

func (r *FollowRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type FollowResponse struct{}

// Unfollow

type UnfollowRequest struct {
	FolloweeID string `validate:"required"`
}

func (r *UnfollowRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type UnfollowResponse struct{}

// AddDevice

type AddDeviceRequest struct {
	DeviceID  string
	Token     string
	Oem       string
	Device    string
	OS        string
	Longitude float32
	Latitude  float32
}

type AddDeviceResponse struct{}
