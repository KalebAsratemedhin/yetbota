package contribution

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aarondl/null/v8"
	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/content-service/drivers/constants"
	"github.com/beka-birhanu/yetbota/content-service/drivers/utils"
	ctxRP "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	domainContribution "github.com/beka-birhanu/yetbota/content-service/internal/domain/contribution"
	domainContributionVote "github.com/beka-birhanu/yetbota/content-service/internal/domain/contributionvote"
	domainActivity "github.com/beka-birhanu/yetbota/content-service/internal/domain/placeactivity"
	"github.com/beka-birhanu/yetbota/content-service/internal/domain/processors"
	domainScreening "github.com/beka-birhanu/yetbota/content-service/internal/domain/screening"
	repository "github.com/beka-birhanu/yetbota/content-service/internal/services/repository"
	"github.com/google/uuid"
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

	place, err := s.placeRepo.Read(ctx, req.PlaceID)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	if place.Status != constants.PlaceStatusVisible && !utils.IsAdmin(ctxSess) {
		err = &toddlerr.Error{
			PublicStatusCode:  status.Forbidden,
			ServiceStatusCode: status.Forbidden,
			PublicMessage:     "place is not available for contributions",
			ServiceMessage:    fmt.Sprintf("place %s status %s is not writable", place.ID, place.Status),
		}
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err := s.screen(ctx, req.Kind, req.Content, ctxSess.UserSession.UserID, req.PlaceID); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	entity := &domainContribution.Contribution{
		ID:               uuid.NewString(),
		PlaceID:          req.PlaceID,
		UserID:           ctxSess.UserSession.UserID,
		Kind:             req.Kind,
		Content:          strings.TrimSpace(req.Content),
		ParentID:         null.String{},
		ModerationStatus: constants.ModerationStatusVisible,
		Origin:           constants.ContributionOriginUser,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := s.contributionRepo.Add(ctx, nil, entity); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	if req.Kind == KindChange || req.Kind == KindImportant {
		if err := s.placeRepo.RecordContributionActivity(ctx, nil, entity.PlaceID, entity.CreatedAt); err != nil {
			fmt.Printf("unable to record place contribution activity: %s\n", err)
		}
	}
	s.recordActivity(ctx, entity, ctxSess.UserSession.UserID)
	if err := s.executor.TriggerContributionIngestWorkflow(ctx, processors.ContributionIngestWorkflowInput{
		ContributionID: entity.ID,
	}); err != nil {
		fmt.Printf("unable to trigger contribution ingest workflow: %s\n", err)
	}
	return &AddResponse{Contribution: entity}, nil
}

func (s *svc) AddAnswer(ctx context.Context, ctxSess *ctxRP.Context, req *AddAnswerRequest) (*AddResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	if err := utils.AllowAccess(ctxSess); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	parent, err := s.contributionRepo.Read(ctx, req.ContributionID)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	if parent.Kind != KindQuestion {
		err = &toddlerr.Error{
			PublicStatusCode:  status.BadRequest,
			ServiceStatusCode: status.BadRequest,
			PublicMessage:     "answers can only be attached to question contributions",
			ServiceMessage:    fmt.Sprintf("parent contribution %s kind is %s", parent.ID, parent.Kind),
		}
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	if err := s.screen(ctx, KindAnswer, req.Content, ctxSess.UserSession.UserID, parent.PlaceID); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	entity := &domainContribution.Contribution{
		ID:               uuid.NewString(),
		PlaceID:          parent.PlaceID,
		UserID:           ctxSess.UserSession.UserID,
		Kind:             KindAnswer,
		Content:          strings.TrimSpace(req.Content),
		ParentID:         null.StringFrom(parent.ID),
		ModerationStatus: constants.ModerationStatusVisible,
		Origin:           constants.ContributionOriginUser,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	if err := s.contributionRepo.Add(ctx, nil, entity); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	s.recordActivity(ctx, entity, ctxSess.UserSession.UserID)
	if err := s.executor.TriggerContributionIngestWorkflow(ctx, processors.ContributionIngestWorkflowInput{
		ContributionID: entity.ID,
	}); err != nil {
		fmt.Printf("unable to trigger contribution ingest workflow: %s\n", err)
	}
	return &AddResponse{Contribution: entity}, nil
}

func (s *svc) Vote(ctx context.Context, ctxSess *ctxRP.Context, req *VoteRequest) (*VoteResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	if err := utils.AllowAccess(ctxSess); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	contribution, err := s.contributionRepo.Read(ctx, req.ContributionID)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	exists, err := s.contributionVoteRepo.Exists(ctx, ctxSess.UserSession.UserID, req.ContributionID)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	var vote *domainContributionVote.Vote
	if exists {
		vote, err = s.contributionVoteRepo.Read(ctx, ctxSess.UserSession.UserID, req.ContributionID)
		if err != nil {
			ctxSess.SetErrorMessage(err.Error())
			return nil, err
		}
	}
	if exists && vote.VoteType == req.VoteType {
		return &VoteResponse{Upvotes: contribution.Upvotes, Downvotes: contribution.Downvotes}, nil
	}

	upDelta, downDelta := 0, 0
	switch req.VoteType {
	case "like":
		upDelta = 1
		if exists {
			vote.VoteType = "like"
			downDelta = -1
		}
	case "dislike":
		downDelta = 1
		if exists {
			vote.VoteType = "dislike"
			upDelta = -1
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
		err = s.contributionVoteRepo.Update(ctx, tx, vote)
	} else {
		err = s.contributionVoteRepo.Add(ctx, tx, &domainContributionVote.Vote{
			ContributionID: req.ContributionID,
			UserID:         ctxSess.UserSession.UserID,
			VoteType:       req.VoteType,
			CreatedAt:      time.Now(),
		})
	}
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	if err = s.contributionRepo.UpdateVoteCounts(ctx, tx, req.ContributionID, upDelta, downDelta, contribution.Upvotes, contribution.Downvotes); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	if err = repository.CommitTx(tx); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	return &VoteResponse{Upvotes: contribution.Upvotes + upDelta, Downvotes: contribution.Downvotes + downDelta}, nil
}

func (s *svc) ListByPlace(ctx context.Context, ctxSess *ctxRP.Context, req *ListRequest) (*ListResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	opts := &domainContribution.ListOptions{
		PlaceID:     req.PlaceID,
		Kind:        req.Kind,
		Page:        req.Page,
		PageSize:    req.PageSize,
		OnlyVisible: visibleOnlyForRole(ctxSess.UserSession.Role),
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
	ids := make([]string, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	votes, err := s.contributionVoteRepo.List(ctx, &domainContributionVote.ListOptions{
		UserID:          ctxSess.UserSession.UserID,
		ContributionIDs: ids,
	})
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	out := make([]*ContributionItem, 0, len(items))
	for _, item := range items {
		out = append(out, &ContributionItem{
			Contribution: item,
			UserVote:     votes[item.ID],
		})
	}
	return &ListResponse{
		Items:    out,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (s *svc) screen(ctx context.Context, kind, content, userID, placeID string) error {
	if s.screening == nil {
		return nil
	}
	if s.screeningHook != nil && !s.screeningHook.ShouldScreen(kind) {
		return nil
	}
	resp, err := s.screening.Screen(ctx, &domainScreening.Request{
		Kind:    kind,
		Text:    strings.TrimSpace(content),
		UserID:  userID,
		PlaceID: placeID,
	})
	if err != nil {
		return err
	}
	if resp == nil || resp.OK {
		return nil
	}
	return &toddlerr.Error{
		PublicStatusCode:  status.BadRequest,
		ServiceStatusCode: status.BadRequest,
		PublicMessage:     resp.Reason,
		ServiceMessage:    fmt.Sprintf("screening rejected contribution: %s", resp.Reason),
	}
}

func (s *svc) recordActivity(ctx context.Context, entity *domainContribution.Contribution, actorID string) {
	if s.activityRepo == nil || entity == nil {
		return
	}
	eventType := domainActivity.EventContributionAdded
	summary := "New contribution on place"
	if entity.Kind == KindQuestion {
		eventType = domainActivity.EventQuestionOpened
		summary = "A question was opened"
	}
	_ = s.activityRepo.Add(ctx, nil, &domainActivity.Event{
		ID:        uuid.NewString(),
		PlaceID:   entity.PlaceID,
		ActorID:   null.StringFrom(actorID),
		EventType: eventType,
		RefID:     null.StringFrom(entity.ID),
		Summary:   summary,
		CreatedAt: entity.CreatedAt,
	})
}
