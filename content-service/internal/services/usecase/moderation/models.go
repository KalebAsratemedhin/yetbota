package moderation

import (
	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/yetbota/content-service/drivers/constants"
	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	domainModeration "github.com/beka-birhanu/yetbota/content-service/internal/domain/moderation"
)

type ReportRequest struct {
	ContentType string `validate:"required,oneof=POST COMMENT"`
	ContentID   string `validate:"required,uuid"`
	Reason      string `validate:"required,oneof=SPAM OFFENSIVE INCORRECT OTHER"`
	Details     string
}

func (r *ReportRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	if r.Reason == domainModeration.ReasonOther && r.Details == "" {
		return badRequest("details are required when reason is OTHER")
	}
	return nil
}

type ReportResponse struct {
	CaseID      string
	ReportCount int
	Status      string
	AutoHidden  bool
}

type ListCasesRequest struct {
	Status      string `validate:"omitempty,oneof=PENDING RESOLVED REJECTED"`
	Reason      string `validate:"omitempty,oneof=SPAM OFFENSIVE INCORRECT OTHER"`
	ContentType string `validate:"omitempty,oneof=POST COMMENT"`
	Page        int
	PageSize    int
}

func (r *ListCasesRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	if r.Status == "" {
		r.Status = domainModeration.CaseStatusPending
	}
	if r.Page <= 0 {
		r.Page = 1
	}
	if r.PageSize <= 0 {
		r.PageSize = constants.DefaultPaginationLength
	}
	if r.PageSize > constants.MaxPaginationLength {
		r.PageSize = constants.MaxPaginationLength
	}
	return nil
}

type ContentPreview struct {
	ContentType      string
	ContentID        string
	AuthorID         string
	Title            string
	Snippet          string
	ModerationStatus string
	Missing          bool
}

type CaseView struct {
	Case    *domainModeration.ModerationCase
	Preview *ContentPreview
}

type ListCasesResponse struct {
	Cases    []*CaseView
	Total    int64
	Page     int
	PageSize int
}

type GetCaseRequest struct {
	ID string `validate:"required,uuid"`
}

func (r *GetCaseRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type GetCaseResponse struct {
	Case    *domainModeration.ModerationCase
	Reports []*domainModeration.Report
	Post    *dbmodels.Post
	Comment *dbmodels.Comment
}

type ActOnCaseRequest struct {
	CaseID    string `validate:"required,uuid"`
	Action    string `validate:"required,oneof=DELETE DISMISS BAN"`
	Note      string
	BanReason string
	Version   int
}

func (r *ActOnCaseRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type ActOnCaseResponse struct {
	CaseID     string
	Status     string
	Resolution string
}
