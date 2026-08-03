package place

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aarondl/null/v8"
	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/content-service/drivers/constants"
	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
	"github.com/beka-birhanu/yetbota/content-service/drivers/geotypes"
	"github.com/beka-birhanu/yetbota/content-service/drivers/utils"
	ctxRP "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	domainContribution "github.com/beka-birhanu/yetbota/content-service/internal/domain/contribution"
	domainContributionVote "github.com/beka-birhanu/yetbota/content-service/internal/domain/contributionvote"
	domainPlace "github.com/beka-birhanu/yetbota/content-service/internal/domain/place"
	domainPlacePhoto "github.com/beka-birhanu/yetbota/content-service/internal/domain/placephoto"
	"github.com/beka-birhanu/yetbota/content-service/internal/domain/processors"
	domainScreening "github.com/beka-birhanu/yetbota/content-service/internal/domain/screening"
	repository "github.com/beka-birhanu/yetbota/content-service/internal/services/repository"
	"github.com/twpayne/go-geom"
)

func (s *svc) Add(ctx context.Context, ctxSess *ctxRP.Context, req *AddRequest) (*AddResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	if err := utils.AllowAccess(ctxSess); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	screenReq := domainScreeningRequest(req, ctxSess.UserSession.UserID)
	if err := screenOrReject(ctx, s.screening, &screenReq); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	place := placeFromAddReq(req, ctxSess.UserSession.UserID)
	uploaded, err := s.uploadPhotos(ctx, place.ID, req.Photos)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = s.deleteUploads(ctx, uploaded.photos)
		}
	}()

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

	if err = s.placeRepo.Add(ctx, tx, place); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	if len(uploaded.photos) > 0 {
		if err = s.photoRepo.AddBulk(ctx, tx, uploaded.photos); err != nil {
			ctxSess.SetErrorMessage(err.Error())
			return nil, err
		}
		if err = s.placePhotoRepo.AddBulk(ctx, tx, uploaded.placePhotos); err != nil {
			ctxSess.SetErrorMessage(err.Error())
			return nil, err
		}
	}
	if err = repository.CommitTx(tx); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	ordered, err := s.assembleOrderedPhoto(ctx, uploaded.placePhotos, PhotoResolutionOriginal)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	if err = s.executor.TriggerPlaceIngestWorkflow(ctx, processors.PlaceIngestWorkflowInput{PlaceID: place.ID}); err != nil {
		fmt.Printf("unable to trigger place ingest workflow: %s\n", err)
	}
	return &AddResponse{Place: place, Photos: ordered}, nil
}

