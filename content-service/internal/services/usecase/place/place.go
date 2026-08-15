package place

import (
	"context"
	"fmt"
	"time"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/content-service/drivers/constants"
	ctxRP "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	domainContribution "github.com/beka-birhanu/yetbota/content-service/internal/domain/contribution"
	domainContributionVote "github.com/beka-birhanu/yetbota/content-service/internal/domain/contributionvote"
	domainPlacePhoto "github.com/beka-birhanu/yetbota/content-service/internal/domain/placephoto"
)

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
		var userVote *string
		if v := votes[c.ID]; v != nil {
			vt := v.VoteType
			userVote = &vt
		}
		item := &ContributionView{Contribution: c, UserVote: userVote}
		grouped[c.Kind] = append(grouped[c.Kind], item)
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
