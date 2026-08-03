package post

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	ctxYB "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	domainPost "github.com/beka-birhanu/yetbota/content-service/internal/domain/post"
	postSvc "github.com/beka-birhanu/yetbota/content-service/internal/services/usecase/post"
)

func decodePostAddHTTP(ctx context.Context, r *http.Request) (any, error) {
	var in struct {
		Title       string   `json:"title"`
		Description string   `json:"description"`
		Tags        []string `json:"tags"`
		IsQuestion  bool     `json:"is_question"`
		Address     string   `json:"address"`
		AttachedPostID string `json:"attached_post_id"`
		Photos      []struct {
			PhotoBase64 string `json:"photo_base64"`
			Position    int    `json:"position"`
		} `json:"photos"`
		Location *struct {
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		} `json:"location"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		return nil, badRequest("invalid json", err)
	}

	photos := make([]*postSvc.OrderedPhotoUpload, 0, len(in.Photos))
	for _, p := range in.Photos {
		b, err := base64.StdEncoding.DecodeString(strings.TrimSpace(p.PhotoBase64))
		if err != nil {
			return nil, badRequest("invalid photo_base64", err)
		}
		photos = append(photos, &postSvc.OrderedPhotoUpload{
			Photo:    b,
			Position: p.Position,
		})
	}

	var lat, lon float64
	if in.Location != nil {
		lat = in.Location.Latitude
		lon = in.Location.Longitude
	}

	req := &postSvc.AddRequest{
		Title:       in.Title,
		Description: in.Description,
		Tags:        in.Tags,
		IsQuestion:  in.IsQuestion,
		Photos:      photos,
		Latitude:    lat,
		Longitude:   lon,
		Address:     in.Address,
		AttachedPostID: strings.TrimSpace(in.AttachedPostID),
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	setCtxRequest(ctx, req)
	return req, nil
}

func decodePostReadHTTP(ctx context.Context, r *http.Request) (any, error) {
	resolution := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("resolution")))
	if resolution == "" {
		resolution = string(postSvc.PhotoResolutionOriginal)
	}

	req := &postSvc.ReadRequest{
		ID:              r.PathValue("id"),
		PhotoResolution: postSvc.PhotoResolution(resolution),
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	setCtxRequest(ctx, req)
	return req, nil
}

func decodePostUpdateHTTP(ctx context.Context, r *http.Request) (any, error) {
	var in struct {
		Title        string   `json:"title"`
		Description  string   `json:"description"`
		Tags         []string `json:"tags"`
		Address      string   `json:"address"`
		AttachedPostID *string `json:"attached_post_id"`
		UpsertPhotos []struct {
			PhotoBase64 string `json:"photo_base64"`
			Position    int    `json:"position"`
		} `json:"upsert_photos"`
		Location *struct {
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		} `json:"location"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		return nil, badRequest("invalid json", err)
	}

	photos := make([]*postSvc.OrderedPhotoUpload, 0, len(in.UpsertPhotos))
	for _, p := range in.UpsertPhotos {
		b, err := base64.StdEncoding.DecodeString(strings.TrimSpace(p.PhotoBase64))
		if err != nil {
			return nil, badRequest("invalid photo_base64", err)
		}
		photos = append(photos, &postSvc.OrderedPhotoUpload{
			Photo:    b,
			Position: p.Position,
		})
	}

	var lat, lon float64
	if in.Location != nil {
		lat = in.Location.Latitude
		lon = in.Location.Longitude
	}

	var attachedPostID *string
	if in.AttachedPostID != nil {
		trimmed := strings.TrimSpace(*in.AttachedPostID)
		attachedPostID = &trimmed
	}

	req := &postSvc.UpdateRequest{
		ID:           r.PathValue("id"),
		Title:        in.Title,
		Description:  in.Description,
		Tags:         in.Tags,
		UpsertPhotos: photos,
		Latitude:     lat,
		Longitude:    lon,
		Address:      in.Address,
		AttachedPostID: attachedPostID,
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	setCtxRequest(ctx, req)
	return req, nil
}

func decodePostVoteHTTP(ctx context.Context, r *http.Request) (any, error) {
	var in struct {
		VoteType string `json:"vote_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		return nil, badRequest("invalid json", err)
	}

	req := &postSvc.PostVoteRequest{
		PostID:   r.PathValue("id"),
		VoteType: strings.ToLower(strings.TrimSpace(in.VoteType)),
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	setCtxRequest(ctx, req)
	return req, nil
}

func decodePostSaveHTTP(ctx context.Context, r *http.Request) (any, error) {
	req := &postSvc.SaveRequest{
		PostID: r.PathValue("id"),
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	setCtxRequest(ctx, req)
	return req, nil
}

func decodePostUnsaveHTTP(ctx context.Context, r *http.Request) (any, error) {
	req := &postSvc.UnsaveRequest{
		PostID: r.PathValue("id"),
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	setCtxRequest(ctx, req)
	return req, nil
}

func decodePostListHTTP(ctx context.Context, r *http.Request) (any, error) {
	q := r.URL.Query()

	page, _ := strconv.Atoi(q.Get("page"))
	pageSize, _ := strconv.Atoi(q.Get("page_size"))

	opts := postSvc.ListOptions{
		UserID:    q.Get("user_id"),
		Tags:      q["tags"],
		Search:    q.Get("search"),
		SortField: mapSortFieldHTTP(q.Get("sort_by")),
		SortDir:   mapSortDirHTTP(q.Get("sort_dir")),
		Page:      page,
		PageSize:  pageSize,
	}

	if v := q.Get("is_question"); v != "" {
		b := strings.EqualFold(v, "true")
		opts.IsQuestion = &b
	}

	if strings.EqualFold(q.Get("saved"), "true") {
		opts.SavedOnly = true
	}

	if strings.EqualFold(q.Get("non_visible"), "true") {
		opts.NonVisible = true
	}

	opts.AttachedPostID = q.Get("attached_post_id")

	if latStr := q.Get("near_lat"); latStr != "" {
		lat, err := strconv.ParseFloat(latStr, 64)
		if err != nil {
			return nil, badRequest("invalid near_lat", err)
		}
		lon, err := strconv.ParseFloat(q.Get("near_lon"), 64)
		if err != nil {
			return nil, badRequest("invalid near_lon", err)
		}
		km, err := strconv.ParseFloat(q.Get("radius_km"), 64)
		if err != nil {
			return nil, badRequest("invalid radius_km", err)
		}
		opts.NearLat = &lat
		opts.NearLon = &lon
		opts.RadiusKm = &km
	}

	resolution := strings.ToUpper(strings.TrimSpace(q.Get("resolution")))

	req := &postSvc.ListRequest{
		ListOptions:     opts,
		PhotoResolution: postSvc.PhotoResolution(resolution),
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	setCtxRequest(ctx, req)
	return req, nil
}

func mapSortFieldHTTP(s string) domainPost.ListSortField {
	switch strings.ToLower(s) {
	case "likes":
		return domainPost.ListSortFieldLikes
	case "dislikes":
		return domainPost.ListSortFieldDislikes
	case "comments":
		return domainPost.ListSortFieldComments
	default:
		return domainPost.ListSortFieldCreatedAt
	}
}

func mapSortDirHTTP(s string) domainPost.ListSortDir {
	if strings.EqualFold(s, "asc") {
		return domainPost.ListSortDirAsc
	}
	return domainPost.ListSortDirDesc
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
