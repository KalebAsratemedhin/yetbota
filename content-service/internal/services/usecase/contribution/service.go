package contribution

import (
	"context"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	ctxRP "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	domainContribution "github.com/beka-birhanu/yetbota/content-service/internal/domain/contribution"
	domainContributionVote "github.com/beka-birhanu/yetbota/content-service/internal/domain/contributionvote"
	domainNotification "github.com/beka-birhanu/yetbota/content-service/internal/domain/notification"
	domainPlace "github.com/beka-birhanu/yetbota/content-service/internal/domain/place"
	domainActivity "github.com/beka-birhanu/yetbota/content-service/internal/domain/placeactivity"
	"github.com/beka-birhanu/yetbota/content-service/internal/domain/processors"
	domainScreening "github.com/beka-birhanu/yetbota/content-service/internal/domain/screening"
	domainUser "github.com/beka-birhanu/yetbota/content-service/internal/domain/user"
)

type Service interface {
	Add(ctx context.Context, ctxSess *ctxRP.Context, req *AddRequest) (*AddResponse, error)
	AddAnswer(ctx context.Context, ctxSess *ctxRP.Context, req *AddAnswerRequest) (*AddResponse, error)
	Vote(ctx context.Context, ctxSess *ctxRP.Context, req *VoteRequest) (*VoteResponse, error)
	ListByPlace(ctx context.Context, ctxSess *ctxRP.Context, req *ListRequest) (*ListResponse, error)
	AcceptAnswer(ctx context.Context, ctxSess *ctxRP.Context, req *AcceptAnswerRequest) (*AcceptAnswerResponse, error)
	ListOpenQuestions(ctx context.Context, ctxSess *ctxRP.Context, req *ListOpenQuestionsRequest) (*ListOpenQuestionsResponse, error)
}

type Config struct {
	ContributionRepo     domainContribution.Repository     `validate:"required"`
	ContributionVoteRepo domainContributionVote.Repository `validate:"required"`
	PlaceRepo            domainPlace.Repository            `validate:"required"`
	Executor             processors.Executor               `validate:"required"`
	UserRepo             domainUser.Repository
	NotificationRepo     domainNotification.Repository
	ActivityRepo         domainActivity.Repository
	Screening            domainScreening.Client
	ScreeningHook        ScreeningHook
}

func (c *Config) Validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type svc struct {
	contributionRepo     domainContribution.Repository
	contributionVoteRepo domainContributionVote.Repository
	placeRepo            domainPlace.Repository
	executor             processors.Executor
	userRepo             domainUser.Repository
	notificationRepo     domainNotification.Repository
	activityRepo         domainActivity.Repository
	screening            domainScreening.Client
	screeningHook        ScreeningHook
}

func NewService(cfg *Config) (Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &svc{
		contributionRepo:     cfg.ContributionRepo,
		contributionVoteRepo: cfg.ContributionVoteRepo,
		placeRepo:            cfg.PlaceRepo,
		executor:             cfg.Executor,
		userRepo:             cfg.UserRepo,
		notificationRepo:     cfg.NotificationRepo,
		activityRepo:         cfg.ActivityRepo,
		screening:            cfg.Screening,
		screeningHook:        cfg.ScreeningHook,
	}, nil
}
