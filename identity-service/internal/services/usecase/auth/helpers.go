package auth

import (
	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
)

func invalidCredentialsError() error {
	return &toddlerr.Error{
		PublicStatusCode:  status.Unauthorized,
		ServiceStatusCode: status.Unauthorized,
		PublicMessage:     "Invalid username or password",
		ServiceMessage:    "invalid credentials",
	}
}
