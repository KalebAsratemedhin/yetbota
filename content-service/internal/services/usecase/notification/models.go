package notification

import (
	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
)


type ListRequest struct {
	Unread bool  `json:"unread"`
	Limit  int32 `json:"limit"`
	Page   int32 `json:"page"`
}

func (r *ListRequest) Validate() error {
	if r.Page <= 0 {
		r.Page = 1
	}
	if r.Limit <= 0 {
		r.Limit = 20
	}
	return nil
}

type ListResponse struct {
	Notifications dbmodels.NotificationSlice `json:"notifications"`
	Pagination    *Pagination                `json:"pagination"`
}

type Pagination struct {
	Page       int32 `json:"page"`
	Total      int32 `json:"total"`
	Length     int32 `json:"length"`
	TotalPages int32 `json:"total_pages"`
}

type MarkAsReadRequest struct {
	IDs []string `json:"ids" validate:"required,dive,uuid4"`
}

func (r *MarkAsReadRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type MarkAsReadResponse struct {
	Success []string `json:"success"`
	Failure []string `json:"failure"`
}

type DeleteRequest struct {
	ID string `json:"id" validate:"required,uuid4"`
}

func (r *DeleteRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}
