package place

import (
	"fmt"
	"strings"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/content-service/drivers/constants"
	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	ctxRP "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	domainContribution "github.com/beka-birhanu/yetbota/content-service/internal/domain/contribution"
	domainPlace "github.com/beka-birhanu/yetbota/content-service/internal/domain/place"
	domainPlacePhoto "github.com/beka-birhanu/yetbota/content-service/internal/domain/placephoto"
	domainProblem "github.com/beka-birhanu/yetbota/content-service/internal/domain/placeproblem"
)

type OrderedPhotoUpload struct {
	Photo    []byte `mask:"true"`
	Position int
}

type OrderedPhoto struct {
	ID       string
	PlaceID  string
	URL      string
	Position int
}

type PhotoResolution string

const (
	PhotoResolutionUnspecified PhotoResolution = ""
	PhotoResolutionOriginal    PhotoResolution = "ORIGINAL"
	PhotoResolutionMobile      PhotoResolution = "MOBILE"
	PhotoResolutionWeb         PhotoResolution = "WEB"
)

type AddRequest struct {
	Name        string `validate:"required,max=255"`
	Description string `validate:"omitempty,max=5000"`
	Address     string `validate:"omitempty,max=255"`
	Photos      []*OrderedPhotoUpload
	Latitude    float64
	Longitude   float64
}

func (r *AddRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type AddResponse struct {
	Place  *domainPlace.Place
	Photos []*OrderedPhoto
}

type ReadRequest struct {
	ID              string          `validate:"required"`
	PhotoResolution PhotoResolution `validate:"required,oneof=ORIGINAL MOBILE WEB"`
}

func (r *ReadRequest) Validate() error {
	if r.PhotoResolution == PhotoResolutionUnspecified {
		r.PhotoResolution = PhotoResolutionOriginal
	}
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type ContributionView struct {
	Contribution *domainContribution.Contribution
	UserVote     *string
}

type QuestionThread struct {
	Question       *ContributionView
	Answers        []*ContributionView
	AcceptedAnswer *ContributionView
	IsResolved     bool
}

type ReadResponse struct {
	Place               *domainPlace.Place
	Photos              []*OrderedPhoto
	ContributionsByKind map[string][]*ContributionView
	Questions           []*QuestionThread
	IsBookmarked        bool
}

type UpdateRequest struct {
	ID           string              `validate:"required"`
	Name         *string             `validate:"omitempty,max=255"`
	Description  *string             `validate:"omitempty,max=5000"`
	Address      *string             `validate:"omitempty,max=255"`
	UpsertPhotos []*OrderedPhotoUpload
	Latitude     *float64            `validate:"omitempty,min=-90,max=90"`
	Longitude    *float64            `validate:"omitempty,min=-180,max=180"`
	Status       *string             `validate:"omitempty,oneof=PENDING VISIBLE HIDDEN REMOVED DUPLICATE"`
}

func (r *UpdateRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	if (r.Latitude != nil && r.Longitude == nil) || (r.Latitude == nil && r.Longitude != nil) {
		return &toddlerr.Error{
			PublicStatusCode:  status.BadRequest,
			ServiceStatusCode: status.BadRequest,
			PublicMessage:     "both latitude and longitude are required together",
			ServiceMessage:    "only one coordinate provided",
		}
	}
	return nil
}

type UpdateResponse struct {
	Place *domainPlace.Place
}

type ListOptions struct {
	NearLat    *float64
	NearLon    *float64
	RadiusKm   *float64
	Search     string
	Page       int
	PageSize   int
	Status     string
	NeedsCheck *bool
}

func (r *ListOptions) toDomain(ctxSess *ctxRP.Context) *domainPlace.ListOptions {
	out := &domainPlace.ListOptions{
		NearLat:    r.NearLat,
		NearLon:    r.NearLon,
		RadiusKm:   r.RadiusKm,
		Search:     strings.TrimSpace(r.Search),
		Page:       r.Page,
		PageSize:   r.PageSize,
		NeedsCheck: r.NeedsCheck,
	}
	if r.Status != "" {
		s := strings.ToUpper(strings.TrimSpace(r.Status))
		out.Status = &s
	} else if ctxSess == nil || !isAdmin(ctxSess) {
		s := constants.PlaceStatusVisible
		out.Status = &s
	}
	return out
}

type ListRequest struct {
	ListOptions
	PhotoResolution PhotoResolution `validate:"omitempty,oneof=ORIGINAL MOBILE WEB"`
}

func (r *ListRequest) Validate() error {
	if r.Page <= 0 {
		r.Page = 1
	}
	if r.PageSize <= 0 {
		r.PageSize = 20
	}
	if r.PageSize > 100 {
		r.PageSize = 100
	}
	if r.PhotoResolution == PhotoResolutionUnspecified {
		r.PhotoResolution = PhotoResolutionMobile
	}
	if r.Status != "" {
		normalized := strings.ToUpper(strings.TrimSpace(r.Status))
		switch normalized {
		case constants.PlaceStatusPending, constants.PlaceStatusVisible, constants.PlaceStatusHidden, constants.PlaceStatusRemoved, constants.PlaceStatusDuplicate:
			r.Status = normalized
		default:
			return &toddlerr.Error{
				PublicStatusCode:  status.BadRequest,
				ServiceStatusCode: status.BadRequest,
				PublicMessage:     "invalid status filter",
				ServiceMessage:    fmt.Sprintf("unsupported status %q", r.Status),
			}
		}
	}
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type ListResponse struct {
	Places   []*domainPlace.Place
	Photos   map[string][]*OrderedPhoto
	Total    int64
	Page     int
	PageSize int
}

type BookmarkRequest struct {
	PlaceID string `validate:"required"`
}

func (r *BookmarkRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type BookmarkResponse struct{}

type UnbookmarkRequest struct {
	PlaceID string `validate:"required"`
}

func (r *UnbookmarkRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type UnbookmarkResponse struct{}

type ConfirmRequest struct {
	PlaceID string `validate:"required"`
	Note    string `validate:"omitempty,max=2000"`
}

func (r *ConfirmRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type ConfirmResponse struct {
	Place *domainPlace.Place
}

type ReportProblemRequest struct {
	PlaceID     string `validate:"required"`
	Kind        string `validate:"required,oneof=PERMANENTLY_CLOSED DOES_NOT_EXIST DUPLICATE WRONG_LOCATION"`
	Note        string `validate:"omitempty,max=2000"`
	DuplicateOf string `validate:"omitempty,uuid"`
}

func (r *ReportProblemRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	r.Kind = strings.ToUpper(strings.TrimSpace(r.Kind))
	if r.Kind == "DUPLICATE" && strings.TrimSpace(r.DuplicateOf) == "" {
		return &toddlerr.Error{
			PublicStatusCode:  status.BadRequest,
			ServiceStatusCode: status.BadRequest,
			PublicMessage:     "duplicate_of is required for DUPLICATE reports",
			ServiceMessage:    "missing duplicate_of",
		}
	}
	return nil
}

type ReportProblemResponse struct {
	Report *domainProblem.Report
}

type SummaryCitation struct {
	ContributionID string `json:"contribution_id"`
	Kind           string `json:"kind"`
	Excerpt        string `json:"excerpt"`
}

type UpdateSummaryRequest struct {
	PlaceID   string `validate:"required"`
	SummaryText string
	Citations []SummaryCitation
}

func (r *UpdateSummaryRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type UpdateSummaryResponse struct {
	Place *domainPlace.Place
}

type uploadPhotosResponse struct {
	photos      dbmodels.PhotoSlice
	placePhotos []*domainPlacePhoto.PlacePhoto
}
