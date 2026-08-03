package place

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	ctxYB "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	contributionSvc "github.com/beka-birhanu/yetbota/content-service/internal/services/usecase/contribution"
	placeSvc "github.com/beka-birhanu/yetbota/content-service/internal/services/usecase/place"
)

func decodePlaceAddHTTP(ctx context.Context, r *http.Request) (any, error) {
	var in struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Address     string `json:"address"`
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
	photos := make([]*placeSvc.OrderedPhotoUpload, 0, len(in.Photos))
	for _, p := range in.Photos {
		b, err := base64.StdEncoding.DecodeString(strings.TrimSpace(p.PhotoBase64))
		if err != nil {
			return nil, badRequest("invalid photo_base64", err)
		}
		photos = append(photos, &placeSvc.OrderedPhotoUpload{Photo: b, Position: p.Position})
	}
	var lat, lon float64
	if in.Location != nil {
		lat = in.Location.Latitude
		lon = in.Location.Longitude
	}
	req := &placeSvc.AddRequest{
		Name:        in.Name,
		Description: in.Description,
		Address:     in.Address,
		Photos:      photos,
		Latitude:    lat,
		Longitude:   lon,
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	setCtxRequest(ctx, req)
	return req, nil
}

func decodePlaceReadHTTP(ctx context.Context, r *http.Request) (any, error) {
	resolution := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("resolution")))
	if resolution == "" {
		resolution = string(placeSvc.PhotoResolutionOriginal)
	}
	req := &placeSvc.ReadRequest{
		ID:              r.PathValue("id"),
		PhotoResolution: placeSvc.PhotoResolution(resolution),
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	setCtxRequest(ctx, req)
	return req, nil
}

func decodePlaceListHTTP(ctx context.Context, r *http.Request) (any, error) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	pageSize, _ := strconv.Atoi(q.Get("page_size"))
	opts := placeSvc.ListOptions{
		Search:   q.Get("search"),
		Page:     page,
		PageSize: pageSize,
		Status:   q.Get("status"),
	}
	if v := strings.TrimSpace(q.Get("needs_check")); v != "" {
		switch strings.ToLower(v) {
		case "true", "1", "yes":
			t := true
			opts.NeedsCheck = &t
		case "false", "0", "no":
			f := false
			opts.NeedsCheck = &f
		default:
			return nil, badRequest("invalid needs_check", fmt.Errorf("invalid needs_check %q", v))
		}
	}
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
	req := &placeSvc.ListRequest{
		ListOptions:     opts,
		PhotoResolution: placeSvc.PhotoResolution(resolution),
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	setCtxRequest(ctx, req)
	return req, nil
}

func decodePlaceUpdateHTTP(ctx context.Context, r *http.Request) (any, error) {
	var in struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		Address     *string `json:"address"`
		Status      *string `json:"status"`
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
	photos := make([]*placeSvc.OrderedPhotoUpload, 0, len(in.UpsertPhotos))
	for _, p := range in.UpsertPhotos {
		b, err := base64.StdEncoding.DecodeString(strings.TrimSpace(p.PhotoBase64))
		if err != nil {
			return nil, badRequest("invalid photo_base64", err)
		}
		photos = append(photos, &placeSvc.OrderedPhotoUpload{Photo: b, Position: p.Position})
	}
	var lat, lon *float64
	if in.Location != nil {
		lat = &in.Location.Latitude
		lon = &in.Location.Longitude
	}
	req := &placeSvc.UpdateRequest{
		ID:           r.PathValue("id"),
		Name:         in.Name,
		Description:  in.Description,
		Address:      in.Address,
		UpsertPhotos: photos,
		Latitude:     lat,
		Longitude:    lon,
		Status:       in.Status,
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	setCtxRequest(ctx, req)
	return req, nil
}

func decodePlaceContributionAddHTTP(ctx context.Context, r *http.Request) (any, error) {
	var in struct {
		Kind    string `json:"kind"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		return nil, badRequest("invalid json", err)
	}
	req := &contributionSvc.AddRequest{
		PlaceID:  r.PathValue("id"),
		Kind:    strings.ToLower(strings.TrimSpace(in.Kind)),
		Content: in.Content,
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	setCtxRequest(ctx, req)
	return req, nil
}

func decodePlaceContributionListHTTP(ctx context.Context, r *http.Request) (any, error) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	pageSize, _ := strconv.Atoi(q.Get("page_size"))
	req := &contributionSvc.ListRequest{
		PlaceID:  r.PathValue("id"),
		Kind:     strings.ToLower(strings.TrimSpace(q.Get("kind"))),
		Page:     page,
		PageSize: pageSize,
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	setCtxRequest(ctx, req)
	return req, nil
}

func decodePlaceBookmarkHTTP(ctx context.Context, r *http.Request) (any, error) {
	req := &placeSvc.BookmarkRequest{PlaceID: r.PathValue("id")}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	setCtxRequest(ctx, req)
	return req, nil
}

func decodePlaceUnbookmarkHTTP(ctx context.Context, r *http.Request) (any, error) {
	req := &placeSvc.UnbookmarkRequest{PlaceID: r.PathValue("id")}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	setCtxRequest(ctx, req)
	return req, nil
}

func decodePlaceConfirmHTTP(ctx context.Context, r *http.Request) (any, error) {
	var in struct {
		Note string `json:"note"`
	}
	if r.Body != nil && r.ContentLength != 0 {
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			return nil, badRequest("invalid json", err)
		}
	}
	req := &placeSvc.ConfirmRequest{PlaceID: r.PathValue("id"), Note: in.Note}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	setCtxRequest(ctx, req)
	return req, nil
}

func decodePlaceReportProblemHTTP(ctx context.Context, r *http.Request) (any, error) {
	var in struct {
		Kind        string `json:"kind"`
		Note        string `json:"note"`
		DuplicateOf string `json:"duplicate_of"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		return nil, badRequest("invalid json", err)
	}
	req := &placeSvc.ReportProblemRequest{
		PlaceID:     r.PathValue("id"),
		Kind:        in.Kind,
		Note:        in.Note,
		DuplicateOf: in.DuplicateOf,
	}
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
