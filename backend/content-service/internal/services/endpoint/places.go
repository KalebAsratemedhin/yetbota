package endpoint

import (
	"context"
	"strings"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/google/uuid"

	"github.com/yetbota/content-service/internal/domain/apperr"
	"github.com/yetbota/content-service/internal/domain/auth"
	domain "github.com/yetbota/content-service/internal/domain/place"
	placeuc "github.com/yetbota/content-service/internal/services/usecase/place"
)

type PlaceService interface {
	Get(ctx context.Context, id uuid.UUID) (domain.Place, error)
	GetByPostID(ctx context.Context, postID uuid.UUID) (domain.Place, error)
	List(ctx context.Context, in placeuc.ListInput) (placeuc.PlacePageResult, error)
	ListNearbyInternal(ctx context.Context, lat, lng, radiusM float64, limit int) ([]domain.Place, error)
	ListPosts(ctx context.Context, placeID uuid.UUID, limit int, cursor string) (placeuc.PostPageResult, error)
	ListEntries(ctx context.Context, placeID uuid.UUID, category string, limit int, cursor string) (placeuc.EntryPageResult, error)
	AttachExisting(ctx context.Context, in placeuc.AttachToExistingInput) (domain.Place, error)
	CreateAndAttach(ctx context.Context, in placeuc.CreatePlaceAttachInput) (domain.Place, error)
	UpdateSummary(ctx context.Context, placeID uuid.UUID, summary, model string) (domain.Place, error)
	AddEntries(ctx context.Context, in placeuc.AddEntriesInput) ([]domain.Entry, error)
}

type ListPlacesRequest struct {
	Principal auth.Principal
	Query     string
	Lat       *float64
	Lng       *float64
	RadiusM   float64
	Limit     int
	Cursor    string
}

type PlaceIDRequest struct {
	Principal auth.Principal
	PlaceID   uuid.UUID
	Limit     int
	Cursor    string
	Category  string
}

type InternalNearbyRequest struct {
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
	RadiusM float64 `json:"radius_m"`
	Limit   int     `json:"limit"`
}

type InternalAttachRequest struct {
	PlaceID      *string   `json:"place_id"`
	PostID       string    `json:"post_id"`
	Name         string    `json:"name"`
	Lat          float64   `json:"lat"`
	Lng          float64   `json:"lng"`
	AttachScore  float64   `json:"attach_score"`
	AttachMethod string    `json:"attach_method"`
	Embedding    []float64 `json:"embedding"`
}

type InternalSummaryRequest struct {
	PlaceID uuid.UUID
	Summary string `json:"summary"`
	Model   string `json:"model"`
}

type InternalEntriesRequest struct {
	PlaceID uuid.UUID
	Entries []struct {
		Category     string  `json:"category"`
		Body         string  `json:"body"`
		Source       string  `json:"source"`
		SourcePostID *string `json:"source_post_id"`
	} `json:"entries"`
}

type PlaceJSON struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Lat              float64   `json:"lat"`
	Lng              float64   `json:"lng"`
	Summary          string    `json:"summary"`
	SummaryModel     string    `json:"summary_model,omitempty"`
	SummaryUpdatedAt *string   `json:"summary_updated_at,omitempty"`
	PostCount        int64     `json:"post_count"`
	Embedding        []float64 `json:"embedding,omitempty"`
	CreatedAt        string    `json:"created_at"`
	UpdatedAt        string    `json:"updated_at"`
}

type PlacesPageJSON struct {
	Places     []PlaceJSON `json:"places"`
	NextCursor string      `json:"next_cursor,omitempty"`
}

type EntryJSON struct {
	ID           string  `json:"id"`
	PlaceID      string  `json:"place_id"`
	Category     string  `json:"category"`
	Body         string  `json:"body"`
	Source       string  `json:"source"`
	SourcePostID *string `json:"source_post_id,omitempty"`
	AuthorID     *string `json:"author_id,omitempty"`
	CreatedAt    string  `json:"created_at"`
}

type EntriesPageJSON struct {
	Entries    []EntryJSON `json:"entries"`
	NextCursor string      `json:"next_cursor,omitempty"`
}

func MakeListPlacesEndpoint(svc PlaceService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(ListPlacesRequest)
		page, err := svc.List(ctx, placeuc.ListInput{
			Query: req.Query, Limit: req.Limit, Cursor: req.Cursor,
			Lat: req.Lat, Lng: req.Lng, Radius: req.RadiusM,
		})
		if err != nil {
			return nil, err
		}
		return PlacesPageJSON{Places: toPlacesJSON(page.Places, false), NextCursor: page.NextCursor}, nil
	}
}

func MakeGetPlaceEndpoint(svc PlaceService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(PlaceIDRequest)
		p, err := svc.Get(ctx, req.PlaceID)
		if err != nil {
			return nil, err
		}
		return toPlaceJSON(p, false), nil
	}
}

func MakeListPlacePostsEndpoint(svc PlaceService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(PlaceIDRequest)
		page, err := svc.ListPosts(ctx, req.PlaceID, req.Limit, req.Cursor)
		if err != nil {
			return nil, err
		}
		posts := make([]PostJSON, 0, len(page.Posts))
		for _, p := range page.Posts {
			posts = append(posts, toPostJSON(p))
		}
		return PostsPageJSON{Posts: posts, NextCursor: page.NextCursor}, nil
	}
}

