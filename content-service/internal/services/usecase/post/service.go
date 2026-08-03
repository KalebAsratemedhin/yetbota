package post

import (
	"context"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	domainCommentvote "github.com/beka-birhanu/yetbota/content-service/internal/domain/commentvote"
	ctxRP "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	domainFollower "github.com/beka-birhanu/yetbota/content-service/internal/domain/follower"
	domainPhoto "github.com/beka-birhanu/yetbota/content-service/internal/domain/photo"
	domainPost "github.com/beka-birhanu/yetbota/content-service/internal/domain/post"
	domainPostphoto "github.com/beka-birhanu/yetbota/content-service/internal/domain/postphoto"
	domainPostvote "github.com/beka-birhanu/yetbota/content-service/internal/domain/postvote"
	"github.com/beka-birhanu/yetbota/content-service/internal/domain/processors"
	domainSavedpost "github.com/beka-birhanu/yetbota/content-service/internal/domain/savedpost"
	domainStorage "github.com/beka-birhanu/yetbota/content-service/internal/domain/storage"
)

type Service interface {
	Add(ctx context.Context, ctxSess *ctxRP.Context, req *AddRequest) (*AddResponse, error)
	Read(ctx context.Context, ctxSess *ctxRP.Context, req *ReadRequest) (*ReadResponse, error)
	Update(ctx context.Context, ctxSess *ctxRP.Context, req *UpdateRequest) (*UpdateResponse, error)
	Vote(ctx context.Context, ctxSess *ctxRP.Context, req *PostVoteRequest) (*PostVoteResponse, error)
	List(ctx context.Context, ctxSess *ctxRP.Context, req *ListRequest) (*ListResponse, error)
	Save(ctx context.Context, ctxSess *ctxRP.Context, req *SaveRequest) (*SaveResponse, error)
	Unsave(ctx context.Context, ctxSess *ctxRP.Context, req *UnsaveRequest) (*UnsaveResponse, error)
}

type Config struct {
	PostRepo        domainPost.Repository        `validate:"required"`
	PostVoteRepo    domainPostvote.Repository    `validate:"required"`
	CommentVoteRepo domainCommentvote.Repository `validate:"required"`
	FollowerRepo    domainFollower.Repository    `validate:"required"`
	SavedPostRepo   domainSavedpost.Repository   `validate:"required"`
	PostPhotoRepo   domainPostphoto.Repository   `validate:"required"`
	PhotoRepo       domainPhoto.Repository       `validate:"required"`
	Bucket          domainStorage.Bucket         `validate:"required"`
	Executor        processors.Executor          `validate:"required"`
	ScoringStream   domainStorage.Stream         `validate:"required"`
}

func (c *Config) Validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type svc struct {
	postRepo        domainPost.Repository
	postVoteRepo    domainPostvote.Repository
	commentVoteRepo domainCommentvote.Repository
	followerRepo    domainFollower.Repository
	savedPostRepo   domainSavedpost.Repository
	postPhotoRepo   domainPostphoto.Repository
	photoRepo       domainPhoto.Repository
	bucket          domainStorage.Bucket
	executor        processors.Executor
	scoringStream   domainStorage.Stream
}

func NewService(cfg *Config) (Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &svc{
		postRepo:        cfg.PostRepo,
		postVoteRepo:    cfg.PostVoteRepo,
		commentVoteRepo: cfg.CommentVoteRepo,
		followerRepo:    cfg.FollowerRepo,
		savedPostRepo:   cfg.SavedPostRepo,
		postPhotoRepo:   cfg.PostPhotoRepo,
		photoRepo:       cfg.PhotoRepo,
		bucket:          cfg.Bucket,
		executor:        cfg.Executor,
		scoringStream:   cfg.ScoringStream,
	}, nil
}
