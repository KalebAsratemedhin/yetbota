package moderation

import (
	"context"

	"github.com/beka-birhanu/yetbota/content-service/drivers/utils"
	ctxRP "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	domainModeration "github.com/beka-birhanu/yetbota/content-service/internal/domain/moderation"
	"golang.org/x/sync/errgroup"
)

func snippet(s string) string {
	const max = 140
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max]) + "..."
}

func (s *svc) buildPreview(ctx context.Context, c *domainModeration.ModerationCase) *ContentPreview {
	preview := &ContentPreview{ContentType: c.ContentType, ContentID: c.ContentID}
	switch c.ContentType {
	case domainModeration.ContentTypePost:
		p, err := s.postRepo.Read(ctx, c.ContentID)
		if err != nil {
			preview.Missing = true
			return preview
		}
		preview.AuthorID = p.UserID
		preview.Title = p.Title
		preview.Snippet = snippet(p.Description)
		preview.ModerationStatus = p.ModerationStatus
	case domainModeration.ContentTypeComment:
		cm, err := s.commentRepo.Read(ctx, c.ContentID)
		if err != nil {
			preview.Missing = true
			return preview
		}
		preview.AuthorID = cm.UserID
		preview.Snippet = snippet(cm.Content)
		preview.ModerationStatus = cm.ModerationStatus
	}
	return preview
}

func (s *svc) ListCases(ctx context.Context, ctxSess *ctxRP.Context, req *ListCasesRequest) (*ListCasesResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err := utils.AllowAdminOrCSAAccess(ctxSess); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	opts := &domainModeration.CaseListOptions{
		Status:      req.Status,
		Reason:      req.Reason,
		ContentType: req.ContentType,
		Page:        req.Page,
		PageSize:    req.PageSize,
	}

	errGrp, egCtx := errgroup.WithContext(ctx)
	var cases []*domainModeration.ModerationCase
	var total int64
	errGrp.Go(func() error {
		var err error
		cases, err = s.repo.ListCases(egCtx, opts)
		return err
	})
	errGrp.Go(func() error {
		var err error
		total, err = s.repo.CountCases(egCtx, opts)
		return err
	})
	if err := errGrp.Wait(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	views := make([]*CaseView, 0, len(cases))
	for _, c := range cases {
		views = append(views, &CaseView{Case: c, Preview: s.buildPreview(ctx, c)})
	}

	return &ListCasesResponse{
		Cases:    views,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (s *svc) GetCase(ctx context.Context, ctxSess *ctxRP.Context, req *GetCaseRequest) (*GetCaseResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err := utils.AllowAdminOrCSAAccess(ctxSess); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	mc, err := s.repo.GetCase(ctx, req.ID)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	reports, err := s.repo.ListReports(ctx, mc.ContentType, mc.ContentID)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	resp := &GetCaseResponse{Case: mc, Reports: reports}
	switch mc.ContentType {
	case domainModeration.ContentTypePost:
		if p, e := s.postRepo.Read(ctx, mc.ContentID); e == nil {
			resp.Post = p
		}
	case domainModeration.ContentTypeComment:
		if cm, e := s.commentRepo.Read(ctx, mc.ContentID); e == nil {
			resp.Comment = cm
		}
	}
	return resp, nil
}
