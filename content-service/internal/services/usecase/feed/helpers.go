package feed

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
	feedDomain "github.com/beka-birhanu/yetbota/content-service/internal/domain/feed"
	postSvc "github.com/beka-birhanu/yetbota/content-service/internal/services/usecase/post"
)

// collectUnseenFeedItems fetches personal feed items and celebrity feed items for userID,
// merges them by score descending, filters seen ones, and returns up to pageSize items
// plus a cursor for the next page.
func (s *svc) collectUnseenFeedItems(ctx context.Context, userID string, opts *feedDomain.ListOptions, pageSize int, celebIDs []string) ([]*feedDomain.FeedItem, string, error) {
	result := make([]*feedDomain.FeedItem, 0, pageSize)
	var lastItem *feedDomain.FeedItem

	// Celebrity items are score-bounded by opts.MaxScore; fetch once — they don't grow with the loop.
	var celebItems []*feedDomain.FeedItem
	if len(celebIDs) > 0 {
		var err error
		celebItems, err = s.feedRepo.GetCelebrityPosts(ctx, celebIDs, opts)
		if err != nil {
			return nil, "", err
		}
	}

	for {
		items, err := s.feedRepo.List(ctx, userID, opts)
		if err != nil {
			return nil, "", err
		}

		merged := mergeByScore(items, celebItems)

		keys := make([]string, 0, len(merged))
		for _, item := range merged {
			keys = append(keys, seenFeedKey(userID, item.PostID))
		}

		seenMap, err := s.seenCache.Exists(ctx, keys)
		if err != nil {
			return nil, "", err
		}

		for _, item := range merged {
			if seenMap[seenFeedKey(userID, item.PostID)] {
				continue
			}
			result = append(result, item)
			lastItem = item
			if len(result) == pageSize {
				break
			}
		}

		if len(result) >= pageSize || len(items) < opts.Limit {
			break
		}
		opts.Limit *= 2
	}

	var nextCursor string
	if len(result) >= pageSize {
		nextCursor = buildNextCursor(lastItem)
	}
	return result, nextCursor, nil
}

// mergeByScore merges two score-descending slices into one score-descending slice.
func mergeByScore(a, b []*feedDomain.FeedItem) []*feedDomain.FeedItem {
	out := make([]*feedDomain.FeedItem, 0, len(a)+len(b))
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		if a[i].Score >= b[j].Score {
			out = append(out, a[i])
			i++
		} else {
			out = append(out, b[j])
			j++
		}
	}
	out = append(out, a[i:]...)
	out = append(out, b[j:]...)
	return out
}

func orderPosts(unordered []*dbmodels.Post, ids []string) []*dbmodels.Post {
	byID := make(map[string]*dbmodels.Post, len(unordered))
	for _, p := range unordered {
		byID[p.ID] = p
	}
	ordered := make([]*dbmodels.Post, 0, len(ids))
	for _, id := range ids {
		if p, ok := byID[id]; ok {
			ordered = append(ordered, p)
		}
	}
	return ordered
}

func groupPhotosByPost(photos dbmodels.PostPhotoSlice) map[string][]*postSvc.OrderedPhoto {
	m := make(map[string][]*postSvc.OrderedPhoto)
	for _, pp := range photos {
		var photoURL string
		if pp.R != nil && pp.R.Photo != nil {
			photoURL = pp.R.Photo.URL
		}
		m[pp.PostID] = append(m[pp.PostID], &postSvc.OrderedPhoto{
			ID:       pp.PhotoID,
			PostID:   pp.PostID,
			URL:      photoURL,
			Position: pp.Position,
		})
	}
	return m
}

// buildNextCursor encodes score and postID so the next page can start after this item.
func buildNextCursor(item *feedDomain.FeedItem) string {
	if item == nil {
		return ""
	}
	return fmt.Sprintf("cursor:%g", item.Score)
}

// parseCursor decodes a cursor string into (maxScore, afterPostID).
func parseCursor(cursor string) (float64, error) {
	if cursor == "" {
		return 0, nil
	}

	scoreStr := strings.TrimPrefix(cursor, "cursor:")
	score, err := strconv.ParseFloat(scoreStr, 64)
	if err != nil {
		return 0, &toddlerr.Error{
			PublicStatusCode:  status.BadRequest,
			ServiceStatusCode: status.BadRequest,
			PublicMessage:     "Invalid Request",
			ServiceMessage:    fmt.Sprintf("Invalid cursor: %v", err),
		}
	}

	return score, nil
}

// seenFeedKey builds a key for the seen feed DB table.
func seenFeedKey(userID, postID string) string {
	return fmt.Sprintf("%s:%s", userID, postID)
}
