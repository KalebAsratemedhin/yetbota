package feed

import (
	"context"
	"maps"
	"slices"

	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
	ctxRP "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	feedDomain "github.com/beka-birhanu/yetbota/content-service/internal/domain/feed"
	"github.com/beka-birhanu/yetbota/content-service/internal/domain/post"
	domainPostphoto "github.com/beka-birhanu/yetbota/content-service/internal/domain/postphoto"
	domainPostvote "github.com/beka-birhanu/yetbota/content-service/internal/domain/postvote"
)

func (s *svc) List(ctx context.Context, ctxSess *ctxRP.Context, req *ListRequest) (*ListResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	count, err := s.feedRepo.Count(ctx, ctxSess.UserSession.UserID)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	if count == 0 {
		// TODO: trigger background feed update
		return &ListResponse{}, nil
	}

	fetchOpts := &feedDomain.ListOptions{Limit: req.PageSize}
	if req.Cursor != "" {
		fetchOpts.MaxScore, err = parseCursor(req.Cursor)
		if err != nil {
			ctxSess.SetErrorMessage(err.Error())
			return nil, err
		}
	}

	celebIDs, err := s.followerRepo.ListCelebrityFollows(ctx, ctxSess.UserSession.UserID, s.celebrityThreshold)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	unseenItems, nextCursor, err := s.collectUnseenFeedItems(ctx, ctxSess.UserSession.UserID, fetchOpts, req.PageSize, celebIDs)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	unseenIDs := make([]string, len(unseenItems))
	for i, item := range unseenItems {
		unseenIDs[i] = item.PostID
	}

	var posts []*dbmodels.Post
	var photos dbmodels.PostPhotoSlice
	var votes map[string]*dbmodels.PostVote
	var followingAuthor map[string]bool
	var saved map[string]bool

	posts, err = s.postRepo.List(ctx, &post.ListOptions{IDs: unseenIDs, PageSize: len(unseenIDs), Page: 1, OnlyVisible: true})
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	photos, err = s.postPhotoRepo.List(ctx, &domainPostphoto.Options{PostIDs: unseenIDs, LoadPhoto: true},
		&domainPostphoto.SortOptions{Field: domainPostphoto.SortFieldPosition, Direction: domainPostphoto.SortDirectionAsc})
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	votes, err = s.postVoteRepo.List(ctx, &domainPostvote.ListOptions{
		UserID:  ctxSess.UserSession.UserID,
		PostIDs: unseenIDs,
	})
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	saved, err = s.savedPostRepo.Exists(ctx, ctxSess.UserSession.UserID, unseenIDs)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	authorIDSet := make(map[string]struct{}, len(posts))
	for _, p := range posts {
		authorIDSet[p.UserID] = struct{}{}
	}
	followingAuthor, err = s.followerRepo.IsFollowing(ctx, ctxSess.UserSession.UserID, slices.Collect(maps.Keys(authorIDSet)))
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	return &ListResponse{
		Posts:           orderPosts(posts, unseenIDs),
		Photos:          groupPhotosByPost(photos),
		Votes:           votes,
		FollowingAuthor: followingAuthor,
		Saved:           saved,
		NextCursor:      nextCursor,
	}, nil
}

func (s *svc) MarkViewed(ctx context.Context, ctxSess *ctxRP.Context, req *MarkViewedRequest) error {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return err
	}

	userID := ctxSess.UserSession.UserID

	if err := s.seenRepo.AddBulk(ctx, userID, req.PostIDs); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return err
	}

	for _, postID := range req.PostIDs {
		key := seenFeedKey(userID, postID)
		if err := s.seenCache.Add(ctx, key, s.seenCacheTTL); err != nil {
			ctxSess.SetErrorMessage(err.Error())
			return err
		}
	}
	return nil
}
