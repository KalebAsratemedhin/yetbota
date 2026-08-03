package user

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/constants"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/dbmodels"
	ctxYB "github.com/beka-birhanu/yetbota/identity-service/internal/domain/context"
	domainUser "github.com/beka-birhanu/yetbota/identity-service/internal/domain/user"
	userSvc "github.com/beka-birhanu/yetbota/identity-service/internal/services/usecase/user"
)

func decodeUserListHTTP(ctx context.Context, r *http.Request) (any, error) {
	q := r.URL.Query()

	opts := &domainUser.Options{
		FirstName: q.Get("first_name"),
		Surname:   q.Get("surname"),
		Username:  q.Get("username"),
		Status:    strings.ToUpper(strings.TrimSpace(q.Get("status"))),
		Role:      roleFromQuery(q.Get("role")),
		Search:    strings.TrimSpace(q.Get("q")),
	}
	if v := q.Get("min_rating"); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, badRequest("invalid min_rating", err)
		}
		opts.MinScore = &n
	}
	if v := q.Get("max_rating"); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, badRequest("invalid max_rating", err)
		}
		opts.MaxScore = &n
	}
	if v := q.Get("created_from"); v != "" {
		t, err := parseListDate(v)
		if err != nil {
			return nil, badRequest("invalid created_from", err)
		}
		opts.CreatedFrom = &t
	}
	if v := q.Get("created_to"); v != "" {
		t, err := parseListDate(v)
		if err != nil {
			return nil, badRequest("invalid created_to", err)
		}
		opts.CreatedTo = &t
	}

	pagination := &domainUser.Pagination{
		Limit: constants.DefaultPaginationLength,
	}
	if limitStr := q.Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			return nil, badRequest("invalid limit", err)
		}
		pagination.Limit = limit
	}
	if pageStr := q.Get("page"); pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			return nil, badRequest("invalid page", err)
		}
		pagination.Page = page
	}

	sort := &domainUser.SortOption{
		Field:     domainUser.SortFieldRating,
		Direction: domainUser.SortDirectionDesc,
	}
	if f := q.Get("sort_field"); f != "" {
		sort.Field = sortFieldFromQuery(f)
	}
	if d := q.Get("sort_direction"); d != "" {
		sort.Direction = sortDirectionFromQuery(d)
	}

	req := &userSvc.ListRequest{
		Options:    opts,
		Pagination: pagination,
		Sort:       sort,
	}
	if res := q.Get("resolution"); res != "" {
		req.Resolution = userSvc.PhotoResolution(res)
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}
	setCtxRequest(ctx, req)
	return req, nil
}

func roleFromQuery(v string) string {
	switch strings.ToUpper(strings.TrimSpace(v)) {
	case "USER":
		return dbmodels.RolesUSER
	case "ADMIN":
		return dbmodels.RolesADMIN
	default:
		return ""
	}
}

func sortFieldFromQuery(v string) domainUser.SortField {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "rating":
		return domainUser.SortFieldRating
	case "followers":
		return domainUser.SortFieldFollowers
	case "following":
		return domainUser.SortFieldFollowing
	case "contributions":
		return domainUser.SortFieldContributions
	case "created_at", "createdat":
		return domainUser.SortFieldCreatedAt
	case "updated_at", "updatedat":
		return domainUser.SortFieldUpdatedAt
	default:
		return ""
	}
}

func parseListDate(v string) (time.Time, error) {
	for _, layout := range []string{time.RFC3339, "2006-01-02"} {
		if t, err := time.Parse(layout, v); err == nil {
			return t, nil
		}
	}
	return time.Time{}, errors.New("expected ISO-8601 date (YYYY-MM-DD or RFC3339)")
}

func sortDirectionFromQuery(v string) domainUser.SortDirection {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "asc":
		return domainUser.SortDirectionAsc
	case "desc":
		return domainUser.SortDirectionDesc
	default:
		return ""
	}
}

