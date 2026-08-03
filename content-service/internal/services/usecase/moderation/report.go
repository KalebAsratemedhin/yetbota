package moderation

import (
	"context"
	"time"

	"github.com/beka-birhanu/yetbota/content-service/drivers/constants"
	"github.com/beka-birhanu/yetbota/content-service/drivers/utils"
	ctxRP "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	domainModeration "github.com/beka-birhanu/yetbota/content-service/internal/domain/moderation"
	repository "github.com/beka-birhanu/yetbota/content-service/internal/services/repository"
	"github.com/google/uuid"
)

func (s *svc) Report(ctx context.Context, ctxSess *ctxRP.Context, req *ReportRequest) (*ReportResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err := utils.AllowAccess(ctxSess); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	reporter := ctxSess.UserSession.UserID

	allowed, err := s.rateLimiter.Allow(ctx, reporter, s.rateLimitMax, s.rateLimitWindow)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	if !allowed {
		err = tooManyRequests("report rate limit exceeded for user " + reporter)
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	ownerID, modStatus, err := s.contentMeta(ctx, req.ContentType, req.ContentID)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	if modStatus == constants.ModerationStatusRemoved {
		err = notFound("content not found")
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	if ownerID == reporter {
		err = badRequest("you cannot report your own content")
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	now := time.Now()
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

	report := &domainModeration.Report{
		ID:          uuid.New().String(),
		ContentType: req.ContentType,
		ContentID:   req.ContentID,
		ReporterID:  reporter,
		Reason:      req.Reason,
		Details:     req.Details,
	}

	var inserted bool
	inserted, err = s.repo.CreateReport(ctx, tx, report)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	if !inserted {
		err = conflict("you have already reported this content")
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	var mc *domainModeration.ModerationCase
	mc, err = s.repo.UpsertCaseOnReport(ctx, tx, uuid.New().String(), req.ContentType, req.ContentID, now)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	autoHidden := mc.AutoHidden
	if !mc.AutoHidden && mc.Status == domainModeration.CaseStatusPending && mc.ReportCount >= s.autoHideThreshold {
		if err = s.setContentStatus(ctx, tx, req.ContentType, req.ContentID, constants.ModerationStatusHidden); err != nil {
			ctxSess.SetErrorMessage(err.Error())
			return nil, err
		}
		if err = s.repo.MarkCaseAutoHidden(ctx, tx, mc.ID); err != nil {
			ctxSess.SetErrorMessage(err.Error())
			return nil, err
		}
		autoHidden = true
	}

	if err = repository.CommitTx(tx); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	return &ReportResponse{
		CaseID:      mc.ID,
		ReportCount: mc.ReportCount,
		Status:      mc.Status,
		AutoHidden:  autoHidden,
	}, nil
}
