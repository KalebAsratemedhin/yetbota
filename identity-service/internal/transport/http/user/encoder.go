package user

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/dbmodels"
	ctxYB "github.com/beka-birhanu/yetbota/identity-service/internal/domain/context"
	userSvc "github.com/beka-birhanu/yetbota/identity-service/internal/services/usecase/user"
	"github.com/beka-birhanu/yetbota/identity-service/internal/transport/http/shared"
)

type userDTO struct {
	ID            string    `json:"id"`
	FirstName     string    `json:"first_name"`
	LastName      string    `json:"last_name"`
	Username      string    `json:"username"`
	Badges        []string  `json:"badges,omitempty"`
	Rating        int       `json:"rating"`
	Contributions int       `json:"contributions"`
	Status        string    `json:"status"`
	Followers     int       `json:"followers"`
	Following     int       `json:"following"`
	Role          string    `json:"role"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	ProfileURL    string    `json:"profile_url,omitempty"`
}

type publicUserDTO struct {
	ID            string    `json:"id"`
	FirstName     string    `json:"first_name"`
	LastName      string    `json:"last_name"`
	Username      string    `json:"username"`
	Badges        []string  `json:"badges,omitempty"`
	Rating        int       `json:"rating"`
	Contributions int       `json:"contributions"`
	Followers     int       `json:"followers"`
	Following     int       `json:"following"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	ProfileURL    string    `json:"profile_url,omitempty"`
}

type paginationDTO struct {
	Total       int64 `json:"total"`
	Limit       int   `json:"limit"`
	CurrentPage int   `json:"current_page"`
}

type listData struct {
	Users      []publicUserDTO `json:"users"`
	Pagination *paginationDTO  `json:"pagination,omitempty"`
}

func setCtxResponse(ctx context.Context, env shared.Envelope) {
	data := ctx.Value(ctxYB.AppSession)
	ctxSess, ok := data.(*ctxYB.Context)
	if !ok || ctxSess == nil {
		return
	}
	ctxSess.Response = env
}

func toUserDTO(u *dbmodels.User, profileURL string) userDTO {
	if u == nil {
		return userDTO{}
	}
	return userDTO{
		ID:            u.ID,
		FirstName:     u.FirstName,
		LastName:      u.LastName,
		Username:      u.Username,
		Badges:        []string(u.Badges),
		Rating:        int(u.Score),
		Contributions: u.Contributions,
		Status:        u.Status,
		Followers:     u.Followers,
		Following:     u.Following,
		Role:          u.Role,
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
		ProfileURL:    profileURL,
	}
}

func toPublicUserDTO(u *dbmodels.User, profileURL string) publicUserDTO {
	if u == nil {
		return publicUserDTO{}
	}
	return publicUserDTO{
		ID:            u.ID,
		FirstName:     u.FirstName,
		LastName:      u.LastName,
		Username:      u.Username,
		Badges:        []string(u.Badges),
		Rating:        int(u.Score),
		Contributions: u.Contributions,
		Followers:     u.Followers,
		Following:     u.Following,
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
		ProfileURL:    profileURL,
	}
}

func encodeUserListHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	out, ok := resp.(*userSvc.ListResponse)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(shared.Envelope{Success: false, Message: "something went wrong"})
	}

	users := make([]publicUserDTO, 0, len(out.Users))
	for _, uw := range out.Users {
		if uw == nil || uw.User == nil {
			continue
		}
		users = append(users, toPublicUserDTO(uw.User, uw.ProfileURL))
	}

	var pg *paginationDTO
	if out.Pagination != nil {
		pg = &paginationDTO{
			Total:       out.Pagination.Total,
			Limit:       out.Pagination.Limit,
			CurrentPage: out.Pagination.CurrentPage,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	env := shared.Envelope{
		Success: true,
		Data: listData{
			Users:      users,
			Pagination: pg,
		},
	}
	setCtxResponse(ctx, env)
	return json.NewEncoder(w).Encode(env)
}

type readPublicData struct {
	User userDTO `json:"user"`
}

type publicUserData struct {
	User publicUserDTO `json:"user"`
}

func encodeUserReadHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	out, ok := resp.(*userSvc.ReadResponse)
	if !ok || out == nil || out.UserWrapper == nil || out.User == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(shared.Envelope{Success: false, Message: "something went wrong"})
	}

	u := out.User

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	env := shared.Envelope{
		Success: true,
		Data: readPublicData{
			User: toUserDTO(u, out.ProfileURL),
		},
	}
	setCtxResponse(ctx, env)
	return json.NewEncoder(w).Encode(env)
}

func encodeUserReadPublicHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	out, ok := resp.(*userSvc.ReadPublicResponse)
	if !ok || out == nil || out.UserWrapper == nil || out.User == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(shared.Envelope{Success: false, Message: "something went wrong"})
	}

	u := out.User

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	env := shared.Envelope{
		Success: true,
		Data: publicUserData{
			User: toPublicUserDTO(u, out.ProfileURL),
		},
	}
	setCtxResponse(ctx, env)
	return json.NewEncoder(w).Encode(env)
}

type userData struct {
	User userDTO `json:"user"`
}

func encodeUserRegisterHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	out, ok := resp.(*userSvc.RegisterResponse)
	if !ok || out == nil || out.User == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(shared.Envelope{Success: false, Message: "something went wrong"})
	}

	env := shared.Envelope{Success: true, Data: userData{User: toUserDTO(out.User, "")}}
	setCtxResponse(ctx, env)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(env)
}

func encodeUserUpdateHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	out, ok := resp.(*userSvc.UpdateResponse)
	if !ok || out == nil || out.User == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(shared.Envelope{Success: false, Message: "something went wrong"})
	}

	env := shared.Envelope{Success: true, Data: userData{User: toUserDTO(out.User, "")}}
	setCtxResponse(ctx, env)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(env)
}

func encodeUserUpdateSelfHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	out, ok := resp.(*userSvc.UpdateSelfResponse)
	if !ok || out == nil || out.User == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(shared.Envelope{Success: false, Message: "something went wrong"})
	}

	env := shared.Envelope{Success: true, Data: userData{User: toUserDTO(out.User, "")}}
	setCtxResponse(ctx, env)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(env)
}

type uploadProfileData struct {
	URL string `json:"url"`
}

func encodeUserUploadProfileHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	out, ok := resp.(*userSvc.UploadProfileResponse)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(shared.Envelope{Success: false, Message: "something went wrong"})
	}

	env := shared.Envelope{Success: true, Data: uploadProfileData{URL: out.URL}}
	setCtxResponse(ctx, env)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(env)
}

func encodeUserDeleteHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	if _, ok := resp.(*userSvc.DeleteResponse); !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(shared.Envelope{Success: false, Message: "something went wrong"})
	}

	env := shared.Envelope{Success: true}
	setCtxResponse(ctx, env)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(env)
}

func encodeUserDeleteSelfHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	if _, ok := resp.(*userSvc.DeleteSelfResponse); !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(shared.Envelope{Success: false, Message: "something went wrong"})
	}

	env := shared.Envelope{Success: true}
	setCtxResponse(ctx, env)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(env)
}

func encodeUserFollowHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	if _, ok := resp.(*userSvc.FollowResponse); !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(shared.Envelope{Success: false, Message: "something went wrong"})
	}
	env := shared.Envelope{Success: true}
	setCtxResponse(ctx, env)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(env)
}

func encodeUserUnfollowHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	if _, ok := resp.(*userSvc.UnfollowResponse); !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(shared.Envelope{Success: false, Message: "something went wrong"})
	}
	env := shared.Envelope{Success: true}
	setCtxResponse(ctx, env)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(env)
}
