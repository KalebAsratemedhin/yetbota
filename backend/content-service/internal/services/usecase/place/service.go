package place

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/yetbota/content-service/internal/domain/apperr"
	domain "github.com/yetbota/content-service/internal/domain/place"
	domainpost "github.com/yetbota/content-service/internal/domain/post"
)

type Service struct {
	repo            domain.Repository
	posts           domainpost.Repository
	defaultPageSize int
	maxPageSize     int
}

func NewService(repo domain.Repository, posts domainpost.Repository, defaultPageSize, maxPageSize int) *Service {
	if defaultPageSize <= 0 {
		defaultPageSize = 20
	}
	if maxPageSize <= 0 {
		maxPageSize = 50
	}
	return &Service{repo: repo, posts: posts, defaultPageSize: defaultPageSize, maxPageSize: maxPageSize}
}

type ListInput struct {
	Query  string
	Limit  int
	Cursor string
	Lat    *float64
	Lng    *float64
	Radius float64 // meters; used when lat/lng set
}

type PlacePageResult struct {
	Places     []domain.Place
	NextCursor string
}

type PostPageResult struct {
	Posts      []domainpost.Post
	NextCursor string
}

type EntryPageResult struct {
	Entries    []domain.Entry
	NextCursor string
}

type AttachToExistingInput struct {
	PlaceID      uuid.UUID
	PostID       uuid.UUID
	AttachScore  float64
	AttachMethod string
	Embedding    []float64
}

type CreatePlaceAttachInput struct {
	Name         string
	Lat          float64
	Lng          float64
	PostID       uuid.UUID
	AttachScore  float64
	AttachMethod string
	Embedding    []float64
}

type AddEntriesInput struct {
	PlaceID uuid.UUID
	Entries []AddEntryItem
}

