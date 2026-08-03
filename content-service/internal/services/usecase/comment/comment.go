package comment

import (
	"context"
	"fmt"

	"github.com/aarondl/null/v8"
	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
	"github.com/beka-birhanu/yetbota/content-service/drivers/utils"
	domainComment "github.com/beka-birhanu/yetbota/content-service/internal/domain/comment"
	"github.com/beka-birhanu/yetbota/content-service/internal/domain/commentvote"
	ctxRP "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	domainProcessors "github.com/beka-birhanu/yetbota/content-service/internal/domain/processors"
	repository "github.com/beka-birhanu/yetbota/content-service/internal/services/repository"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

func (s *svc) Add(ctx context.Context, ctxSess *ctxRP.Context, req *AddRequest) (*AddResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err := utils.AllowAccess(ctxSess); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	post, err := s.postRepo.Read(ctx, req.PostID)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	comment := &dbmodels.Comment{
		ID:        uuid.New().String(),
		Content:   req.Comment,
		UserID:    ctxSess.UserSession.UserID,
		PostID:    req.PostID,
		CommentID: null.NewString(req.CommentID, req.CommentID != ""),
		IsAnswer:  req.IsAnswer,
	}

	tx, err := repository.BeginNewTx(ctx)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if err = s.commentRepo.Add(ctx, tx, comment); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err = s.postRepo.UpdateCommentCount(ctx, tx, req.PostID, 1, post.CommentCount); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err = repository.CommitTx(tx); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if comment.IsAnswer && s.executor != nil {
		if tErr := s.executor.TriggerAnswerEmbeddingWorkflow(ctx, domainProcessors.AnswerEmbeddingWorkflowInput{
			CommentID: comment.ID,
		}); tErr != nil {
			fmt.Printf("Unable to trigger answer embedding workflow: %s\n", tErr)
		}
	}

	return &AddResponse{Comment: comment}, nil
}

func (s *svc) Read(ctx context.Context, ctxSess *ctxRP.Context, req *ReadRequest) (*ReadResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err := utils.AllowAccess(ctxSess); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	comment, err := s.commentRepo.Read(ctx, req.ID)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err := utils.EnsureVisible(ctxSess, comment.ModerationStatus, "comment", comment.ID); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	return &ReadResponse{Comment: comment}, nil
}

func (s *svc) List(ctx context.Context, ctxSess *ctxRP.Context, req *ListRequest) (*ListResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err := utils.AllowAccess(ctxSess); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	opts := &domainComment.Options{
		PostID:    req.PostID,
		CommentID: req.CommentID,
		Page:      req.Page,
		PageSize:  req.PageSize,
		OnlyVisible: !utils.IsAdmin(ctxSess),
	}

	errGrp, egCtx := errgroup.WithContext(ctx)
	var comments dbmodels.CommentSlice
	var total int64

	errGrp.Go(func() error {
		var err error
		comments, err = s.commentRepo.List(egCtx, opts)
		return err
	})
	errGrp.Go(func() error {
		var err error
		total, err = s.commentRepo.Count(egCtx, opts)
		return err
	})

	if err := errGrp.Wait(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	commentIDs := make([]string, len(comments))
	for i, comment := range comments {
		commentIDs[i] = comment.ID
	}

	interactions, err := s.commentVoteRepo.List(ctx, &commentvote.ListOptions{
		UserID:     ctxSess.UserSession.UserID,
		CommentIDs: commentIDs,
	})
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	return &ListResponse{
		Comments:     comments,
		Interactions: interactions,
		Total:        total,
		Page:         req.Page,
		PageSize:     req.PageSize,
	}, nil
}

func (s *svc) Delete(ctx context.Context, ctxSess *ctxRP.Context, req *DeleteRequest) error {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return err
	}

	if err := utils.AllowAccess(ctxSess); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return err
	}

	comment, err := s.commentRepo.Read(ctx, req.ID)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return err
	}

	if comment.UserID != ctxSess.UserSession.UserID {
		err = &toddlerr.Error{
			PublicStatusCode:  status.NotFound,
			PublicMessage:     "Either the comment does not exist or you don't have access to delete it.",
			ServiceStatusCode: status.Forbidden,
			ServiceMessage:    fmt.Sprintf("user %s tried to delete a comment under the ownership of %s", ctxSess.UserSession.UserID, comment.UserID),
		}
		ctxSess.SetErrorMessage(err.Error())
		return err
	}

	post, err := s.postRepo.Read(ctx, comment.PostID)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return err
	}

	tx, err := repository.BeginNewTx(ctx)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if err = s.commentRepo.Delete(ctx, tx, req.ID); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return err
	}

	if err = s.postRepo.UpdateCommentCount(ctx, tx, comment.PostID, -1, post.CommentCount); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return err
	}

	return repository.CommitTx(tx)
}