func MakeListPlaceEntriesEndpoint(svc PlaceService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(PlaceIDRequest)
		page, err := svc.ListEntries(ctx, req.PlaceID, req.Category, req.Limit, req.Cursor)
		if err != nil {
			return nil, err
		}
		return EntriesPageJSON{Entries: toEntriesJSON(page.Entries), NextCursor: page.NextCursor}, nil
	}
}

func MakeInternalNearbyEndpoint(svc PlaceService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(InternalNearbyRequest)
		places, err := svc.ListNearbyInternal(ctx, req.Lat, req.Lng, req.RadiusM, req.Limit)
		if err != nil {
			return nil, err
		}
		return PlacesPageJSON{Places: toPlacesJSON(places, true)}, nil
	}
}

func MakeInternalAttachEndpoint(svc PlaceService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(InternalAttachRequest)
		postID, err := uuid.Parse(strings.TrimSpace(req.PostID))
		if err != nil {
			return nil, apperr.Validation(apperr.Field{Field: "post_id", Code: "INVALID", Message: "post_id must be a UUID"})
		}
		if req.PlaceID != nil && strings.TrimSpace(*req.PlaceID) != "" {
			placeID, err := uuid.Parse(strings.TrimSpace(*req.PlaceID))
			if err != nil {
				return nil, apperr.Validation(apperr.Field{Field: "place_id", Code: "INVALID", Message: "place_id must be a UUID"})
			}
			p, err := svc.AttachExisting(ctx, placeuc.AttachToExistingInput{
				PlaceID: placeID, PostID: postID,
				AttachScore: req.AttachScore, AttachMethod: req.AttachMethod,
				Embedding: req.Embedding,
			})
			if err != nil {
				return nil, err
			}
			return toPlaceJSON(p, true), nil
		}
		p, err := svc.CreateAndAttach(ctx, placeuc.CreatePlaceAttachInput{
			Name: req.Name, Lat: req.Lat, Lng: req.Lng, PostID: postID,
			AttachScore: req.AttachScore, AttachMethod: req.AttachMethod,
			Embedding: req.Embedding,
		})
		if err != nil {
			return nil, err
		}
		return toPlaceJSON(p, true), nil
	}
}

func MakeInternalSummaryEndpoint(svc PlaceService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(InternalSummaryRequest)
		p, err := svc.UpdateSummary(ctx, req.PlaceID, req.Summary, req.Model)
		if err != nil {
			return nil, err
		}
		return toPlaceJSON(p, false), nil
	}
}

func MakeInternalEntriesEndpoint(svc PlaceService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(InternalEntriesRequest)
		items := make([]placeuc.AddEntryItem, 0, len(req.Entries))
		for _, e := range req.Entries {
			item := placeuc.AddEntryItem{Category: e.Category, Body: e.Body, Source: e.Source}
			if e.SourcePostID != nil && strings.TrimSpace(*e.SourcePostID) != "" {
				id, err := uuid.Parse(strings.TrimSpace(*e.SourcePostID))
				if err != nil {
					return nil, apperr.Validation(apperr.Field{
						Field: "source_post_id", Code: "INVALID", Message: "must be a UUID",
					})
				}
				item.SourcePostID = &id
			}
			items = append(items, item)
		}
		entries, err := svc.AddEntries(ctx, placeuc.AddEntriesInput{PlaceID: req.PlaceID, Entries: items})
		if err != nil {
			return nil, err
		}
		return EntriesPageJSON{Entries: toEntriesJSON(entries)}, nil
	}
}

func MakeInternalGetByPostEndpoint(svc PlaceService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		postID := request.(uuid.UUID)
		p, err := svc.GetByPostID(ctx, postID)
		if err != nil {
			return nil, err
		}
		return toPlaceJSON(p, true), nil
	}
}

func toPlaceJSON(p domain.Place, includeEmbedding bool) PlaceJSON {
	out := PlaceJSON{
		ID: p.ID.String(), Name: p.Name, Lat: p.Lat, Lng: p.Lng,
		Summary: p.Summary, SummaryModel: p.SummaryModel, PostCount: p.PostCount,
		CreatedAt: p.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt: p.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
	if p.SummaryUpdatedAt != nil {
		s := p.SummaryUpdatedAt.UTC().Format(time.RFC3339Nano)
		out.SummaryUpdatedAt = &s
	}
	if includeEmbedding && len(p.Embedding) > 0 {
		out.Embedding = p.Embedding
	}
	return out
}

func toPlacesJSON(places []domain.Place, includeEmbedding bool) []PlaceJSON {
	out := make([]PlaceJSON, 0, len(places))
	for _, p := range places {
		out = append(out, toPlaceJSON(p, includeEmbedding))
	}
	return out
}

func toEntriesJSON(entries []domain.Entry) []EntryJSON {
	out := make([]EntryJSON, 0, len(entries))
	for _, e := range entries {
		item := EntryJSON{
			ID: e.ID.String(), PlaceID: e.PlaceID.String(),
			Category: e.Category, Body: e.Body, Source: e.Source,
			CreatedAt: e.CreatedAt.UTC().Format(time.RFC3339Nano),
		}
		if e.SourcePostID != nil {
			s := e.SourcePostID.String()
			item.SourcePostID = &s
		}
		if e.AuthorID != nil {
			s := e.AuthorID.String()
			item.AuthorID = &s
		}
		out = append(out, item)
	}
	return out
}
