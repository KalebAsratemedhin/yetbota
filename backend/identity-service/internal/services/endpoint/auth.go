package endpoint

import (
	"context"
	"io"

	"github.com/go-kit/kit/endpoint"
	"github.com/google/uuid"

	domain "github.com/yetbota/identity-service/internal/domain/user"
	authuc "github.com/yetbota/identity-service/internal/services/usecase/auth"
)

type AuthService interface {
	SignUp(ctx context.Context, in authuc.Credentials) (authuc.AuthResult, error)
	SignIn(ctx context.Context, in authuc.Credentials) (authuc.AuthResult, error)
	Me(ctx context.Context, userID uuid.UUID) (authuc.UserInfo, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, patch domain.ProfileUpdate) (authuc.UserInfo, error)
	UploadAvatar(ctx context.Context, userID uuid.UUID, filename string, r io.Reader, size int64) (authuc.UserInfo, error)
	DeleteAvatar(ctx context.Context, userID uuid.UUID) (authuc.UserInfo, error)
	Introspect(ctx context.Context, token string) (authuc.IntrospectResult, error)
	Logout(ctx context.Context, in authuc.LogoutInput) error
	ChangePassword(ctx context.Context, in authuc.ChangePasswordInput) error
	Refresh(ctx context.Context, refreshToken string) (authuc.AuthResult, error)
	DeleteAccount(ctx context.Context, claims domain.TokenClaims) error
	AuthenticateToken(ctx context.Context, token string) (domain.TokenClaims, error)
}

type CredentialsRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
}

type AuthResponse struct {
	UserID       string `json:"user_id"`
	Username     string `json:"username"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type UserResponse struct {
	UserID          string      `json:"user_id"`
	Username        string      `json:"username"`
	FullName        string      `json:"full_name"`
	Phone           string      `json:"phone"`
	Address         string      `json:"address"`
	Bio             string      `json:"bio"`
	AvatarURL       string      `json:"avatar_url"`
	ReputationScore int64       `json:"reputation_score"`
	ReputationLevel string      `json:"reputation_level"`
	Badges          []BadgeJSON `json:"badges,omitempty"`
}

type IntrospectRequest struct {
	Token string `json:"token"`
}

type IntrospectResponse struct {
	Active   bool   `json:"active"`
	UserID   string `json:"user_id,omitempty"`
	Username string `json:"username,omitempty"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type LogoutRequest struct {
	Claims       domain.TokenClaims
	RefreshToken string `json:"refresh_token"`
}

type ClaimsRequest struct {
	Claims domain.TokenClaims
}

type UpdateProfileRequest struct {
	Claims   domain.TokenClaims
	FullName *string `json:"full_name"`
	Phone    *string `json:"phone"`
	Address  *string `json:"address"`
	Bio      *string `json:"bio"`
}

func MakeSignUpEndpoint(svc AuthService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CredentialsRequest)
		res, err := svc.SignUp(ctx, authuc.Credentials{
			Username: req.Username,
			Password: req.Password,
			FullName: req.FullName,
		})
		if err != nil {
			return nil, err
		}
		return toAuthResponse(res), nil
	}
}

func MakeSignInEndpoint(svc AuthService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CredentialsRequest)
		res, err := svc.SignIn(ctx, authuc.Credentials{Username: req.Username, Password: req.Password})
		if err != nil {
			return nil, err
		}
		return toAuthResponse(res), nil
	}
}

func MakeMeEndpoint(svc AuthService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(ClaimsRequest)
		res, err := svc.Me(ctx, req.Claims.UserID)
		if err != nil {
			return nil, err
		}
		return toUserResponse(res), nil
	}
}

func MakeUpdateProfileEndpoint(svc AuthService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(UpdateProfileRequest)
		res, err := svc.UpdateProfile(ctx, req.Claims.UserID, domain.ProfileUpdate{
			FullName: req.FullName,
			Phone:    req.Phone,
			Address:  req.Address,
			Bio:      req.Bio,
		})
		if err != nil {
			return nil, err
		}
		return toUserResponse(res), nil
	}
}

type UploadAvatarRequest struct {
	Claims   domain.TokenClaims
	Filename string
	Reader   io.Reader
	Size     int64
}

func MakeUploadAvatarEndpoint(svc AuthService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(UploadAvatarRequest)
		res, err := svc.UploadAvatar(ctx, req.Claims.UserID, req.Filename, req.Reader, req.Size)
		if err != nil {
			return nil, err
		}
		return toUserResponse(res), nil
	}
}

func MakeDeleteAvatarEndpoint(svc AuthService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(ClaimsRequest)
		res, err := svc.DeleteAvatar(ctx, req.Claims.UserID)
		if err != nil {
			return nil, err
		}
		return toUserResponse(res), nil
	}
}

func MakeIntrospectEndpoint(svc AuthService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(IntrospectRequest)
		res, err := svc.Introspect(ctx, req.Token)
		if err != nil {
			return nil, err
		}
		return IntrospectResponse{Active: res.Active, UserID: res.UserID, Username: res.Username}, nil
	}
}

func MakeLogoutEndpoint(svc AuthService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(LogoutRequest)
		if err := svc.Logout(ctx, authuc.LogoutInput{Claims: req.Claims, RefreshToken: req.RefreshToken}); err != nil {
			return nil, err
		}
		return map[string]string{"status": "ok"}, nil
	}
}

func MakeChangePasswordEndpoint(svc AuthService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(struct {
			Claims domain.TokenClaims
			Body   ChangePasswordRequest
		})
		err := svc.ChangePassword(ctx, authuc.ChangePasswordInput{
			UserID:          req.Claims.UserID,
			CurrentPassword: req.Body.CurrentPassword,
			NewPassword:     req.Body.NewPassword,
		})
		if err != nil {
			return nil, err
		}
		return map[string]string{"status": "ok"}, nil
	}
}

func MakeRefreshEndpoint(svc AuthService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(RefreshRequest)
		res, err := svc.Refresh(ctx, req.RefreshToken)
		if err != nil {
			return nil, err
		}
		return toAuthResponse(res), nil
	}
}

func MakeDeleteAccountEndpoint(svc AuthService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(ClaimsRequest)
		if err := svc.DeleteAccount(ctx, req.Claims); err != nil {
			return nil, err
		}
		return map[string]string{"status": "ok"}, nil
	}
}

func toAuthResponse(res authuc.AuthResult) AuthResponse {
	return AuthResponse{
		UserID:       res.UserID,
		Username:     res.Username,
		AccessToken:  res.AccessToken,
		RefreshToken: res.RefreshToken,
	}
}

func toUserResponse(res authuc.UserInfo) UserResponse {
	badges := make([]BadgeJSON, 0, len(res.Badges))
	for _, b := range res.Badges {
		badges = append(badges, BadgeJSON{
			ID: b.ID, Name: b.Name, Description: b.Description, AwardedAt: b.AwardedAt,
		})
	}
	return UserResponse{
		UserID: res.UserID, Username: res.Username, FullName: res.FullName,
		Phone: res.Phone, Address: res.Address, Bio: res.Bio, AvatarURL: res.AvatarURL,
		ReputationScore: res.ReputationScore, ReputationLevel: res.ReputationLevel, Badges: badges,
	}
}
