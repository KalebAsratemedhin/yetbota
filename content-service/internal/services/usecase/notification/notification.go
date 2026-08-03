package notification

import (
	"context"
	"fmt"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
	"github.com/beka-birhanu/yetbota/content-service/drivers/utils"
	ctxRP "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	dmn "github.com/beka-birhanu/yetbota/content-service/internal/domain/notification"
	"golang.org/x/sync/errgroup"
)

func (s *svc) List(ctx context.Context, ctxSess *ctxRP.Context, req *ListRequest) (*ListResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err := utils.AllowAccess(ctxSess); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	eg, egCtx := errgroup.WithContext(ctx)

	var (
		notifications []*dbmodels.Notification
		count         int64
	)

	eg.Go(func() error {
		var err error
		notifications, err = s.notificationRepo.List(egCtx, &dmn.Filter{
			UserID: ctxSess.UserSession.UserID,
			Unread: req.Unread,
		}, req.Limit, req.Page)
		return err
	})

	eg.Go(func() error {
		var err error
		count, err = s.notificationRepo.Count(egCtx, &dmn.Filter{
			UserID: ctxSess.UserSession.UserID,
			Unread: req.Unread,
		})
		return err
	})

	if err := eg.Wait(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	totalPages := int32(0)
	if req.Limit > 0 {
		totalPages = (int32(count) + req.Limit - 1) / req.Limit
	}

	return &ListResponse{
		Notifications: notifications,
		Pagination: &Pagination{
			Page:       req.Page,
			Total:      int32(count),
			Length:     req.Limit,
			TotalPages: totalPages,
		},
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

	n, err := s.notificationRepo.Read(ctx, req.ID)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return err
	}

	if n.UserID != ctxSess.UserSession.UserID {
		err = &toddlerr.Error{
			PublicStatusCode:  status.NotFound,
			ServiceStatusCode: status.ForbiddenNotEnoughPrivilege,
			PublicMessage:     "notification not found or access denied",
			ServiceMessage:    fmt.Sprintf("delete attempt by non-owner: notification=%s requester=%s", n.ID, ctxSess.UserSession.UserID),
		}
		ctxSess.SetErrorMessage(err.Error())
		return err
	}

	if err := s.notificationRepo.Delete(ctx, nil, n.ID); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return err
	}

	return nil
}

func (s *svc) MarkAsRead(ctx context.Context, ctxSess *ctxRP.Context, req *MarkAsReadRequest) (*MarkAsReadResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err := utils.AllowAccess(ctxSess); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	notifications, err := s.notificationRepo.ReadMany(ctx, req.IDs)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	var (
		success     []string
		failure     []string
		toBeUpdated dbmodels.NotificationSlice
	)

	for _, n := range notifications {
		if n.UserID != ctxSess.UserSession.UserID || n.ReadAt.Valid {
			failure = append(failure, n.ID)
			continue
		}
		toBeUpdated = append(toBeUpdated, n)
		success = append(success, n.ID)
	}

	if len(toBeUpdated) > 0 {
		if err := s.notificationRepo.MarkAsRead(ctx, nil, toBeUpdated); err != nil {
			ctxSess.SetErrorMessage(err.Error())
			return nil, err
		}
	}

	return &MarkAsReadResponse{
		Success: success,
		Failure: failure,
	}, nil
}
