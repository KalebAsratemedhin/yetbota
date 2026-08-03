package contribution

import (
	"context"
	"encoding/json"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/types"
	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
	"github.com/beka-birhanu/yetbota/content-service/drivers/utils"
	ctxRP "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	domainContribution "github.com/beka-birhanu/yetbota/content-service/internal/domain/contribution"
	domainActivity "github.com/beka-birhanu/yetbota/content-service/internal/domain/placeactivity"
	repository "github.com/beka-birhanu/yetbota/content-service/internal/services/repository"
	"github.com/google/uuid"
)

const acceptedAnswerScoreBoost int64 = 50

func (s *svc) AcceptAnswer(ctx context.Context, ctxSess *ctxRP.Context, req *AcceptAnswerRequest) (*AcceptAnswerResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	if err := utils.AllowAccess(ctxSess); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	question, err := s.contributionRepo.Read(ctx, req.QuestionID)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	if question.Kind != KindQuestion {
		err = &toddlerr.Error{
			PublicStatusCode:  status.BadRequest,
			ServiceStatusCode: status.BadRequest,
			PublicMessage:     "only questions can accept answers",
			ServiceMessage:    "accept called on non-question",
		}
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	if question.UserID != ctxSess.UserSession.UserID && !utils.IsAdmin(ctxSess) {
		err = &toddlerr.Error{
			PublicStatusCode:  status.Forbidden,
			ServiceStatusCode: status.Forbidden,
			PublicMessage:     "only the question author can accept an answer",
			ServiceMessage:    "accept forbidden",
		}
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	answer, err := s.contributionRepo.Read(ctx, req.AnswerID)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	if answer.Kind != KindAnswer || !answer.ParentID.Valid || answer.ParentID.String != question.ID {
		err = &toddlerr.Error{
			PublicStatusCode:  status.BadRequest,
			ServiceStatusCode: status.BadRequest,
			PublicMessage:     "answer does not belong to this question",
			ServiceMessage:    "answer/question mismatch",
		}
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	now := time.Now().UTC()
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

	if err = s.contributionRepo.ClearAcceptedAnswer(ctx, tx, question.ID); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	if err = s.contributionRepo.AcceptAnswer(ctx, tx, answer.ID, ctxSess.UserSession.UserID, now); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	if err = repository.CommitTx(tx); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	answer.AcceptedAt = null.TimeFrom(now)
	answer.AcceptedBy = null.StringFrom(ctxSess.UserSession.UserID)

	if s.userRepo != nil {
		if u, uErr := s.userRepo.Read(ctx, answer.UserID); uErr == nil && u != nil {
			u.Score += acceptedAnswerScoreBoost
			_ = s.userRepo.Update(ctx, u)
		}
	}

	if s.notificationRepo != nil && answer.UserID != ctxSess.UserSession.UserID {
		data, _ := json.Marshal(map[string]string{
			"type":        "answer_accepted",
			"question_id": question.ID,
			"answer_id":   answer.ID,
			"place_id":    question.PlaceID,
		})
		n := &dbmodels.Notification{
			ID:     uuid.NewString(),
			UserID: answer.UserID,
			Title:  "Your answer was accepted",
			Body:   "Someone accepted your answer on a place question.",
			Data:   types.JSON(data),
			SentAt: now,
		}
		_ = s.notificationRepo.Add(ctx, nil, n)
	}

	if s.activityRepo != nil {
		_ = s.activityRepo.Add(ctx, nil, &domainActivity.Event{
			ID:        uuid.NewString(),
			PlaceID:   question.PlaceID,
			ActorID:   null.StringFrom(ctxSess.UserSession.UserID),
			EventType: domainActivity.EventAnswerAccepted,
			RefID:     null.StringFrom(answer.ID),
			Summary:   "An answer was accepted",
			CreatedAt: now,
		})
	}

	return &AcceptAnswerResponse{Answer: answer}, nil
}

func (s *svc) ListOpenQuestions(ctx context.Context, ctxSess *ctxRP.Context, req *ListOpenQuestionsRequest) (*ListOpenQuestionsResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	opts := &domainContribution.ListOptions{
		Kind:         KindQuestion,
		OpenOnly:     true,
		Page:         req.Page,
		PageSize:     req.PageSize,
		OnlyVisible:  true,
		NonAnswerTop: true,
		NearLat:      req.NearLat,
		NearLon:      req.NearLon,
		RadiusKm:     req.RadiusKm,
	}
	items, err := s.contributionRepo.List(ctx, opts)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	total, err := s.contributionRepo.Count(ctx, opts)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	return &ListOpenQuestionsResponse{
		Questions: items,
		Total:     total,
		Page:      req.Page,
		PageSize:  req.PageSize,
	}, nil
}
