package endpoint

import (
	"github.com/go-kit/kit/endpoint"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/validator"
	authSvc "github.com/beka-birhanu/yetbota/identity-service/internal/services/usecase/auth"
	deviceSvc "github.com/beka-birhanu/yetbota/identity-service/internal/services/usecase/userdevice"
	userSvc "github.com/beka-birhanu/yetbota/identity-service/internal/services/usecase/user"
)

type Endpoints struct {
	Login          endpoint.Endpoint
	Refresh        endpoint.Endpoint
	Logout         endpoint.Endpoint
	Authorization  endpoint.Endpoint
	ChangePassword endpoint.Endpoint

	UserList          endpoint.Endpoint
	UserRead          endpoint.Endpoint
	UserReadPublic    endpoint.Endpoint
	UserUpdate        endpoint.Endpoint
	UserUpdateSelf    endpoint.Endpoint
	UserRegister      endpoint.Endpoint
	UserDelete        endpoint.Endpoint
	UserDeleteSelf    endpoint.Endpoint
	UserUploadProfile endpoint.Endpoint
	UserFollow        endpoint.Endpoint
	UserUnfollow      endpoint.Endpoint

	UserDeviceRegister endpoint.Endpoint
}

type Config struct {
	AuthService   authSvc.Service   `validate:"required"`
	UserService   userSvc.Service   `validate:"required"`
	DeviceService deviceSvc.Service `validate:"required"`
}

func (c *Config) Validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

func NewEndpoints(c *Config) (*Endpoints, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	return &Endpoints{
		Login:          makeLoginEndpoint(c.AuthService),
		Refresh:        makeRefreshEndpoint(c.AuthService),
		Logout:         makeLogoutEndpoint(c.AuthService),
		Authorization:  makeAuthorizationEndpoint(c.AuthService),
		ChangePassword: makeChangePasswordEndpoint(c.AuthService),

		UserList:          makeUserListEndpoint(c.UserService),
		UserRead:          makeUserReadEndpoint(c.UserService),
		UserReadPublic:    makeUserReadPublicEndpoint(c.UserService),
		UserUpdate:        makeUserUpdateEndpoint(c.UserService),
		UserUpdateSelf:    makeUserUpdateSelfEndpoint(c.UserService),
		UserRegister:      makeUserRegisterEndpoint(c.UserService),
		UserDelete:        makeUserDeleteEndpoint(c.UserService),
		UserDeleteSelf:    makeUserDeleteSelfEndpoint(c.UserService),
		UserUploadProfile: makeUserUploadProfileEndpoint(c.UserService),
		UserFollow:        makeUserFollowEndpoint(c.UserService),
		UserUnfollow:      makeUserUnfollowEndpoint(c.UserService),

		UserDeviceRegister: makeUserDeviceRegisterEndpoint(c.DeviceService),
	}, nil
}
