package placefeed

import (
	"context"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/yetbota/content-service/drivers/utils"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	ctxRP "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	domainActivity "github.com/beka-birhanu/yetbota/content-service/internal/domain/placeactivity"
	domainBookmark "github.com/beka-birhanu/yetbota/content-service/internal/domain/placebookmark"
)

type Service interface {
	List(ctx context.Context, ctxSess *ctxRP.Context, req *ListRequest) (*ListResponse, error)
}

type Config struct {
	ActivityRepo domainActivity.Repository `validate:"required"`
	BookmarkRepo domainBookmark.Repository `validate:"required"`
}

func (c *Config) Validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type svc struct {
	activityRepo domainActivity.Repository
	bookmarkRepo domainBookmark.Repository
}

func NewService(cfg *Config) (Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &svc{
		activityRepo: cfg.ActivityRepo,
		bookmarkRepo: cfg.BookmarkRepo,
	}, nil
}

type ListRequest struct {
	BookmarkedOnly bool
	Page           int
	PageSize       int
}

func (r *ListRequest) Validate() error {
	if r.Page <= 0 {
		r.Page = 1
	}
	if r.PageSize <= 0 {
		r.PageSize = 20
	}
	if r.PageSize > 100 {
		r.PageSize = 100
	}
	return nil
}

type ListResponse struct {
	Events   []*domainActivity.Event
	Total    int64
	Page     int
	PageSize int
}

func (s *svc) List(ctx context.Context, ctxSess *ctxRP.Context, req *ListRequest) (*ListResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	opts := &domainActivity.ListOptions{
		Page:     req.Page,
		PageSize: req.PageSize,
	}
	if req.BookmarkedOnly {
		if err := utils.AllowAccess(ctxSess); err != nil {
			ctxSess.SetErrorMessage(err.Error())
			return nil, err
		}
		ids, err := s.bookmarkRepo.ListPlaceIDs(ctx, ctxSess.UserSession.UserID)
		if err != nil {
			ctxSess.SetErrorMessage(err.Error())
			return nil, err
		}
		if len(ids) == 0 {
			return &ListResponse{Events: []*domainActivity.Event{}, Page: req.Page, PageSize: req.PageSize}, nil
		}
		opts.PlaceIDs = ids
	}
	events, err := s.activityRepo.List(ctx, opts)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	total, err := s.activityRepo.Count(ctx, opts)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	return &ListResponse{
		Events:   events,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}
