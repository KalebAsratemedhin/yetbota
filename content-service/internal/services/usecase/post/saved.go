package post

import (
	"context"

	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
	"github.com/beka-birhanu/yetbota/content-service/drivers/utils"
	ctxRP "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
)

func (s *svc) Save(ctx context.Context, ctxSess *ctxRP.Context, req *SaveRequest) (*SaveResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err := utils.AllowAccess(ctxSess); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if _, err := s.postRepo.Read(ctx, req.PostID); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err := s.savedPostRepo.Add(ctx, &dbmodels.SavedPost{
		UserID: ctxSess.UserSession.UserID,
		PostID: req.PostID,
	}); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	return &SaveResponse{}, nil
}

func (s *svc) Unsave(ctx context.Context, ctxSess *ctxRP.Context, req *UnsaveRequest) (*UnsaveResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err := utils.AllowAccess(ctxSess); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	savedPost, err := s.savedPostRepo.Read(ctx, ctxSess.UserSession.UserID, req.PostID)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err := s.savedPostRepo.Delete(ctx, savedPost); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	return &UnsaveResponse{}, nil
}