func (s *svc) Read(ctx context.Context, ctxSess *ctxRP.Context, req *ReadRequest) (*ReadResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	place, err := s.placeRepo.Read(ctx, req.ID)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	if !isAdmin(ctxSess) && place.Status != constants.PlaceStatusVisible {
		err = &toddlerr.Error{
			PublicStatusCode:  status.NotFound,
			ServiceStatusCode: status.NotFound,
			PublicMessage:     "Place not found",
			ServiceMessage:    fmt.Sprintf("place %s is not visible", place.ID),
		}
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	pps, err := s.placePhotoRepo.List(ctx, &domainPlacePhoto.Options{
		PlaceIDs:  []string{place.ID},
		LoadPhoto: true,
	}, &domainPlacePhoto.SortOptions{
		Field:     domainPlacePhoto.SortFieldPosition,
		Direction: domainPlacePhoto.SortDirectionAsc,
	})
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	ordered, err := s.assembleOrderedPhoto(ctx, pps, req.PhotoResolution)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	contribs, err := s.contributionRepo.List(ctx, &domainContribution.ListOptions{
		PlaceID:     place.ID,
		OnlyVisible: !isAdmin(ctxSess),
		Page:        1,
		PageSize:    200,
	})
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	since := time.Now().UTC().Add(-time.Duration(s.decayDays) * 24 * time.Hour)
	confirmCount, err := s.confirmationRepo.CountSince(ctx, place.ID, since)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	place.ConfidenceScore = computeConfidence(contribs, confirmCount, lastFreshnessTime(place), s.decayDays, s.verificationBoost)

	votes := map[string]*domainContributionVote.Vote(nil)
	isBookmarked := false
	if uid := ctxSess.UserSession.UserID; uid != "" {
		ids := make([]string, 0, len(contribs))
		for _, c := range contribs {
			ids = append(ids, c.ID)
		}
		votes, err = s.contributionVoteRepo.List(ctx, &domainContributionVote.ListOptions{
			UserID:          uid,
			ContributionIDs: ids,
		})
		if err != nil {
			ctxSess.SetErrorMessage(err.Error())
			return nil, err
		}
		isBookmarked, err = s.bookmarkRepo.Exists(ctx, uid, place.ID)
		if err != nil {
			ctxSess.SetErrorMessage(err.Error())
			return nil, err
		}
	}

	grouped := make(map[string][]*ContributionView)
	for _, c := range contribs {
		grouped[c.Kind] = append(grouped[c.Kind], &ContributionView{
			Contribution: c,
			UserVote:     votesByContribution(votes, c.ID),
		})
	}
	return &ReadResponse{
		Place:               place,
		Photos:              ordered,
		ContributionsByKind: grouped,
		Questions:           buildQuestionThreads(grouped),
		IsBookmarked:        isBookmarked,
	}, nil
}

func (s *svc) List(ctx context.Context, ctxSess *ctxRP.Context, req *ListRequest) (*ListResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	opts := req.ListOptions.toDomain(ctxSess)
	places, err := s.placeRepo.List(ctx, opts)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	total, err := s.placeRepo.Count(ctx, opts)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	ids := make([]string, 0, len(places))
	for _, p := range places {
		ids = append(ids, p.ID)
	}
	placePhotos, err := s.placePhotoRepo.List(ctx, &domainPlacePhoto.Options{
		PlaceIDs:  ids,
		LoadPhoto: true,
	}, &domainPlacePhoto.SortOptions{
		Field:     domainPlacePhoto.SortFieldPosition,
		Direction: domainPlacePhoto.SortDirectionAsc,
	})
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	ordered, err := s.assembleOrderedPhoto(ctx, placePhotos, req.PhotoResolution)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	photosByPlace := make(map[string][]*OrderedPhoto, len(ids))
	for _, p := range ordered {
		photosByPlace[p.PlaceID] = append(photosByPlace[p.PlaceID], p)
	}

	return &ListResponse{
		Places:   places,
		Photos:   photosByPlace,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (s *svc) Update(ctx context.Context, ctxSess *ctxRP.Context, req *UpdateRequest) (*UpdateResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	if err := utils.AllowAccess(ctxSess); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	place, err := s.placeRepo.Read(ctx, req.ID)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	if !isAdmin(ctxSess) && place.CreatedBy != ctxSess.UserSession.UserID {
		err = &toddlerr.Error{
			PublicStatusCode:  status.Forbidden,
			ServiceStatusCode: status.Forbidden,
			PublicMessage:     "You do not have access to update this place.",
			ServiceMessage:    fmt.Sprintf("user %s tried to update place %s", ctxSess.UserSession.UserID, place.ID),
		}
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	screenReq := domainScreeningRequestUpdate(req, place, ctxSess.UserSession.UserID)
	if err := screenOrReject(ctx, s.screening, &screenReq); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	var uploaded *uploadPhotosResponse
	var existing []*domainPlacePhoto.PlacePhoto
	var oldPhotoURLs []string
	if len(req.UpsertPhotos) > 0 {
		uploaded, err = s.uploadPhotos(ctx, place.ID, req.UpsertPhotos)
		if err != nil {
			ctxSess.SetErrorMessage(err.Error())
			return nil, err
		}
		defer func() {
			if err != nil {
				_ = s.deleteUploads(ctx, uploaded.photos)
			}
		}()
		existing, err = s.placePhotoRepo.List(ctx, &domainPlacePhoto.Options{
			PlaceIDs:  []string{place.ID},
			LoadPhoto: true,
		}, &domainPlacePhoto.SortOptions{
			Field:     domainPlacePhoto.SortFieldPosition,
			Direction: domainPlacePhoto.SortDirectionAsc,
		})
		if err != nil {
			ctxSess.SetErrorMessage(err.Error())
			return nil, err
		}
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

	if req.Name != nil {
		place.Name = *req.Name
	}
	if req.Description != nil {
		place.Description = *req.Description
	}
	if req.Address != nil {
		trim := strings.TrimSpace(*req.Address)
		place.Address = null.NewString(trim, trim != "")
	}
	if req.Latitude != nil && req.Longitude != nil {
		place.Location = geotypes.NullPoint{
			Point: geom.NewPoint(geom.XY).MustSetCoords([]float64{*req.Longitude, *req.Latitude}),
			Valid: true,
		}
	}
	if req.Status != nil && isAdmin(ctxSess) {
		place.Status = *req.Status
	}
	place.UpdatedAt = time.Now()
	if err = s.placeRepo.Update(ctx, tx, place); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if len(req.UpsertPhotos) > 0 {
		existingByPos := make(map[int]*domainPlacePhoto.PlacePhoto, len(existing))
		for _, ep := range existing {
			existingByPos[ep.Position] = ep
		}
		for i, newPP := range uploaded.placePhotos {
			newPhoto := uploaded.photos[i]
			if current, ok := existingByPos[newPP.Position]; ok {
				if current.PhotoData != nil {
					oldPhotoURLs = append(oldPhotoURLs, current.PhotoData.URL)
				}
				oldPhotoID := current.PhotoID
				if err = s.photoRepo.Add(ctx, tx, newPhoto); err != nil {
					ctxSess.SetErrorMessage(err.Error())
					return nil, err
				}
				current.PhotoID = newPhoto.ID
				if err = s.placePhotoRepo.Update(ctx, tx, current); err != nil {
					ctxSess.SetErrorMessage(err.Error())
					return nil, err
				}
				if err = s.photoRepo.Delete(ctx, tx, oldPhotoID); err != nil {
					ctxSess.SetErrorMessage(err.Error())
					return nil, err
				}
			} else {
				if err = s.photoRepo.Add(ctx, tx, newPhoto); err != nil {
					ctxSess.SetErrorMessage(err.Error())
					return nil, err
				}
				if err = s.placePhotoRepo.Add(ctx, tx, newPP); err != nil {
					ctxSess.SetErrorMessage(err.Error())
					return nil, err
				}
			}
		}
	}

	if err = repository.CommitTx(tx); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	if len(oldPhotoURLs) > 0 {
		oldPhotos := make([]*dbmodels.Photo, 0, len(oldPhotoURLs))
		for _, u := range oldPhotoURLs {
			oldPhotos = append(oldPhotos, &dbmodels.Photo{URL: u})
		}
		if delErr := s.deleteUploads(ctx, oldPhotos); delErr != nil {
			ctxSess.SetErrorMessage(delErr.Error())
		}
	}
	return &UpdateResponse{Place: place}, nil
}

func domainScreeningRequest(req *AddRequest, userID string) domainScreening.Request {
	text := strings.TrimSpace(req.Name)
	if req.Description != "" {
		text = strings.TrimSpace(text + "\n\n" + req.Description)
	}
	return domainScreening.Request{
		Kind:   "place",
		Text:   text,
		UserID: userID,
	}
}

func domainScreeningRequestUpdate(req *UpdateRequest, current *domainPlace.Place, userID string) domainScreening.Request {
	name := current.Name
	if req.Name != nil {
		name = *req.Name
	}
	desc := current.Description
	if req.Description != nil {
		desc = *req.Description
	}
	text := strings.TrimSpace(name)
	if desc != "" {
		text = strings.TrimSpace(text + "\n\n" + desc)
	}
	return domainScreening.Request{
		Kind:    "place",
		Text:    text,
		UserID:  userID,
		PlaceID: current.ID,
	}
}
