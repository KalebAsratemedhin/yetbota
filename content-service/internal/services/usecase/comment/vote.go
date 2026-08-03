package comment

import (
	"context"

	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
	"github.com/beka-birhanu/yetbota/content-service/drivers/utils"
	ctxRP "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	repository "github.com/beka-birhanu/yetbota/content-service/internal/services/repository"
)

func (s *svc) Vote(ctx context.Context, ctxSess *ctxRP.Context, req *VoteRequest) (*VoteResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err := utils.AllowAccess(ctxSess); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	comment, err := s.commentRepo.Read(ctx, req.CommentID)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	exists, err := s.commentVoteRepo.Exists(ctx, ctxSess.UserSession.UserID, req.CommentID)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	var vote *dbmodels.CommentVote
	if exists {
		vote, err = s.commentVoteRepo.Read(ctx, ctxSess.UserSession.UserID, req.CommentID)
		if err != nil {
			ctxSess.SetErrorMessage(err.Error())
			return nil, err
		}
	}

	if exists && vote.VoteType == req.VoteType {
		return &VoteResponse{Upvote: comment.Upvote, Downvote: comment.Downvote}, nil
	}

	var upvoteDelta, downvoteDelta int
	switch req.VoteType {
	case dbmodels.CommentVoteTypeUpvote:
		upvoteDelta = 1
		if exists {
			vote.VoteType = dbmodels.CommentVoteTypeUpvote
			downvoteDelta = -1
		}
	case dbmodels.CommentVoteTypeDownvote:
		downvoteDelta = 1
		if exists {
			vote.VoteType = dbmodels.CommentVoteTypeDownvote
			upvoteDelta = -1
		}
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

	if exists {
		err = s.commentVoteRepo.Update(ctx, tx, vote)
	} else {
		err = s.commentVoteRepo.Add(ctx, tx, &dbmodels.CommentVote{
			UserID:    ctxSess.UserSession.UserID,
			CommentID: req.CommentID,
			VoteType:  req.VoteType,
		})
	}
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	err = s.commentVoteRepo.UpdateCounts(ctx, tx, req.CommentID, upvoteDelta, downvoteDelta, comment.Upvote, comment.Downvote)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err = repository.CommitTx(tx); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	return &VoteResponse{Upvote: comment.Upvote + upvoteDelta, Downvote: comment.Downvote + downvoteDelta}, nil
}
