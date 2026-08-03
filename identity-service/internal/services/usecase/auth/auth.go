package auth

import (
	"context"

	"github.com/aarondl/sqlboiler/v4/boil"
	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/dbmodels"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/utils"
	domainAuth "github.com/beka-birhanu/yetbota/identity-service/internal/domain/auth"
	contextYB "github.com/beka-birhanu/yetbota/identity-service/internal/domain/context"
	domainDevice "github.com/beka-birhanu/yetbota/identity-service/internal/domain/userdevice"
)

func (s *service) Login(ctx context.Context, ctxSess *contextYB.Context, req *LoginRequest) (*LoginResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err := req.normalize(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	user, err := s.userRepo.ReadByUsername(ctx, req.Username, nil)
	if err != nil {
		err := invalidCredentialsError()
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if user.Status != dbmodels.UserStatusACTIVE {
		err := &toddlerr.Error{
			PublicStatusCode:  status.Unauthorized,
			ServiceStatusCode: status.Unauthorized,
			PublicMessage:     "Invalid username or password",
			ServiceMessage:    "user status is not active",
		}
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err := s.hasher.Verify(user.Password, req.Password); err != nil {
		err := invalidCredentialsError()
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	td, err := s.sessionManager.NewSessionDetails(ctx, &domainAuth.SessionInfo{
		Username:   user.Username,
		UserID:     user.ID,
		Role:       user.Role,
		RefreshTTL: s.refreshTTL,
		AccessTTL:  s.accessTTL,
	})
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err := s.sessionManager.SaveSession(ctx, td); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	return &LoginResponse{
		AccessToken:     td.AccessToken,
		AccessTokenTTL:  int64(td.AccessTtl.Seconds()),
		RefreshToken:    td.RefreshToken,
		RefreshTokenTTL: int64(td.RefreshTtl.Seconds()),
	}, nil
}

func (s *service) Refresh(ctx context.Context, ctxSess *contextYB.Context, req *RefreshRequest) (*RefreshResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	userSess, err := s.sessionManager.ExtractUserSession(ctx, &domainAuth.TokenInfo{
		TokenType: domainAuth.RefreshToken,
		Token:     req.RefreshToken,
	})
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if userSess.Username != req.Username {
		err := &toddlerr.Error{
			PublicStatusCode:  status.Unauthorized,
			ServiceStatusCode: status.Unauthorized,
			PublicMessage:     "Invalid refresh token",
			ServiceMessage:    "username mismatch in refresh token",
		}
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	deleted, err := s.sessionManager.DeleteSession(ctx, userSess)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	if deleted == 0 {
		err := &toddlerr.Error{
			PublicStatusCode:  status.Unauthorized,
			ServiceStatusCode: status.Unauthorized,
			PublicMessage:     "Session expired",
			ServiceMessage:    "refresh session not found in redis",
		}
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	user, err := s.userRepo.ReadByUsername(ctx, req.Username, nil)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if user.Status != dbmodels.UserStatusACTIVE {
		err := &toddlerr.Error{
			PublicStatusCode:  status.Unauthorized,
			ServiceStatusCode: status.Unauthorized,
			PublicMessage:     "Session expired",
			ServiceMessage:    "user status is not active",
		}
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	td, err := s.sessionManager.NewSessionDetails(ctx, &domainAuth.SessionInfo{
		Username:   user.Username,
		UserID:     user.ID,
		Role:       user.Role,
		RefreshTTL: s.refreshTTL,
		AccessTTL:  s.accessTTL,
	})
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err := s.sessionManager.SaveSession(ctx, td); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	return &RefreshResponse{
		AccessToken:     td.AccessToken,
		AccessTokenTTL:  int64(td.AccessTtl.Seconds()),
		RefreshToken:    td.RefreshToken,
		RefreshTokenTTL: int64(td.RefreshTtl.Seconds()),
	}, nil
}

func (s *service) Logout(ctx context.Context, ctxSess *contextYB.Context, req *LogoutRequest) (*LogoutResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	userSess, err := s.sessionManager.ExtractUserSession(ctx, &domainAuth.TokenInfo{
		TokenType: domainAuth.RefreshToken,
		Token:     req.RefreshToken,
	})
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if _, err := s.sessionManager.DeleteSession(ctx, userSess); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if req.DeviceID != "" && s.deviceRepo != nil {
		s.removeDevice(ctx, userSess.UserID, req.DeviceID)
	}

	return &LogoutResponse{}, nil
}

func (s *service) Authorization(ctx context.Context, ctxSess *contextYB.Context, req *AuthorizationRequest) (*AuthorizationResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	return &AuthorizationResponse{}, nil
}

func (s *service) ChangePassword(ctx context.Context, ctxSess *contextYB.Context, req *ChangePasswordRequest) (*ChangePasswordResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err := utils.AllowAccess(ctxSess); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	user, err := s.userRepo.Read(ctx, ctxSess.UserSession.UserID, nil)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err := s.hasher.Verify(user.Password, req.CurrentPassword); err != nil {
		err := &toddlerr.Error{
			PublicStatusCode:  status.BadRequest,
			ServiceStatusCode: status.BadRequest,
			PublicMessage:     "Current password is incorrect",
			ServiceMessage:    "current password hash mismatch",
		}
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	hashed, err := s.hasher.Hash(req.NewPassword)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	user.Password = hashed
	if err := s.userRepo.Update(ctx, nil, user, boil.Whitelist(dbmodels.UserColumns.Password)); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	return &ChangePasswordResponse{}, nil
}

func (s *service) removeDevice(ctx context.Context, userID, deviceID string) {
	opts := &domainDevice.Options{
		UserIDs:   []string{userID},
		DeviceIDs: []string{deviceID},
	}
	devices, err := s.deviceRepo.List(ctx, opts, nil, nil)
	if err != nil {
		return
	}
	for _, d := range devices {
		_ = s.deviceRepo.Delete(ctx, nil, d.ID)
	}
}
