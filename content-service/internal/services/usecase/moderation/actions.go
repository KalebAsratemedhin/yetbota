package moderation

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/beka-birhanu/yetbota/content-service/drivers/constants"
	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
	"github.com/beka-birhanu/yetbota/content-service/drivers/utils"
	ctxRP "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	domainModeration "github.com/beka-birhanu/yetbota/content-service/internal/domain/moderation"
	domainNotification "github.com/beka-birhanu/yetbota/content-service/internal/domain/notification"
	repository "github.com/beka-birhanu/yetbota/content-service/internal/services/repository"
	"github.com/google/uuid"
)

const targetTypeUser = "USER"

func (s *svc) ActOnCase(ctx context.Context, ctxSess *ctxRP.Context, req *ActOnCaseRequest) (*ActOnCaseResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err := utils.AllowAdminOrCSAAccess(ctxSess); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	mc, err := s.repo.GetCase(ctx, req.CaseID)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	if mc.Status != domainModeration.CaseStatusPending {
		err = conflict("case is already resolved")
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	adminID := ctxSess.UserSession.UserID

	var newStatus, resolution, actionType string
	switch req.Action {
	case domainModeration.ActionDelete:
		newStatus, resolution, actionType = domainModeration.CaseStatusResolved, domainModeration.ResolutionDeleted, domainModeration.ActionDelete
	case domainModeration.ActionDismiss:
		newStatus, resolution, actionType = domainModeration.CaseStatusRejected, domainModeration.ResolutionDismissed, domainModeration.ActionDismiss
	case domainModeration.ActionBan:
		newStatus, resolution, actionType = domainModeration.CaseStatusResolved, domainModeration.ResolutionDeleted, domainModeration.ActionBan
	default:
		err = badRequest("invalid action")
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if req.Action == domainModeration.ActionBan && s.banner == nil {
		err = serverError("ban is not configured")
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	ownerID, _, err := s.contentMeta(ctx, mc.ContentType, mc.ContentID)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
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

	var ok bool
	ok, err = s.repo.ResolveCase(ctx, tx, mc.ID, req.Version, newStatus, resolution, adminID)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	if !ok {
		err = conflict("case was modified by someone else, refresh and retry")
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	switch req.Action {
	case domainModeration.ActionDelete, domainModeration.ActionBan:
		if err = s.setContentStatus(ctx, tx, mc.ContentType, mc.ContentID, constants.ModerationStatusRemoved); err != nil {
			ctxSess.SetErrorMessage(err.Error())
			return nil, err
		}
	case domainModeration.ActionDismiss:
		if mc.AutoHidden {
			if err = s.setContentStatus(ctx, tx, mc.ContentType, mc.ContentID, constants.ModerationStatusVisible); err != nil {
				ctxSess.SetErrorMessage(err.Error())
				return nil, err
			}
		}
	}

	action := &domainModeration.ModerationAction{
		ID:         uuid.New().String(),
		CaseID:     mc.ID,
		AdminID:    adminID,
		Action:     actionType,
		TargetType: mc.ContentType,
		TargetID:   mc.ContentID,
		Note:       req.Note,
	}
	if req.Action == domainModeration.ActionBan {
		action.TargetType = targetTypeUser
		action.TargetID = ownerID
	}
	if err = s.repo.CreateAction(ctx, tx, action); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if req.Action == domainModeration.ActionBan {
		if err = s.banner.Ban(ctx, ownerID, req.BanReason, adminID); err != nil {
			ctxSess.SetErrorMessage(err.Error())
			return nil, err
		}
	}

	if err = repository.CommitTx(tx); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	s.notifyActionOutcome(ctx, req.Action, ownerID, mc)

	return &ActOnCaseResponse{
		CaseID:     mc.ID,
		Status:     newStatus,
		Resolution: resolution,
	}, nil
}

func (s *svc) notifyActionOutcome(ctx context.Context, action, ownerID string, mc *domainModeration.ModerationCase) {
	contentType := mc.ContentType

	var title, body string
	switch action {
	case domainModeration.ActionBan:
		title = "Account suspended"
		body = "Your account has been suspended for violating our community guidelines."
	case domainModeration.ActionDelete:
		title = "Content removed"
		body = fmt.Sprintf("Your %s was removed for violating our community guidelines.", contentType)
	case domainModeration.ActionDismiss:
		title = "Report reviewed"
		body = fmt.Sprintf("A report about your %s was reviewed and no action was taken.", contentType)
	default:
		return
	}
	s.notifyUser(ctx, ownerID, title, body)

	if action != domainModeration.ActionDelete && action != domainModeration.ActionDismiss {
		return
	}
	s.notifyReporters(ctx, action, ownerID, mc)
}

func (s *svc) notifyReporters(ctx context.Context, action, ownerID string, mc *domainModeration.ModerationCase) {
	reports, err := s.repo.ListReports(ctx, mc.ContentType, mc.ContentID)
	if err != nil {
		log.Printf("moderation: failed to list reporters for content %s: %v", mc.ContentID, err)
		return
	}

	var title, body string
	if action == domainModeration.ActionDelete {
		title = "Report actioned"
		body = fmt.Sprintf("The %s you reported was removed. Thanks for helping keep the community safe.", mc.ContentType)
	} else {
		title = "Report reviewed"
		body = fmt.Sprintf("We reviewed the %s you reported and found no violation.", mc.ContentType)
	}

	notified := make(map[string]bool)
	for _, r := range reports {
		if r.ReporterID == "" || r.ReporterID == ownerID || notified[r.ReporterID] {
			continue
		}
		notified[r.ReporterID] = true
		s.notifyUser(ctx, r.ReporterID, title, body)
	}
}

func (s *svc) notifyUser(ctx context.Context, userID, title, body string) {
	if userID == "" {
		return
	}

	if s.notificationRepo != nil {
		n := &dbmodels.Notification{
			ID:     uuid.New().String(),
			UserID: userID,
			Title:  title,
			Body:   body,
			SentAt: time.Now(),
		}
		if err := s.notificationRepo.Add(ctx, nil, n); err != nil {
			log.Printf("moderation: failed to persist notification for user %s: %v", userID, err)
		}
	}

	if s.notificationSender == nil || s.userRepo == nil {
		return
	}
	tokens, err := s.userRepo.GetDeviceTokens(ctx, userID)
	if err != nil {
		log.Printf("moderation: failed to fetch device tokens for user %s: %v", userID, err)
		return
	}
	if len(tokens) == 0 {
		return
	}
	if err := s.notificationSender.Send(ctx, &domainNotification.NotificationData{
		Title:     title,
		Body:      body,
		Receivers: tokens,
	}); err != nil {
		log.Printf("moderation: failed to send notification to user %s: %v", userID, err)
	}
}
