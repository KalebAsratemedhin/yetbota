package post

import (
	"context"
	"fmt"
	"maps"
	"slices"

	"github.com/aarondl/null/v8"
	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
	"github.com/beka-birhanu/yetbota/content-service/drivers/geotypes"
	"github.com/beka-birhanu/yetbota/content-service/drivers/utils"
	ctxRP "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	domainPostphoto "github.com/beka-birhanu/yetbota/content-service/internal/domain/postphoto"
	domainPostvote "github.com/beka-birhanu/yetbota/content-service/internal/domain/postvote"
	"github.com/beka-birhanu/yetbota/content-service/internal/domain/processors"
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

	if err := s.validateAttachedPost(ctx, req.AttachedPostID, "", req.IsQuestion); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	post := postFromAddReq(req)
	post.UserID = ctxSess.UserSession.UserID

	uploaded, err := s.uploadPhotos(ctx, post.ID, req.Photos)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	defer func() {
		if err != nil {
			err = s.deleteUploads(ctx, uploaded.photos)
			if err != nil {
				ctxSess.SetErrorMessage(ctxSess.ErrorMessage + "\n" + err.Error())
			}
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

	err = s.postRepo.Add(ctx, tx, post)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	err = s.photoRepo.AddBulk(ctx, tx, uploaded.photos)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	err = s.postPhotoRepo.AddBulk(ctx, tx, uploaded.postPhotos)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err = repository.CommitTx(tx); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	orderedPhotos, err := s.assembleOrderedPhoto(ctx, uploaded.postPhotos, PhotoResolutionOriginal)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err = s.executor.TriggerNewPostWorkflow(ctx, processors.NewPostWorkflowInput{
		PostID: post.ID,
	}); err != nil {
		fmt.Printf("Unable to trigger new post workflow: %s\n", err)
	}

	if err = s.scoringStream.Add(ctx, map[string]any{
		"postID":    post.ID,
		"createdAt": post.CreatedAt.Unix(),
	}); err != nil {
		fmt.Printf("Unable to publish to scoring stream: %s\n", err)
	}

	return &AddResponse{Post: post, Photos: orderedPhotos}, nil
}

func (s *svc) Read(ctx context.Context, ctxSess *ctxRP.Context, req *ReadRequest) (*ReadResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	post, err := s.postRepo.Read(ctx, req.ID)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err := utils.EnsureVisible(ctxSess, post.ModerationStatus, "post", post.ID); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	photos, err := s.postPhotoRepo.List(ctx, &domainPostphoto.Options{
		PostIDs:   []string{post.ID},
		LoadPhoto: true,
	}, &domainPostphoto.SortOptions{
		Field:     domainPostphoto.SortFieldPosition,
		Direction: domainPostphoto.SortDirectionAsc,
	})
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	orderedPhotos, err := s.assembleOrderedPhoto(ctx, photos, req.PhotoResolution)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	resp := &ReadResponse{Post: post, Photos: orderedPhotos}

	if post.AttachedPostID.Valid && post.AttachedPostID.String != "" {
		attached, attachedErr := s.postRepo.Read(ctx, post.AttachedPostID.String)
		if attachedErr == nil && utils.EnsureVisible(ctxSess, attached.ModerationStatus, "post", attached.ID) == nil {
			attachedPhotos, photoErr := s.postPhotos(ctx, attached.ID, req.PhotoResolution)
			if photoErr != nil {
				ctxSess.SetErrorMessage(photoErr.Error())
				return nil, photoErr
			}
			resp.AttachedPost = attached
			resp.AttachedPostPhotos = attachedPhotos
		}
	}

	if uid := ctxSess.UserSession.UserID; uid != "" {
		votes, err := s.postVoteRepo.List(ctx, &domainPostvote.ListOptions{
			UserID:  uid,
			PostIDs: []string{post.ID},
		})
		if err != nil {
			ctxSess.SetErrorMessage(err.Error())
			return nil, err
		}
		resp.Vote = votes[post.ID]

		following, err := s.followerRepo.IsFollowing(ctx, uid, []string{post.UserID})
		if err != nil {
			ctxSess.SetErrorMessage(err.Error())
			return nil, err
		}
		fa := following[post.UserID]
		resp.FollowingAuthor = &fa

		saved, err := s.savedPostRepo.Exists(ctx, uid, []string{post.ID})
		if err != nil {
			ctxSess.SetErrorMessage(err.Error())
			return nil, err
		}
		sv := saved[post.ID]
		resp.Saved = &sv
	}

	return resp, nil
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

	post, err := s.postRepo.Read(ctx, req.ID)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if req.AttachedPostID != nil {
		if err = s.validateAttachedPost(ctx, *req.AttachedPostID, post.ID, post.IsQuestion); err != nil {
			ctxSess.SetErrorMessage(err.Error())
			return nil, err
		}
	}

	var uploaded *uploadPhotosResponse
	var positionMap map[int]*dbmodels.PostPhoto
	var oldPhotoURLs []string

	if len(req.UpsertPhotos) > 0 {
		uploaded, err = s.uploadPhotos(ctx, post.ID, req.UpsertPhotos)
		if err != nil {
			ctxSess.SetErrorMessage(err.Error())
			return nil, err
		}
		defer func() {
			if err != nil {
				_ = s.deleteUploads(ctx, uploaded.photos)
			}
		}()

		existing, listErr := s.postPhotoRepo.List(ctx, &domainPostphoto.Options{
			PostIDs:   []string{post.ID},
			LoadPhoto: true,
		}, nil)
		if listErr != nil {
			err = listErr
			ctxSess.SetErrorMessage(err.Error())
			return nil, err
		}

		positionMap = make(map[int]*dbmodels.PostPhoto, len(existing))
		for _, pp := range existing {
			positionMap[pp.Position] = pp
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

	post.Title = req.Title
	post.Description = req.Description
	post.Tags = req.Tags
	post.Location = geotypes.NullPoint{Point: geom.NewPoint(geom.XY).MustSetCoords([]float64{req.Longitude, req.Latitude}), Valid: true}
	post.Address = null.NewString(req.Address, req.Address != "")
	if req.AttachedPostID != nil {
		post.AttachedPostID = null.NewString(*req.AttachedPostID, *req.AttachedPostID != "")
	}

	err = s.postRepo.Update(ctx, tx, post)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if len(req.UpsertPhotos) > 0 {
		toInsertPhotos := make(dbmodels.PhotoSlice, 0)
		toInsertPostPhotos := make(dbmodels.PostPhotoSlice, 0)

		for i, newPP := range uploaded.postPhotos {
			newPhoto := uploaded.photos[i]

			if existingPP, exists := positionMap[newPP.Position]; exists {
				if existingPP.R != nil && existingPP.R.Photo != nil {
					oldPhotoURLs = append(oldPhotoURLs, existingPP.R.Photo.URL)
				}
				oldPhotoID := existingPP.PhotoID

				err = s.photoRepo.Add(ctx, tx, newPhoto)
				if err != nil {
					ctxSess.SetErrorMessage(err.Error())
					return nil, err
				}

				existingPP.PhotoID = newPhoto.ID
				err = s.postPhotoRepo.Update(ctx, tx, existingPP)
				if err != nil {
					ctxSess.SetErrorMessage(err.Error())
					return nil, err
				}

				err = s.photoRepo.Delete(ctx, tx, oldPhotoID)
				if err != nil {
					ctxSess.SetErrorMessage(err.Error())
					return nil, err
				}
			} else {
				toInsertPhotos = append(toInsertPhotos, newPhoto)
				toInsertPostPhotos = append(toInsertPostPhotos, newPP)
			}
		}

		if len(toInsertPhotos) > 0 {
			err = s.photoRepo.AddBulk(ctx, tx, toInsertPhotos)
			if err != nil {
				ctxSess.SetErrorMessage(err.Error())
				return nil, err
			}
			err = s.postPhotoRepo.AddBulk(ctx, tx, toInsertPostPhotos)
			if err != nil {
				ctxSess.SetErrorMessage(err.Error())
				return nil, err
			}
		}
	}

	if err = repository.CommitTx(tx); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if len(oldPhotoURLs) > 0 {
		oldPhotos := make(dbmodels.PhotoSlice, 0, len(oldPhotoURLs))
		for _, url := range oldPhotoURLs {
			oldPhotos = append(oldPhotos, &dbmodels.Photo{URL: url})
		}
		err = s.deleteUploads(ctx, oldPhotos)
		if err != nil {
			ctxSess.SetErrorMessage(err.Error())
		}
	}

	return &UpdateResponse{Post: post}, nil
}

func (s *svc) List(ctx context.Context, ctxSess *ctxRP.Context, req *ListRequest) (*ListResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	listOpts := req.toDomain(ctxSess)

	posts, err := s.postRepo.List(ctx, listOpts)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	total, err := s.postRepo.Count(ctx, listOpts)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	postIDs := make([]string, len(posts))
	authorIDSet := make(map[string]struct{}, len(posts))
	for i, post := range posts {
		postIDs[i] = post.ID
		authorIDSet[post.UserID] = struct{}{}
	}
	authorIDs := slices.Collect(maps.Keys(authorIDSet))

	var photos dbmodels.PostPhotoSlice
	var votes map[string]*dbmodels.PostVote
	var followingAuthor map[string]bool
	var saved map[string]bool

	photos, err = s.postPhotoRepo.List(ctx, &domainPostphoto.Options{
		LoadPhoto: true,
		PostIDs:   postIDs,
	}, &domainPostphoto.SortOptions{
		Field:     domainPostphoto.SortFieldPosition,
		Direction: domainPostphoto.SortDirectionAsc,
	})
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	votes, err = s.postVoteRepo.List(ctx, &domainPostvote.ListOptions{
		UserID:  ctxSess.UserSession.UserID,
		PostIDs: postIDs,
	})
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	followingAuthor, err = s.followerRepo.IsFollowing(ctx, ctxSess.UserSession.UserID, authorIDs)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	saved, err = s.savedPostRepo.Exists(ctx, ctxSess.UserSession.UserID, postIDs)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	ordered, err := s.assembleOrderedPhoto(ctx, photos, req.PhotoResolution)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	photosByPost := make(map[string][]*OrderedPhoto, len(posts))
	for _, op := range ordered {
		photosByPost[op.PostID] = append(photosByPost[op.PostID], op)
	}

	return &ListResponse{
		Posts:           posts,
		Photos:          photosByPost,
		Votes:           votes,
		FollowingAuthor: followingAuthor,
		Saved:           saved,
		Total:           total,
		Page:            req.Page,
		PageSize:        req.PageSize,
	}, nil
}

func (s *svc) Vote(ctx context.Context, ctxSess *ctxRP.Context, req *PostVoteRequest) (*PostVoteResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err := utils.AllowAccess(ctxSess); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	post, err := s.postRepo.Read(ctx, req.PostID)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	exists, err := s.postVoteRepo.Exists(ctx, ctxSess.UserSession.UserID, req.PostID)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	var vote *dbmodels.PostVote
	if exists {
		vote, err = s.postVoteRepo.Read(ctx, ctxSess.UserSession.UserID, req.PostID)
		if err != nil {
			ctxSess.SetErrorMessage(err.Error())
			return nil, err
		}
	}

	if exists && vote.VoteType == req.VoteType {
		return &PostVoteResponse{Likes: post.Likes, Dislikes: post.Dislikes}, nil
	}

	var likesDelta, dislikesDelta int
	switch req.VoteType {
	case dbmodels.PostVoteTypeLike:
		likesDelta = 1
		if exists {
			vote.VoteType = dbmodels.PostVoteTypeLike
			dislikesDelta = -1
		}
	case dbmodels.PostVoteTypeDislike:
		dislikesDelta = 1
		if exists {
			vote.VoteType = dbmodels.PostVoteTypeDislike
			likesDelta = -1
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

	if exists {
		err = s.postVoteRepo.Update(ctx, tx, vote)
	} else {
		err = s.postVoteRepo.Add(ctx, tx, &dbmodels.PostVote{
			UserID:   ctxSess.UserSession.UserID,
			PostID:   req.PostID,
			VoteType: req.VoteType,
		})
	}
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	err = s.postVoteRepo.UpdateCounts(ctx, tx, req.PostID, likesDelta, dislikesDelta, post.Likes, post.Dislikes)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err = repository.CommitTx(tx); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err = s.executor.TriggerFeedUpdateWorkflow(ctx, processors.FeedUpdateWorkflowInput{
		PostID:       post.ID,
		InteractorID: ctxSess.UserSession.UserID,
	}); err != nil {
		fmt.Printf("Unable to trigger feed update workflow: %s\n", err)
	}

	return &PostVoteResponse{Likes: post.Likes + likesDelta, Dislikes: post.Dislikes + dislikesDelta}, nil
}
