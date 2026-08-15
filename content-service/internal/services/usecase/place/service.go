package place

import (
	"context"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	ctxRP "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	domainContribution "github.com/beka-birhanu/yetbota/content-service/internal/domain/contribution"
	domainContributionVote "github.com/beka-birhanu/yetbota/content-service/internal/domain/contributionvote"
	domainPhoto "github.com/beka-birhanu/yetbota/content-service/internal/domain/photo"
	domainPlace "github.com/beka-birhanu/yetbota/content-service/internal/domain/place"
	domainActivity "github.com/beka-birhanu/yetbota/content-service/internal/domain/placeactivity"
	domainBookmark "github.com/beka-birhanu/yetbota/content-service/internal/domain/placebookmark"
	domainConfirmation "github.com/beka-birhanu/yetbota/content-service/internal/domain/placeconfirmation"
	domainPlacePhoto "github.com/beka-birhanu/yetbota/content-service/internal/domain/placephoto"
	domainProblem "github.com/beka-birhanu/yetbota/content-service/internal/domain/placeproblem"
	"github.com/beka-birhanu/yetbota/content-service/internal/domain/processors"
	domainScreening "github.com/beka-birhanu/yetbota/content-service/internal/domain/screening"
	domainStorage "github.com/beka-birhanu/yetbota/content-service/internal/domain/storage"
)

type Service interface {
	Read(ctx context.Context, ctxSess *ctxRP.Context, req *ReadRequest) (*ReadResponse, error)
	List(ctx context.Context, ctxSess *ctxRP.Context, req *ListRequest) (*ListResponse, error)
	Bookmark(ctx context.Context, ctxSess *ctxRP.Context, req *BookmarkRequest) (*BookmarkResponse, error)
	Unbookmark(ctx context.Context, ctxSess *ctxRP.Context, req *UnbookmarkRequest) (*UnbookmarkResponse, error)
	Confirm(ctx context.Context, ctxSess *ctxRP.Context, req *ConfirmRequest) (*ConfirmResponse, error)
	ReportProblem(ctx context.Context, ctxSess *ctxRP.Context, req *ReportProblemRequest) (*ReportProblemResponse, error)
	UpdateSummary(ctx context.Context, ctxSess *ctxRP.Context, req *UpdateSummaryRequest) (*UpdateSummaryResponse, error)
}

type Config struct {
	PlaceRepo            domainPlace.Repository            `validate:"required"`
	PlacePhotoRepo       domainPlacePhoto.Repository       `validate:"required"`
	ContributionRepo     domainContribution.Repository     `validate:"required"`
	ContributionVoteRepo domainContributionVote.Repository `validate:"required"`
	BookmarkRepo         domainBookmark.Repository         `validate:"required"`
	ConfirmationRepo     domainConfirmation.Repository     `validate:"required"`
	ProblemRepo          domainProblem.Repository          `validate:"required"`
	PhotoRepo            domainPhoto.Repository            `validate:"required"`
	Bucket               domainStorage.Bucket              `validate:"required"`
	Executor             processors.Executor               `validate:"required"`
	ActivityRepo         domainActivity.Repository
	Screening            domainScreening.Client
	DecayDays            int
	VerificationBoost    float64
}

func (c *Config) Validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type svc struct {
	placeRepo            domainPlace.Repository
	placePhotoRepo       domainPlacePhoto.Repository
	contributionRepo     domainContribution.Repository
	contributionVoteRepo domainContributionVote.Repository
	bookmarkRepo         domainBookmark.Repository
	confirmationRepo     domainConfirmation.Repository
	problemRepo          domainProblem.Repository
	photoRepo            domainPhoto.Repository
	bucket               domainStorage.Bucket
	executor             processors.Executor
	activityRepo         domainActivity.Repository
	screening            domainScreening.Client
	decayDays            int
	verificationBoost    float64
}

func NewService(cfg *Config) (Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	decayDays := cfg.DecayDays
	if decayDays <= 0 {
		decayDays = 90
	}
	boost := cfg.VerificationBoost
	if boost <= 0 {
		boost = 20
	}
	return &svc{
		placeRepo:            cfg.PlaceRepo,
		placePhotoRepo:       cfg.PlacePhotoRepo,
		contributionRepo:     cfg.ContributionRepo,
		contributionVoteRepo: cfg.ContributionVoteRepo,
		bookmarkRepo:         cfg.BookmarkRepo,
		confirmationRepo:     cfg.ConfirmationRepo,
		problemRepo:          cfg.ProblemRepo,
		photoRepo:            cfg.PhotoRepo,
		bucket:               cfg.Bucket,
		executor:             cfg.Executor,
		activityRepo:         cfg.ActivityRepo,
		screening:            cfg.Screening,
		decayDays:            decayDays,
		verificationBoost:    boost,
	}, nil
}
