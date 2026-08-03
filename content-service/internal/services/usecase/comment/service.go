package comment

import (
	"context"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	domainComment "github.com/beka-birhanu/yetbota/content-service/internal/domain/comment"
	domainCommentvote "github.com/beka-birhanu/yetbota/content-service/internal/domain/commentvote"
	ctxRP "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	domainPost "github.com/beka-birhanu/yetbota/content-service/internal/domain/post"
	domainProcessors "github.com/beka-birhanu/yetbota/content-service/internal/domain/processors"
)

type Service interface {
	Add(ctx context.Context, ctxSess *ctxRP.Context, req *AddRequest) (*AddResponse, error)
	Read(ctx context.Context, ctxSess *ctxRP.Context, req *ReadRequest) (*ReadResponse, error)
	List(ctx context.Context, ctxSess *ctxRP.Context, req *ListRequest) (*ListResponse, error)
	Delete(ctx context.Context, ctxSess *ctxRP.Context, req *DeleteRequest) error
	Vote(ctx context.Context, ctxSess *ctxRP.Context, req *VoteRequest) (*VoteResponse, error)
}

type Config struct {
	CommentRepo     domainComment.Repository     `validate:"required"`
	CommentVoteRepo domainCommentvote.Repository `validate:"required"`
	PostRepo        domainPost.Repository        `validate:"required"`
	Executor        domainProcessors.Executor
}

func (c *Config) Validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type svc struct {
	commentRepo     domainComment.Repository
	commentVoteRepo domainCommentvote.Repository
	postRepo        domainPost.Repository
	executor        domainProcessors.Executor
}

func NewService(cfg *Config) (Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &svc{commentRepo: cfg.CommentRepo, commentVoteRepo: cfg.CommentVoteRepo, postRepo: cfg.PostRepo, executor: cfg.Executor}, nil
}
