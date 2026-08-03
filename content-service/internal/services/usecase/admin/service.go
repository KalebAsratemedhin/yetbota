package admin

import (
	"context"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	domainAdmin "github.com/beka-birhanu/yetbota/content-service/internal/domain/admin"
	ctxRP "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
)

const defaultReputationThreshold = 2000

type Service interface {
	OverviewStats(ctx context.Context, ctxSess *ctxRP.Context) (*OverviewStatsResponse, error)
	OverviewGrowth(ctx context.Context, ctxSess *ctxRP.Context, req *OverviewGrowthRequest) (*OverviewGrowthResponse, error)
	UserStats(ctx context.Context, ctxSess *ctxRP.Context) (*UserStatsResponse, error)
	SystemAudit(ctx context.Context, ctxSess *ctxRP.Context, req *SystemAuditRequest) (*SystemAuditResponse, error)
}

type Config struct {
	Repo                domainAdmin.Repository `validate:"required"`
	ReputationThreshold int64
}

func (c *Config) Validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type svc struct {
	repo                domainAdmin.Repository
	reputationThreshold int64
}

func NewService(cfg *Config) (Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	s := &svc{
		repo:                cfg.Repo,
		reputationThreshold: cfg.ReputationThreshold,
	}
	if s.reputationThreshold <= 0 {
		s.reputationThreshold = defaultReputationThreshold
	}
	return s, nil
}

func badRequest(msg string) error {
	return &toddlerr.Error{
		PublicStatusCode:  status.BadRequest,
		ServiceStatusCode: status.BadRequest,
		PublicMessage:     msg,
		ServiceMessage:    msg,
	}
}
