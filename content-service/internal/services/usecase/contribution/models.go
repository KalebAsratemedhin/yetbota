package contribution

import (
	"fmt"
	"strings"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/content-service/drivers/constants"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	domainContribution "github.com/beka-birhanu/yetbota/content-service/internal/domain/contribution"
	domainContributionVote "github.com/beka-birhanu/yetbota/content-service/internal/domain/contributionvote"
)

const (
	KindTip       = "tip"
	KindChange    = "change"
	KindImportant = "important"
	KindQuestion  = "question"
	KindAnswer    = "answer"
)

var allowedKinds = map[string]struct{}{
	KindTip: {},
	KindChange: {},
	KindImportant: {},
	KindQuestion: {},
	KindAnswer: {},
}

type AddRequest struct {
	PlaceID string `validate:"required"`
	Kind    string `validate:"required"`
	Content string `validate:"required,max=10000"`
}

func (r *AddRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	r.Kind = strings.ToLower(strings.TrimSpace(r.Kind))
	if _, ok := allowedKinds[r.Kind]; !ok || r.Kind == KindAnswer {
		return invalidKindErr(r.Kind, "answer is only allowed via /answers")
	}
	return nil
}

type AddAnswerRequest struct {
	ContributionID string `validate:"required"`
	Content        string `validate:"required,max=10000"`
}

func (r *AddAnswerRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type AddResponse struct {
	Contribution *domainContribution.Contribution
}

type VoteRequest struct {
	ContributionID string `validate:"required"`
	VoteType       string `validate:"required,oneof=like dislike"`
}

func (r *VoteRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type VoteResponse struct {
	Upvotes   int
	Downvotes int
}

type ListRequest struct {
	PlaceID   string `validate:"required"`
	Kind      string `validate:"omitempty"`
	Page      int
	PageSize  int
}

func (r *ListRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	if r.Page <= 0 {
		r.Page = 1
	}
	if r.PageSize <= 0 {
		r.PageSize = 20
	}
	if r.PageSize > 100 {
		r.PageSize = 100
	}
	if r.Kind != "" {
		r.Kind = strings.ToLower(strings.TrimSpace(r.Kind))
		if _, ok := allowedKinds[r.Kind]; !ok {
			return invalidKindErr(r.Kind, "")
		}
	}
	return nil
}

type ContributionItem struct {
	Contribution *domainContribution.Contribution
	UserVote     *domainContributionVote.Vote
}

type ListResponse struct {
	Items    []*ContributionItem
	Total    int64
	Page     int
	PageSize int
}

type AcceptAnswerRequest struct {
	QuestionID string `validate:"required"`
	AnswerID   string `validate:"required"`
}

func (r *AcceptAnswerRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type AcceptAnswerResponse struct {
	Answer *domainContribution.Contribution
}

type ListOpenQuestionsRequest struct {
	NearLat  *float64
	NearLon  *float64
	RadiusKm *float64
	Page     int
	PageSize int
}

func (r *ListOpenQuestionsRequest) Validate() error {
	if r.Page <= 0 {
		r.Page = 1
	}
	if r.PageSize <= 0 {
		r.PageSize = 20
	}
	if r.PageSize > 100 {
		r.PageSize = 100
	}
	hasNear := r.NearLat != nil || r.NearLon != nil || r.RadiusKm != nil
	if hasNear && (r.NearLat == nil || r.NearLon == nil || r.RadiusKm == nil) {
		return &toddlerr.Error{
			PublicStatusCode:  status.BadRequest,
			ServiceStatusCode: status.BadRequest,
			PublicMessage:     "near_lat, near_lon, and radius_km are required together",
			ServiceMessage:    "incomplete near filter",
		}
	}
	if r.RadiusKm != nil && *r.RadiusKm <= 0 {
		return &toddlerr.Error{
			PublicStatusCode:  status.BadRequest,
			ServiceStatusCode: status.BadRequest,
			PublicMessage:     "radius_km must be positive",
			ServiceMessage:    "invalid radius_km",
		}
	}
	return nil
}

type ListOpenQuestionsResponse struct {
	Questions []*domainContribution.Contribution
	Total     int64
	Page      int
	PageSize  int
}

type ScreeningHook interface {
	ShouldScreen(kind string) bool
}

func invalidKindErr(kind string, detail string) error {
	msg := fmt.Sprintf("invalid contribution kind: %s", kind)
	if detail != "" {
		msg += ", " + detail
	}
	return &toddlerr.Error{
		PublicStatusCode:  status.BadRequest,
		ServiceStatusCode: status.BadRequest,
		PublicMessage:     "invalid contribution kind",
		ServiceMessage:    msg,
	}
}

func visibleOnlyForRole(role string) bool {
	return role != constants.RoleAdmin
}
