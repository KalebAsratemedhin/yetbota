package feed

import (
	"net/http"
	"regexp"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	postSvc "github.com/beka-birhanu/yetbota/content-service/internal/services/usecase/post"
)

var cursorRegex = regexp.MustCompile(`^cursor:[0-9]+(?:\.[0-9]+)?$`)

type ListRequest struct {
	Cursor   string `validate:"omitempty"`
	PageSize int    `validate:"required,min=1,max=100"`
}

func (r *ListRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	if r.Cursor != "" {
		if !cursorRegex.MatchString(r.Cursor) {
			return &toddlerr.Error{
				PublicStatusCode:  http.StatusBadRequest,
				ServiceStatusCode: http.StatusBadRequest,
				PublicMessage:     "invalid cursor",
				ServiceMessage:    "invalid cursor",
			}
		}
	}
	return nil
}

type ListResponse struct {
	Posts           []*dbmodels.Post
	Photos          map[string][]*postSvc.OrderedPhoto
	Votes           map[string]*dbmodels.PostVote
	FollowingAuthor map[string]bool
	Saved           map[string]bool
	NextCursor      string
}

type MarkViewedRequest struct {
	PostIDs []string `validate:"required,min=1,dive,uuid"`
}

func (r *MarkViewedRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}