func decodeUserReadMeHTTP(ctx context.Context, r *http.Request) (any, error) {
	req := &userSvc.ReadRequest{}
	if res := r.URL.Query().Get("resolution"); res != "" {
		req.Resolution = userSvc.PhotoResolution(res)
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	setCtxRequest(ctx, req)
	return req, nil
}

func decodeUserReadPublicHTTP(ctx context.Context, r *http.Request) (any, error) {
	id := r.PathValue("id")
	if id == "me" {
		return nil, badRequest(
			`use GET /v1/users/me with Authorization for the current user; "me" is not a user id`,
			errors.New("reserved path segment"),
		)
	}
	req := &userSvc.ReadPublicRequest{
		ID: id,
	}
	if res := r.URL.Query().Get("resolution"); res != "" {
		req.Resolution = userSvc.PhotoResolution(res)
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	setCtxRequest(ctx, req)
	return req, nil
}

func decodeUserRegisterHTTP(ctx context.Context, r *http.Request) (any, error) {
	var in struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Username  string `json:"username"`
		Password  string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		return nil, badRequest("invalid json", err)
	}
	req := &userSvc.RegisterRequest{
		FirstName: in.FirstName,
		LastName:  in.LastName,
		Username:  in.Username,
		Password:  in.Password,
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	setCtxRequest(ctx, req)
	return req, nil
}

func decodeUserUpdateHTTP(ctx context.Context, r *http.Request) (any, error) {
	var in struct {
		Status string `json:"status"`
		Role   string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		return nil, badRequest("invalid json", err)
	}
	req := &userSvc.UpdateRequest{
		ID:     r.PathValue("id"),
		Status: in.Status,
		Role:   in.Role,
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	setCtxRequest(ctx, req)
	return req, nil
}

func decodeUserUpdateSelfHTTP(ctx context.Context, r *http.Request) (any, error) {
	var in struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Username  string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		return nil, badRequest("invalid json", err)
	}
	req := &userSvc.UpdateSelfRequest{
		FirstName: in.FirstName,
		LastName:  in.LastName,
		Username:  in.Username,
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	setCtxRequest(ctx, req)
	return req, nil
}

func decodeUserDeleteHTTP(ctx context.Context, r *http.Request) (any, error) {
	req := &userSvc.DeleteRequest{ID: r.PathValue("id")}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	setCtxRequest(ctx, req)
	return req, nil
}

func decodeUserDeleteSelfHTTP(ctx context.Context, _ *http.Request) (any, error) {
	req := &userSvc.DeleteSelfRequest{}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	setCtxRequest(ctx, req)
	return req, nil
}

func decodeUserUploadProfileHTTP(ctx context.Context, r *http.Request) (any, error) {
	var in struct {
		ImageBase64 string `json:"image_base64"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		return nil, badRequest("invalid json", err)
	}
	b, err := base64.StdEncoding.DecodeString(in.ImageBase64)
	if err != nil {
		return nil, badRequest("invalid image_base64", err)
	}
	req := &userSvc.UploadProfileRequest{Image: b}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	setCtxRequest(ctx, req)
	return req, nil
}

func decodeUserFollowHTTP(ctx context.Context, r *http.Request) (any, error) {
	var in struct {
		FolloweeID string `json:"followee_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		return nil, badRequest("invalid json", err)
	}
	req := &userSvc.FollowRequest{FolloweeID: in.FolloweeID}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	setCtxRequest(ctx, req)
	return req, nil
}

func decodeUserUnfollowHTTP(ctx context.Context, r *http.Request) (any, error) {
	var in struct {
		FolloweeID string `json:"followee_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		return nil, badRequest("invalid json", err)
	}
	req := &userSvc.UnfollowRequest{FolloweeID: in.FolloweeID}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	setCtxRequest(ctx, req)
	return req, nil
}

func setCtxRequest(ctx context.Context, req any) {
	data := ctx.Value(ctxYB.AppSession)
	ctxSess, ok := data.(*ctxYB.Context)
	if !ok || ctxSess == nil {
		return
	}
	ctxSess.SetRequest(req)
}

func badRequest(publicMsg string, err error) error {
	return &toddlerr.Error{
		PublicStatusCode:  status.BadRequest,
		ServiceStatusCode: status.BadRequest,
		PublicMessage:     publicMsg,
		ServiceMessage:    err.Error(),
	}
}

