package post

import (
	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	ctxRP "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	domainPost "github.com/beka-birhanu/yetbota/content-service/internal/domain/post"
)

type OrderedPhotoUpload struct {
	Photo    []byte `mask:"true"`
	Position int
}

type OrderedPhoto struct {
	ID       string
	PostID   string
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

// Add

type AddRequest struct {
	Title          string   `validate:"required"`
	Description    string   `validate:"required"`
	Tags           []string `validate:"dive,min=1,max=20"`
	IsQuestion     bool
	Photos         []*OrderedPhotoUpload
	Latitude       float64
	Longitude      float64
	Address        string `validate:"omitempty,max=255"`
	AttachedPostID string `validate:"omitempty,uuid"`
}

func (r *AddRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type AddResponse struct {
	Post   *dbmodels.Post
	Photos []*OrderedPhoto
}

// Read

type ReadRequest struct {
	ID              string          `validate:"required"`
	PhotoResolution PhotoResolution `validate:"required,oneof=ORIGINAL MOBILE WEB"`
}

func (r *ReadRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type ReadResponse struct {
	Post   *dbmodels.Post
	Photos []*OrderedPhoto
	// AttachedPost is the post this one references via attached_post_id, when set.
	AttachedPost       *dbmodels.Post
	AttachedPostPhotos []*OrderedPhoto
	// Viewer state, populated only for authenticated callers.
	Vote            *dbmodels.PostVote
	FollowingAuthor *bool
	Saved           *bool
}

// Update

type UpdateRequest struct {
	ID           string   `validate:"required"`
	Title        string   `validate:"required"`
	Description  string   `validate:"required"`
	Tags         []string `validate:"dive,min=1,max=20"`
	UpsertPhotos []*OrderedPhotoUpload
	Latitude     float64 `validate:"omitempty,min=-90,max=90"`
	Longitude    float64 `validate:"omitempty,min=-180,max=180"`
	Address      string  `validate:"omitempty,max=255"`
	// AttachedPostID is presence-aware: nil leaves the existing attachment
	// untouched, "" clears it, a uuid sets/replaces it.
	AttachedPostID *string `validate:"omitempty,uuid"`
}

func (r *UpdateRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type UpdateResponse struct {
	Post *dbmodels.Post
}

type uploadPhotosResponse struct {
	photos     dbmodels.PhotoSlice
	postPhotos dbmodels.PostPhotoSlice
}

type PostVoteRequest struct {
	PostID   string `validate:"required"`
	VoteType string `validate:"required" oneof:"like dislike"`
}

func (r *PostVoteRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}

	return nil
}

type PostVoteResponse struct {
	Likes    int
	Dislikes int
}

// Save

type SaveRequest struct {
	PostID string `validate:"required"`
}

func (r *SaveRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type SaveResponse struct{}

// Unsave

type UnsaveRequest struct {
	PostID string `validate:"required"`
}

func (r *UnsaveRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type UnsaveResponse struct{}

// List

type ListOptions struct {
	IDs            []string
	UserID         string
	Tags           []string
	IsQuestion     *bool
	AttachedPostID string
	Search         string
	NearLat        *float64
	NearLon        *float64
	RadiusKm       *float64
	SavedOnly      bool
	NonVisible     bool
	SortField      domainPost.ListSortField
	SortDir        domainPost.ListSortDir
	Page           int
	PageSize       int
}

func (r *ListOptions) toDomain(ctxSession *ctxRP.Context) *domainPost.ListOptions {
	opts := &domainPost.ListOptions{
		IDs:            r.IDs,
		UserID:         r.UserID,
		Tags:           r.Tags,
		IsQuestion:     r.IsQuestion,
		AttachedPostID: r.AttachedPostID,
		Search:         r.Search,
		NearLat:        r.NearLat,
		NearLon:        r.NearLon,
		RadiusKm:       r.RadiusKm,
		SortField:      r.SortField,
		SortDir:        r.SortDir,
		Page:           r.Page,
		PageSize:       r.PageSize,
	}
	if ctxSession != nil && r.SavedOnly && ctxSession.UserSession.UserID != "" {
		opts.SavedBy = &ctxSession.UserSession.UserID
	}
	if r.NonVisible {
		opts.NonVisibleOnly = true
	} else {
		opts.OnlyVisible = true
	}

	return opts
}

type ListRequest struct {
	ListOptions
	PhotoResolution PhotoResolution
}

func (r *ListRequest) Validate() error {
	if r.Page <= 0 {
		r.Page = 1
	}
	if r.PageSize <= 0 {
		r.PageSize = 20
	}
	if r.PhotoResolution == PhotoResolutionUnspecified {
		r.PhotoResolution = PhotoResolutionMobile
	}
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type ListResponse struct {
	Posts           []*dbmodels.Post
	Photos          map[string][]*OrderedPhoto
	Votes           map[string]*dbmodels.PostVote
	FollowingAuthor map[string]bool
	Saved           map[string]bool
	Total           int64
	Page            int
	PageSize        int
}
