package endpoint

import (
	"context"
	"strconv"
	"strings"

	"github.com/go-kit/kit/endpoint"
	"github.com/google/uuid"

	"github.com/yetbota/identity-service/internal/domain/apperr"
	domain "github.com/yetbota/identity-service/internal/domain/user"
	socialuc "github.com/yetbota/identity-service/internal/services/usecase/social"
)

type SocialService interface {
	GetPublicProfile(ctx context.Context, userID uuid.UUID) (socialuc.PublicProfile, error)
	Search(ctx context.Context, viewerID uuid.UUID, query string) ([]socialuc.PublicProfile, error)
	Follow(ctx context.Context, followerID, followeeID uuid.UUID) error
	Unfollow(ctx context.Context, followerID, followeeID uuid.UUID) error
	Followers(ctx context.Context, userID uuid.UUID, in socialuc.ListInput) (socialuc.ProfilePage, error)
	Following(ctx context.Context, userID uuid.UUID, in socialuc.ListInput) (socialuc.ProfilePage, error)
}

type UserIDRequest struct {
	Claims domain.TokenClaims
	UserID uuid.UUID
}

type SearchRequest struct {
	Claims domain.TokenClaims
	Query  string
}

type UserListRequest struct {
	Claims domain.TokenClaims
	UserID uuid.UUID
	Limit  int
	Cursor string
}

type UsersResponse struct {
	Users []UserResponse `json:"users"`
}

type UsersPageResponse struct {
	Users      []UserResponse `json:"users"`
	NextCursor string         `json:"next_cursor,omitempty"`
	Total      int64          `json:"total"`
}

func toPublicUserResponse(res socialuc.PublicProfile) UserResponse {
	return UserResponse{
		UserID:    res.UserID,
		Username:  res.Username,
		FullName:  res.FullName,
		Phone:     res.Phone,
		Address:   res.Address,
		Bio:       res.Bio,
		AvatarURL: res.AvatarURL,
	}
}

func toUsersPageResponse(page socialuc.ProfilePage) UsersPageResponse {
	users := make([]UserResponse, 0, len(page.Users))
	for _, u := range page.Users {
		users = append(users, toPublicUserResponse(u))
	}
	return UsersPageResponse{
		Users:      users,
		NextCursor: page.NextCursor,
		Total:      page.Total,
	}
}

func MakeGetPublicProfileEndpoint(svc SocialService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(UserIDRequest)
		res, err := svc.GetPublicProfile(ctx, req.UserID)
		if err != nil {
			return nil, err
		}
		return toPublicUserResponse(res), nil
	}
}

func MakeSearchUsersEndpoint(svc SocialService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(SearchRequest)
		rows, err := svc.Search(ctx, req.Claims.UserID, req.Query)
		if err != nil {
			return nil, err
		}
		users := make([]UserResponse, 0, len(rows))
		for _, row := range rows {
			users = append(users, toPublicUserResponse(row))
		}
		return UsersResponse{Users: users}, nil
	}
}

func MakeFollowEndpoint(svc SocialService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(UserIDRequest)
		if err := svc.Follow(ctx, req.Claims.UserID, req.UserID); err != nil {
			return nil, err
		}
		return map[string]string{"status": "ok"}, nil
	}
}

func MakeUnfollowEndpoint(svc SocialService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(UserIDRequest)
		if err := svc.Unfollow(ctx, req.Claims.UserID, req.UserID); err != nil {
			return nil, err
		}
		return map[string]string{"status": "ok"}, nil
	}
}

func MakeFollowersEndpoint(svc SocialService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(UserListRequest)
		page, err := svc.Followers(ctx, req.UserID, socialuc.ListInput{
			Limit: req.Limit, Cursor: req.Cursor,
		})
		if err != nil {
			return nil, err
		}
		return toUsersPageResponse(page), nil
	}
}

func MakeFollowingEndpoint(svc SocialService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(UserListRequest)
		page, err := svc.Following(ctx, req.UserID, socialuc.ListInput{
			Limit: req.Limit, Cursor: req.Cursor,
		})
		if err != nil {
			return nil, err
		}
		return toUsersPageResponse(page), nil
	}
}

func ParseUserID(raw string) (uuid.UUID, error) {
	id, err := uuid.Parse(strings.TrimSpace(raw))
	if err != nil {
		return uuid.Nil, apperr.Validation(apperr.Field{
			Field: "id", Code: "INVALID", Message: "id must be a UUID",
		})
	}
	return id, nil
}

func ParseListLimit(raw string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return 0, apperr.Validation(apperr.Field{
			Field: "limit", Code: "INVALID", Message: "limit must be a non-negative integer",
		})
	}
	return n, nil
}