type AddEntryItem struct {
	Category     string
	Body         string
	Source       string
	SourcePostID *uuid.UUID
	AuthorID     *uuid.UUID
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (domain.Place, error) {
	return s.repo.Get(ctx, id)
}

func (s *Service) GetByPostID(ctx context.Context, postID uuid.UUID) (domain.Place, error) {
	return s.repo.GetByPostID(ctx, postID)
}

func (s *Service) List(ctx context.Context, in ListInput) (PlacePageResult, error) {
	limit := s.clampLimit(in.Limit)
	if in.Lat != nil || in.Lng != nil {
		if in.Lat == nil || in.Lng == nil {
			return PlacePageResult{}, apperr.Validation(apperr.Field{
				Field: "lat/lng", Code: "INVALID", Message: "lat and lng must both be provided or both omitted",
			})
		}
		if *in.Lat < -90 || *in.Lat > 90 || *in.Lng < -180 || *in.Lng > 180 {
			return PlacePageResult{}, apperr.Validation(apperr.Field{
				Field: "lat/lng", Code: "INVALID", Message: "lat/lng out of range",
			})
		}
		radius := in.Radius
		if radius <= 0 {
			radius = 5000 // 5km default for public nearby browse
		}
		places, err := s.repo.ListNearby(ctx, domain.NearbyParams{
			Lat: *in.Lat, Lng: *in.Lng, RadiusM: radius, Limit: limit,
		})
		if err != nil {
			return PlacePageResult{}, err
		}
		if q := strings.TrimSpace(in.Query); q != "" {
			filtered := places[:0]
			lq := strings.ToLower(q)
			for _, p := range places {
				if strings.Contains(strings.ToLower(p.Name), lq) {
					filtered = append(filtered, p)
				}
			}
			places = filtered
		}
		return PlacePageResult{Places: places}, nil
	}

	cur, err := decodeCursor(in.Cursor)
	if err != nil {
		return PlacePageResult{}, err
	}
	page, err := s.repo.List(ctx, domain.ListParams{
		Limit: limit, Cursor: cur, Query: strings.TrimSpace(in.Query),
	})
	if err != nil {
		return PlacePageResult{}, err
	}
	return PlacePageResult{Places: page.Places, NextCursor: encodeCursor(page.NextCursor)}, nil
}

func (s *Service) ListNearbyInternal(ctx context.Context, lat, lng, radiusM float64, limit int) ([]domain.Place, error) {
	if lat < -90 || lat > 90 || lng < -180 || lng > 180 {
		return nil, apperr.Validation(apperr.Field{
			Field: "lat/lng", Code: "INVALID", Message: "lat/lng out of range",
		})
	}
	if radiusM <= 0 {
		radiusM = 150
	}
	if limit <= 0 {
		limit = 20
	}
	return s.repo.ListNearby(ctx, domain.NearbyParams{
		Lat: lat, Lng: lng, RadiusM: radiusM, Limit: limit,
	})
}

func (s *Service) ListPosts(ctx context.Context, placeID uuid.UUID, limit int, cursor string) (PostPageResult, error) {
	if _, err := s.repo.Get(ctx, placeID); err != nil {
		return PostPageResult{}, err
	}
	limit = s.clampLimit(limit)
	cur, err := decodeAttachCursor(cursor)
	if err != nil {
		return PostPageResult{}, err
	}
	page, err := s.repo.ListPostIDs(ctx, placeID, limit, cur)
	if err != nil {
		return PostPageResult{}, err
	}
	posts, err := s.posts.ByIDs(ctx, page.PostIDs)
	if err != nil {
		return PostPageResult{}, err
	}
	return PostPageResult{Posts: posts, NextCursor: encodeAttachCursor(page.NextCursor)}, nil
}

func (s *Service) ListEntries(ctx context.Context, placeID uuid.UUID, category string, limit int, cursor string) (EntryPageResult, error) {
	if _, err := s.repo.Get(ctx, placeID); err != nil {
		return EntryPageResult{}, err
	}
	category = strings.ToLower(strings.TrimSpace(category))
	if category != "" && category != domain.CategoryTip && category != domain.CategoryChange && category != domain.CategoryImportant {
		return EntryPageResult{}, apperr.Validation(apperr.Field{
			Field: "category", Code: "INVALID", Message: "category must be tip, change, or important",
		})
	}
	limit = s.clampLimit(limit)
	cur, err := decodeCursor(cursor)
	if err != nil {
		return EntryPageResult{}, err
	}
	page, err := s.repo.ListEntries(ctx, placeID, domain.EntryListParams{
		Limit: limit, Cursor: cur, Category: category,
	})
	if err != nil {
		return EntryPageResult{}, err
	}
	return EntryPageResult{Entries: page.Entries, NextCursor: encodeCursor(page.NextCursor)}, nil
}

func (s *Service) AttachExisting(ctx context.Context, in AttachToExistingInput) (domain.Place, error) {
	if err := s.validatePost(ctx, in.PostID); err != nil {
		return domain.Place{}, err
	}
	if _, err := s.repo.Get(ctx, in.PlaceID); err != nil {
		return domain.Place{}, err
	}
	method := strings.TrimSpace(in.AttachMethod)
	if method == "" {
		method = "geo_name_embed"
	}
	return s.repo.AttachPost(ctx, domain.AttachInput{
		PlaceID: in.PlaceID, PostID: in.PostID,
		AttachScore: in.AttachScore, AttachMethod: method,
		Embedding: in.Embedding,
	})
}

func (s *Service) CreateAndAttach(ctx context.Context, in CreatePlaceAttachInput) (domain.Place, error) {
	name := strings.TrimSpace(in.Name)
	var details []apperr.Field
	if name == "" {
		details = append(details, apperr.Required("name"))
	} else if utf8.RuneCountInString(name) > 200 {
		details = append(details, apperr.Field{Field: "name", Code: "TOO_LONG", Message: "name must be at most 200 characters"})
	}
	if in.Lat < -90 || in.Lat > 90 || in.Lng < -180 || in.Lng > 180 {
		details = append(details, apperr.Field{Field: "lat/lng", Code: "INVALID", Message: "lat/lng out of range"})
	}
	if len(details) > 0 {
		return domain.Place{}, apperr.Validation(details...)
	}
	if err := s.validatePost(ctx, in.PostID); err != nil {
		return domain.Place{}, err
	}
	method := strings.TrimSpace(in.AttachMethod)
	if method == "" {
		method = "created_seed"
	}
	return s.repo.CreateAndAttach(ctx, domain.CreateAndAttachInput{
		Name: name, Lat: in.Lat, Lng: in.Lng, PostID: in.PostID,
		AttachScore: in.AttachScore, AttachMethod: method, Embedding: in.Embedding,
	})
}

func (s *Service) UpdateSummary(ctx context.Context, placeID uuid.UUID, summary, model string) (domain.Place, error) {
	summary = strings.TrimSpace(summary)
	if summary == "" {
		return domain.Place{}, apperr.Validation(apperr.Required("summary"))
	}
	if utf8.RuneCountInString(summary) > 8000 {
		return domain.Place{}, apperr.Validation(apperr.Field{
			Field: "summary", Code: "TOO_LONG", Message: "summary must be at most 8000 characters",
		})
	}
	return s.repo.UpdateSummary(ctx, domain.UpdateSummaryInput{
		PlaceID: placeID, Summary: summary, SummaryModel: strings.TrimSpace(model),
	})
}

func (s *Service) AddEntries(ctx context.Context, in AddEntriesInput) ([]domain.Entry, error) {
	if _, err := s.repo.Get(ctx, in.PlaceID); err != nil {
		return nil, err
	}
	out := make([]domain.Entry, 0, len(in.Entries))
	for i, item := range in.Entries {
		cat := strings.ToLower(strings.TrimSpace(item.Category))
		body := strings.TrimSpace(item.Body)
		src := strings.ToLower(strings.TrimSpace(item.Source))
		if src == "" {
			src = domain.SourceAI
		}
		var details []apperr.Field
		if cat != domain.CategoryTip && cat != domain.CategoryChange && cat != domain.CategoryImportant {
			details = append(details, apperr.Field{
				Field: fmt.Sprintf("entries[%d].category", i), Code: "INVALID",
				Message: "category must be tip, change, or important",
			})
		}
		if body == "" {
			details = append(details, apperr.Field{
				Field: fmt.Sprintf("entries[%d].body", i), Code: "REQUIRED", Message: "required",
			})
		} else if utf8.RuneCountInString(body) > 2000 {
			details = append(details, apperr.Field{
				Field: fmt.Sprintf("entries[%d].body", i), Code: "TOO_LONG", Message: "at most 2000 characters",
			})
		}
		if src != domain.SourceAI && src != domain.SourceUser {
			details = append(details, apperr.Field{
				Field: fmt.Sprintf("entries[%d].source", i), Code: "INVALID", Message: "source must be ai or user",
			})
		}
		if len(details) > 0 {
			return nil, apperr.Validation(details...)
		}
		e, err := s.repo.AddEntry(ctx, domain.AddEntryInput{
			PlaceID: in.PlaceID, Category: cat, Body: body, Source: src,
			SourcePostID: item.SourcePostID, AuthorID: item.AuthorID,
		})
		if err != nil {
			if ae, ok := apperr.As(err); ok && len(ae.Details) == 1 && ae.Details[0].Code == "DUPLICATE" {
				continue
			}
			return nil, err
		}
		out = append(out, e)
	}
	return out, nil
}

func (s *Service) validatePost(ctx context.Context, postID uuid.UUID) error {
	if _, err := s.posts.ByID(ctx, postID); err != nil {
		return err
	}
	return nil
}

func (s *Service) clampLimit(n int) int {
	if n <= 0 {
		return s.defaultPageSize
	}
	if n > s.maxPageSize {
		return s.maxPageSize
	}
	return n
}

func encodeCursor(c *domain.PageCursor) string {
	if c == nil {
		return ""
	}
	raw := fmt.Sprintf("%d|%s", c.CreatedAt.UTC().UnixNano(), c.ID.String())
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func decodeCursor(raw string) (*domain.PageCursor, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	b, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return nil, apperr.Validation(apperr.Field{Field: "cursor", Code: "INVALID", Message: "invalid cursor"})
	}
	parts := strings.SplitN(string(b), "|", 2)
	if len(parts) != 2 {
		return nil, apperr.Validation(apperr.Field{Field: "cursor", Code: "INVALID", Message: "invalid cursor"})
	}
	ns, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return nil, apperr.Validation(apperr.Field{Field: "cursor", Code: "INVALID", Message: "invalid cursor"})
	}
	id, err := uuid.Parse(parts[1])
	if err != nil {
		return nil, apperr.Validation(apperr.Field{Field: "cursor", Code: "INVALID", Message: "invalid cursor"})
	}
	return &domain.PageCursor{CreatedAt: time.Unix(0, ns).UTC(), ID: id}, nil
}

func encodeAttachCursor(c *domain.AttachCursor) string {
	if c == nil {
		return ""
	}
	raw := fmt.Sprintf("%d|%s", c.AttachedAt.UTC().UnixNano(), c.PostID.String())
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func decodeAttachCursor(raw string) (*domain.AttachCursor, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	b, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return nil, apperr.Validation(apperr.Field{Field: "cursor", Code: "INVALID", Message: "invalid cursor"})
	}
	parts := strings.SplitN(string(b), "|", 2)
	if len(parts) != 2 {
		return nil, apperr.Validation(apperr.Field{Field: "cursor", Code: "INVALID", Message: "invalid cursor"})
	}
	ns, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return nil, apperr.Validation(apperr.Field{Field: "cursor", Code: "INVALID", Message: "invalid cursor"})
	}
	id, err := uuid.Parse(parts[1])
	if err != nil {
		return nil, apperr.Validation(apperr.Field{Field: "cursor", Code: "INVALID", Message: "invalid cursor"})
	}
	return &domain.AttachCursor{AttachedAt: time.Unix(0, ns).UTC(), PostID: id}, nil
}
