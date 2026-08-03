package admin

import (
	"context"
	"time"

	"github.com/beka-birhanu/yetbota/content-service/drivers/utils"
	domainAdmin "github.com/beka-birhanu/yetbota/content-service/internal/domain/admin"
	ctxRP "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	"golang.org/x/sync/errgroup"
)

func parseAuditTime(s string) (*time.Time, bool) {
	if s == "" {
		return nil, true
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02"} {
		if t, err := time.Parse(layout, s); err == nil {
			return &t, true
		}
	}
	return nil, false
}

func (s *svc) SystemAudit(ctx context.Context, ctxSess *ctxRP.Context, req *SystemAuditRequest) (*SystemAuditResponse, error) {
	if err := utils.AllowAdminOrCSAAccess(ctxSess); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	from, ok := parseAuditTime(req.From)
	if !ok {
		err := badRequest("invalid 'from' date")
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	to, ok := parseAuditTime(req.To)
	if !ok {
		err := badRequest("invalid 'to' date")
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	opts := &domainAdmin.AuditOptions{
		ActionType: req.ActionType,
		Actor:      req.Actor,
		From:       from,
		To:         to,
		Page:       page,
		PageSize:   pageSize,
	}

	errGrp, egCtx := errgroup.WithContext(ctx)
	var entries []*domainAdmin.AuditEntry
	var total int64
	errGrp.Go(func() error {
		var err error
		entries, err = s.repo.ListAudit(egCtx, opts)
		return err
	})
	errGrp.Go(func() error {
		var err error
		total, err = s.repo.CountAudit(egCtx, opts)
		return err
	})
	if err := errGrp.Wait(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	return &SystemAuditResponse{
		Entries:  entries,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}
