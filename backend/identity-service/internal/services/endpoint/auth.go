package endpoint

import (
	"context"

	"github.com/go-kit/kit/endpoint"

	authuc "github.com/yetbota/identity-service/internal/services/usecase/auth"
)

type AuthService interface {
	SignUp(ctx context.Context, in authuc.Credentials) (authuc.AuthResult, error)
	SignIn(ctx context.Context, in authuc.Credentials) (authuc.AuthResult, error)
}

type CredentialsRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Token    string `json:"token"`
}

func MakeSignUpEndpoint(svc AuthService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CredentialsRequest)
		res, err := svc.SignUp(ctx, authuc.Credentials{
			Username: req.Username,
			Password: req.Password,
		})
		if err != nil {
			return nil, err
		}
		return AuthResponse{UserID: res.UserID, Username: res.Username, Token: res.Token}, nil
	}
}

func MakeSignInEndpoint(svc AuthService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CredentialsRequest)
		res, err := svc.SignIn(ctx, authuc.Credentials{
			Username: req.Username,
			Password: req.Password,
		})
		if err != nil {
			return nil, err
		}
		return AuthResponse{UserID: res.UserID, Username: res.Username, Token: res.Token}, nil
	}
}
