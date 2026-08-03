package moderation

import (
	"context"
	"database/sql"
	"time"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	domainComment "github.com/beka-birhanu/yetbota/content-service/internal/domain/comment"
	ctxRP "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	domainModeration "github.com/beka-birhanu/yetbota/content-service/internal/domain/moderation"
	domainNotification "github.com/beka-birhanu/yetbota/content-service/internal/domain/notification"
	domainPost "github.com/beka-birhanu/yetbota/content-service/internal/domain/post"
	domainUser "github.com/beka-birhanu/yetbota/content-service/internal/domain/user"
)

const statusTooManyRequests = status.StatusCode(4290)

type Service interface {
	Report(ctx context.Context, ctxSess *ctxRP.Context, req *ReportRequest) (*ReportResponse, error)
	ListCases(ctx context.Context, ctxSess *ctxRP.Context, req *ListCasesRequest) (*ListCasesResponse, error)
	GetCase(ctx context.Context, ctxSess *ctxRP.Context, req *GetCaseRequest) (*GetCaseResponse, error)
	ActOnCase(ctx context.Context, ctxSess *ctxRP.Context, req *ActOnCaseRequest) (*ActOnCaseResponse, error)
}

type Config struct {
	ModerationRepo     domainModeration.Repository  `validate:"required"`
	RateLimiter        domainModeration.RateLimiter `validate:"required"`
	PostRepo           domainPost.Repository        `validate:"required"`
	CommentRepo        domainComment.Repository     `validate:"required"`
	Banner             domainModeration.UserBanner
	NotificationRepo   domainNotification.Repository
	NotificationSender domainNotification.Sender
	UserRepo           domainUser.Repository
	AutoHideThreshold  int
	RateLimitMax       int
	RateLimitWindow    time.Duration
}

func (c *Config) Validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type svc struct {
	repo               domainModeration.Repository
	rateLimiter        domainModeration.RateLimiter
	postRepo           domainPost.Repository
	commentRepo        domainComment.Repository
	banner             domainModeration.UserBanner
	notificationRepo   domainNotification.Repository
	notificationSender domainNotification.Sender
	userRepo           domainUser.Repository
	autoHideThreshold  int
	rateLimitMax       int
	rateLimitWindow    time.Duration
}

func NewService(cfg *Config) (Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	s := &svc{
		repo:               cfg.ModerationRepo,
		rateLimiter:        cfg.RateLimiter,
		postRepo:           cfg.PostRepo,
		commentRepo:        cfg.CommentRepo,
		banner:             cfg.Banner,
		notificationRepo:   cfg.NotificationRepo,
		notificationSender: cfg.NotificationSender,
		userRepo:           cfg.UserRepo,
		autoHideThreshold:  cfg.AutoHideThreshold,
		rateLimitMax:       cfg.RateLimitMax,
		rateLimitWindow:    cfg.RateLimitWindow,
	}
	if s.autoHideThreshold <= 0 {
		s.autoHideThreshold = 5
	}
	if s.rateLimitMax <= 0 {
		s.rateLimitMax = 10
	}
	if s.rateLimitWindow <= 0 {
		s.rateLimitWindow = time.Hour
	}
	return s, nil
}

func (s *svc) contentMeta(ctx context.Context, contentType, contentID string) (ownerID, modStatus string, err error) {
	switch contentType {
	case domainModeration.ContentTypePost:
		p, e := s.postRepo.Read(ctx, contentID)
		if e != nil {
			return "", "", e
		}
		return p.UserID, p.ModerationStatus, nil
	case domainModeration.ContentTypeComment:
		c, e := s.commentRepo.Read(ctx, contentID)
		if e != nil {
			return "", "", e
		}
		return c.UserID, c.ModerationStatus, nil
	default:
		return "", "", badRequest("invalid content_type")
	}
}

func (s *svc) setContentStatus(ctx context.Context, tx *sql.Tx, contentType, contentID, modStatus string) error {
	switch contentType {
	case domainModeration.ContentTypePost:
		return s.postRepo.UpdateModerationStatus(ctx, tx, contentID, modStatus)
	case domainModeration.ContentTypeComment:
		return s.commentRepo.UpdateModerationStatus(ctx, tx, contentID, modStatus)
	default:
		return badRequest("invalid content_type")
	}
}

func badRequest(msg string) error {
	return &toddlerr.Error{
		PublicStatusCode:  status.BadRequest,
		ServiceStatusCode: status.BadRequest,
		PublicMessage:     msg,
		ServiceMessage:    msg,
	}
}

func conflict(msg string) error {
	return &toddlerr.Error{
		PublicStatusCode:  status.Conflict,
		ServiceStatusCode: status.Conflict,
		PublicMessage:     msg,
		ServiceMessage:    msg,
	}
}

func serverError(msg string) error {
	return &toddlerr.Error{
		PublicStatusCode:  status.ServerError,
		ServiceStatusCode: status.ServerError,
		PublicMessage:     "something went wrong",
		ServiceMessage:    msg,
	}
}

func notFound(msg string) error {
	return &toddlerr.Error{
		PublicStatusCode:  status.NotFound,
		ServiceStatusCode: status.NotFound,
		PublicMessage:     msg,
		ServiceMessage:    msg,
	}
}

func tooManyRequests(msg string) error {
	return &toddlerr.Error{
		PublicStatusCode:  statusTooManyRequests,
		ServiceStatusCode: statusTooManyRequests,
		PublicMessage:     "too many requests, please slow down",
		ServiceMessage:    msg,
	}
}
