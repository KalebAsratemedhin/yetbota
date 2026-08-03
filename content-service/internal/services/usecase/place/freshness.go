package place

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
	domainPlace "github.com/beka-birhanu/yetbota/content-service/internal/domain/place"
	domainActivity "github.com/beka-birhanu/yetbota/content-service/internal/domain/placeactivity"
	domainConfirmation "github.com/beka-birhanu/yetbota/content-service/internal/domain/placeconfirmation"
	domainProblem "github.com/beka-birhanu/yetbota/content-service/internal/domain/placeproblem"
	"github.com/beka-birhanu/yetbota/content-service/internal/domain/processors"
	"github.com/google/uuid"
)

func (s *svc) Bookmark(ctx context.Context, ctxSess *ctxRP.Context, req *BookmarkRequest) (*BookmarkResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	if err := utils.AllowAccess(ctxSess); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	if _, err := s.requireReadablePlace(ctx, ctxSess, req.PlaceID); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	if err := s.bookmarkRepo.Add(ctx, nil, ctxSess.UserSession.UserID, req.PlaceID); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	return &BookmarkResponse{}, nil
}

func (s *svc) Unbookmark(ctx context.Context, ctxSess *ctxRP.Context, req *UnbookmarkRequest) (*UnbookmarkResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	if err := utils.AllowAccess(ctxSess); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	if err := s.bookmarkRepo.Delete(ctx, nil, ctxSess.UserSession.UserID, req.PlaceID); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	return &UnbookmarkResponse{}, nil
}

func (s *svc) Confirm(ctx context.Context, ctxSess *ctxRP.Context, req *ConfirmRequest) (*ConfirmResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	if err := utils.AllowAccess(ctxSess); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	place, err := s.requireReadablePlace(ctx, ctxSess, req.PlaceID)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	if place.Status != constants.PlaceStatusVisible && !isAdmin(ctxSess) {
		err = &toddlerr.Error{
			PublicStatusCode:  status.Forbidden,
			ServiceStatusCode: status.Forbidden,
			PublicMessage:     "place is not available for confirmation",
			ServiceMessage:    "place not visible for confirm",
		}
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	now := time.Now().UTC()
	note := null.String{}
	if strings.TrimSpace(req.Note) != "" {
		note = null.StringFrom(strings.TrimSpace(req.Note))
	}
	conf := &domainConfirmation.Confirmation{
		ID:          uuid.NewString(),
		PlaceID:     place.ID,
		UserID:      ctxSess.UserSession.UserID,
		Note:        note,
		ConfirmedAt: now,
	}
	if err := s.confirmationRepo.Upsert(ctx, nil, conf); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	contribs, err := s.contributionRepo.List(ctx, &domainContribution.ListOptions{
		PlaceID:     place.ID,
		OnlyVisible: true,
		Page:        1,
		PageSize:    200,
	})
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	since := now.Add(-time.Duration(s.decayDays) * 24 * time.Hour)
	confirmCount, err := s.confirmationRepo.CountSince(ctx, place.ID, since)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	score := computeConfidence(contribs, confirmCount, &now, s.decayDays, s.verificationBoost)
	if err := s.placeRepo.RecordConfirmation(ctx, nil, place.ID, now, score); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	place.LastConfirmedAt = null.TimeFrom(now)
	place.NeedsCheck = false
	place.ConfidenceScore = score
	if s.activityRepo != nil {
		_ = s.activityRepo.Add(ctx, nil, &domainActivity.Event{
			ID:        uuid.NewString(),
			PlaceID:   place.ID,
			ActorID:   null.StringFrom(ctxSess.UserSession.UserID),
			EventType: domainActivity.EventPlaceConfirmed,
			RefID:     null.StringFrom(conf.ID),
			Summary:   "Place confirmed as still accurate",
			CreatedAt: now,
		})
	}
	if err := s.executor.TriggerPlaceSummarizeWorkflow(ctx, processors.PlaceSummarizeWorkflowInput{PlaceID: place.ID}); err != nil {
		fmt.Printf("unable to trigger place summarize: %s\n", err)
	}
	return &ConfirmResponse{Place: place}, nil
}

func (s *svc) ReportProblem(ctx context.Context, ctxSess *ctxRP.Context, req *ReportProblemRequest) (*ReportProblemResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	if err := utils.AllowAccess(ctxSess); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	if _, err := s.requireReadablePlace(ctx, ctxSess, req.PlaceID); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	if req.Kind == domainProblem.KindDuplicate {
		if _, err := s.placeRepo.Read(ctx, req.DuplicateOf); err != nil {
			ctxSess.SetErrorMessage(err.Error())
			return nil, err
		}
	}

	now := time.Now().UTC()
	report := &domainProblem.Report{
		ID:        uuid.NewString(),
		PlaceID:   req.PlaceID,
		UserID:    ctxSess.UserSession.UserID,
		Kind:      req.Kind,
		Status:    domainProblem.StatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if strings.TrimSpace(req.Note) != "" {
		report.Note = null.StringFrom(strings.TrimSpace(req.Note))
	}
	if req.DuplicateOf != "" {
		report.DuplicateOf = null.StringFrom(req.DuplicateOf)
	}
	if err := s.problemRepo.Add(ctx, nil, report); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	if s.activityRepo != nil {
		_ = s.activityRepo.Add(ctx, nil, &domainActivity.Event{
			ID:        uuid.NewString(),
			PlaceID:   report.PlaceID,
			ActorID:   null.StringFrom(ctxSess.UserSession.UserID),
			EventType: domainActivity.EventProblemReported,
			RefID:     null.StringFrom(report.ID),
			Summary:   "A place problem was reported",
			CreatedAt: now,
		})
	}
	return &ReportProblemResponse{Report: report}, nil
}

func (s *svc) requireReadablePlace(ctx context.Context, ctxSess *ctxRP.Context, placeID string) (*domainPlace.Place, error) {
	place, err := s.placeRepo.Read(ctx, placeID)
	if err != nil {
		return nil, err
	}
	if !isAdmin(ctxSess) && place.Status != constants.PlaceStatusVisible {
		return nil, &toddlerr.Error{
			PublicStatusCode:  status.NotFound,
			ServiceStatusCode: status.NotFound,
			PublicMessage:     "Place not found",
			ServiceMessage:    "place not visible",
		}
	}
	return place, nil
}
