package user

import (
	"context"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/validator"
	domainAuth "github.com/beka-birhanu/yetbota/identity-service/internal/domain/auth"
	ctxRP "github.com/beka-birhanu/yetbota/identity-service/internal/domain/context"
	domainFollow "github.com/beka-birhanu/yetbota/identity-service/internal/domain/follow"
	domainPhoto "github.com/beka-birhanu/yetbota/identity-service/internal/domain/photo"
	domainStorage "github.com/beka-birhanu/yetbota/identity-service/internal/domain/storage"
	domainUser "github.com/beka-birhanu/yetbota/identity-service/internal/domain/user"
)

type Service interface {
	List(ctx context.Context, ctxSess *ctxRP.Context, req *ListRequest) (*ListResponse, error)
	Read(ctx context.Context, ctxSess *ctxRP.Context, req *ReadRequest) (*ReadResponse, error)
	ReadPublic(ctx context.Context, ctxSess *ctxRP.Context, req *ReadPublicRequest) (*ReadPublicResponse, error)
	Update(ctx context.Context, ctxSess *ctxRP.Context, req *UpdateRequest) (*UpdateResponse, error)
	UpdateSelf(ctx context.Context, ctxSess *ctxRP.Context, req *UpdateSelfRequest) (*UpdateSelfResponse, error)
	Register(ctx context.Context, ctxSess *ctxRP.Context, req *RegisterRequest) (*RegisterResponse, error)
	Delete(ctx context.Context, ctxSess *ctxRP.Context, req *DeleteRequest) (*DeleteResponse, error)
	DeleteSelf(ctx context.Context, ctxSess *ctxRP.Context, req *DeleteSelfRequest) (*DeleteSelfResponse, error)
	UploadProfile(ctx context.Context, ctxSess *ctxRP.Context, req *UploadProfileRequest) (*UploadProfileResponse, error)
	Follow(ctx context.Context, ctxSess *ctxRP.Context, req *FollowRequest) (*FollowResponse, error)
	Unfollow(ctx context.Context, ctxSess *ctxRP.Context, req *UnfollowRequest) (*UnfollowResponse, error)
}

type Config struct {
	UserRepo   domainUser.Repository   `validate:"required"`
	PhotoRepo  domainPhoto.Repository  `validate:"required"`
	FollowRepo domainFollow.Repository `validate:"required"`
	Hasher     domainAuth.Hasher       `validate:"required"`
	Bucket     domainStorage.Bucket    `validate:"required"`
}

func (c *Config) Validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type svc struct {
	userRepo   domainUser.Repository
	photoRepo  domainPhoto.Repository
	followRepo domainFollow.Repository
	hasher     domainAuth.Hasher
	bucket     domainStorage.Bucket
}

func NewService(cfg *Config) (Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &svc{
		userRepo:   cfg.UserRepo,
		photoRepo:  cfg.PhotoRepo,
		followRepo: cfg.FollowRepo,
		hasher:     cfg.Hasher,
		bucket:     cfg.Bucket,
	}, nil
}
