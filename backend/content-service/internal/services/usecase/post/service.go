package post

import (
	"context"
	"encoding/base64"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/yetbota/content-service/internal/domain/apperr"
	"github.com/yetbota/content-service/internal/domain/media"
	domain "github.com/yetbota/content-service/internal/domain/post"
)

var hashtagRe = regexp.MustCompile(`^[a-z0-9_]{1,50}$`)

type FollowingStore interface {
	FollowingIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
}

type ReputationNotifier interface {
	Enabled() bool
	SyncScore(ctx context.Context, userID uuid.UUID, score, postsCreated, likesReceived int64) error
}

type FeedConfig struct {
	Gravity          float64
	CommentWeight    float64
	ShareWeight      float64
	BookmarkWeight   float64
	FollowBoost      float64
	GeoD0Km          float64
	NearbyRadiusKm   float64
	CandidateLimit   int
	LookbackDays     int
}

// TrendingConfig windows discovery lists (posts + creators). Scoring reuses feed /
// reputation weights; lookback is shorter than feed/reputation defaults.
type TrendingConfig struct {
	LookbackDays      int
	CandidateLimit    int
	CreatorEventLimit int
}

type Service struct {
	posts              domain.Repository
	signer             media.Signer
	follows            FollowingStore
	reputation         ReputationNotifier
	feed               FeedConfig
	trending           TrendingConfig
	repCfg             ReputationConfig
	maxImages          int
	maxAVDurationMs    int
	defaultPageSize    int
	maxPageSize        int
	sharePublicBaseURL string
}

func NewService(
	posts domain.Repository,
	signer media.Signer,
	follows FollowingStore,
	reputation ReputationNotifier,
	feed FeedConfig,
	trending TrendingConfig,
	repCfg ReputationConfig,
	maxImages, maxAVDurationSec, defaultPageSize, maxPageSize int,
	sharePublicBaseURL string,
) *Service {
	if maxImages <= 0 {
		maxImages = 10
	}
	if maxAVDurationSec <= 0 {
		maxAVDurationSec = 10
	}
	if defaultPageSize <= 0 {
		defaultPageSize = 20
	}
	if maxPageSize <= 0 {
		maxPageSize = 50
	}
	if feed.Gravity <= 0 {
		feed.Gravity = 1.5
	}
	if feed.CommentWeight <= 0 {
		feed.CommentWeight = 3
	}
	if feed.ShareWeight <= 0 {
		feed.ShareWeight = 4
	}
	if feed.BookmarkWeight <= 0 {
		feed.BookmarkWeight = 3
	}
	if feed.FollowBoost <= 0 {
		feed.FollowBoost = 0.35
	}
	if feed.GeoD0Km <= 0 {
		feed.GeoD0Km = 5
	}
	if feed.NearbyRadiusKm <= 0 {
		feed.NearbyRadiusKm = 25
	}
	if feed.CandidateLimit <= 0 {
		feed.CandidateLimit = 200
	}
	if feed.LookbackDays <= 0 {
		feed.LookbackDays = 14
	}
	if trending.LookbackDays <= 0 {
		trending.LookbackDays = 7
	}
	if trending.CandidateLimit <= 0 {
		trending.CandidateLimit = 500
	}
	if trending.CreatorEventLimit <= 0 {
		trending.CreatorEventLimit = 5000
	}
	repCfg = repCfg.withDefaults()
	return &Service{
		posts: posts, signer: signer, follows: follows, reputation: reputation,
		feed: feed, trending: trending, repCfg: repCfg,
		maxImages: maxImages, maxAVDurationMs: maxAVDurationSec * 1000,
		defaultPageSize: defaultPageSize, maxPageSize: maxPageSize,
		sharePublicBaseURL: strings.TrimRight(sharePublicBaseURL, "/"),
	}
}

type SignMediaInput struct {
	UserID string
	Kind   string
}

type SignMediaResult struct {
	media.SignResult
	ResourceType string `json:"resource_type"`
}

type CreatePostInput struct {
	AuthorID  uuid.UUID
	Title     string
	Caption   string
	Lat       float64
	Lng       float64
	MediaMode domain.MediaMode
	Media     []domain.CreateMediaInput
	Hashtags  []string
}

type ListInput struct {
	Limit  int
	Cursor string
}

type PageResult struct {
	Posts      []domain.Post
	NextCursor string
}

type CommentPageResult struct {
	Comments   []domain.Comment
	NextCursor string
}

type ShareResult struct {
	Post     domain.Post
	ShareURL string
}

type FeedInput struct {
	ViewerID uuid.UUID
	Lat      *float64
	Lng      *float64
	Limit    int
	Cursor   string
}

type RelatedInput struct {
	SeedID uuid.UUID
	Query  string // search query when opened from search results; optional
	Limit  int
	Cursor string
}

type RelatedResult struct {
	Posts      []domain.Post
	NextCursor string
	// Source is search | similar | mixed — not a home feed.
	Source string
}

type TrendingListInput struct {
	Limit  int
	Cursor string
}

type TrendingCreator struct {
	UserID uuid.UUID
	Score  int64
}

type TrendingCreatorsResult struct {
	Creators []TrendingCreator
}

func (s *Service) SignMedia(ctx context.Context, in SignMediaInput) (SignMediaResult, error) {
	res, err := s.signer.SignUpload(ctx, media.SignRequest{
		UserID: in.UserID,
		Kind:   in.Kind,
	})
	if err != nil {
		return SignMediaResult{}, err
	}
	return SignMediaResult{
		SignResult:   res,
		ResourceType: resourceType(in.Kind),
	}, nil
}

func resourceType(kind string) string {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "video", "audio":
		return "video"
	default:
		return "image"
	}
}

func (s *Service) Create(ctx context.Context, in CreatePostInput) (domain.Post, error) {
	mode, mediaItems, tags, err := s.validateCreate(in)
	if err != nil {
		return domain.Post{}, err
	}
	p, err := s.posts.Create(ctx, domain.CreateInput{
		AuthorID:  in.AuthorID,
		Title:     strings.TrimSpace(in.Title),
		Caption:   strings.TrimSpace(in.Caption),
		Lat:       in.Lat,
		Lng:       in.Lng,
		MediaMode: mode,
		Media:     mediaItems,
		Hashtags:  tags,
	})
	if err != nil {
		return domain.Post{}, err
	}
	s.syncAuthorReputation(ctx, in.AuthorID)
	return p, nil
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (domain.Post, error) {
	return s.posts.ByID(ctx, id)
}

func (s *Service) ListByAuthor(ctx context.Context, authorID uuid.UUID, in ListInput) (PageResult, error) {
	p, err := s.listParams(in)
	if err != nil {
		return PageResult{}, err
	}
	page, err := s.posts.ListByAuthor(ctx, authorID, p)
	if err != nil {
		return PageResult{}, err
	}
	return toPageResult(page), nil
}

func (s *Service) SetReaction(ctx context.Context, postID, userID uuid.UUID, reaction *domain.Reaction) (domain.Post, error) {
	if reaction != nil {
		switch *reaction {
		case domain.ReactionLike, domain.ReactionDislike:
		default:
			return domain.Post{}, apperr.Validation(apperr.Field{
				Field: "reaction", Code: "INVALID", Message: "reaction must be like, dislike, or null",
			})
		}
	}
	p, err := s.posts.SetReaction(ctx, postID, userID, reaction)
	if err != nil {
		return domain.Post{}, err
	}
	if p.AuthorID != userID {
		s.syncAuthorReputation(ctx, p.AuthorID)
	}
	return p, nil
}

func (s *Service) AddComment(ctx context.Context, postID, authorID uuid.UUID, body string, parentID *uuid.UUID) (domain.Comment, error) {
	body = strings.TrimSpace(body)
	if body == "" {
		return domain.Comment{}, apperr.Validation(apperr.Required("body"))
	}
	if utf8.RuneCountInString(body) > 2000 {
		return domain.Comment{}, apperr.Validation(apperr.Field{
			Field: "body", Code: "TOO_LONG", Message: "body must be at most 2000 characters",
		})
	}
	c, err := s.posts.AddComment(ctx, postID, authorID, body, parentID)
	if err != nil {
		return domain.Comment{}, err
	}
	post, err := s.posts.ByID(ctx, postID)
	if err == nil && post.AuthorID != authorID {
		s.syncAuthorReputation(ctx, post.AuthorID)
	}
	return c, nil
}

func (s *Service) ListComments(ctx context.Context, postID uuid.UUID, in ListInput) (CommentPageResult, error) {
	p, err := s.listParams(in)
	if err != nil {
		return CommentPageResult{}, err
	}
	page, err := s.posts.ListComments(ctx, postID, p)
	if err != nil {
		return CommentPageResult{}, err
	}
	out := CommentPageResult{Comments: page.Comments}
	if page.NextCursor != nil {
		out.NextCursor = encodeCursor(*page.NextCursor)
	}
	return out, nil
}

func (s *Service) ListReplies(ctx context.Context, parentID uuid.UUID, in ListInput) (CommentPageResult, error) {
	p, err := s.listParams(in)
	if err != nil {
		return CommentPageResult{}, err
	}
	page, err := s.posts.ListReplies(ctx, parentID, p)
	if err != nil {
		return CommentPageResult{}, err
	}
	out := CommentPageResult{Comments: page.Comments}
	if page.NextCursor != nil {
		out.NextCursor = encodeCursor(*page.NextCursor)
	}
	return out, nil
}

func (s *Service) SetCommentReaction(ctx context.Context, commentID, userID uuid.UUID, reaction *domain.Reaction) (domain.Comment, error) {
	if reaction != nil {
		switch *reaction {
		case domain.ReactionLike, domain.ReactionDislike:
		default:
			return domain.Comment{}, apperr.Validation(apperr.Field{
				Field: "reaction", Code: "INVALID", Message: "reaction must be like, dislike, or null",
			})
		}
	}
	return s.posts.SetCommentReaction(ctx, commentID, userID, reaction)
}

func (s *Service) AddBookmark(ctx context.Context, userID, postID uuid.UUID) (domain.Post, error) {
	p, err := s.posts.AddBookmark(ctx, userID, postID)
	if err != nil {
		return domain.Post{}, err
	}
	if p.AuthorID != userID {
		s.syncAuthorReputation(ctx, p.AuthorID)
	}
	return p, nil
}

func (s *Service) RemoveBookmark(ctx context.Context, userID, postID uuid.UUID) (domain.Post, error) {
	p, err := s.posts.RemoveBookmark(ctx, userID, postID)
	if err != nil {
		return domain.Post{}, err
	}
	if p.AuthorID != userID {
		s.syncAuthorReputation(ctx, p.AuthorID)
	}
	return p, nil
}

func (s *Service) ListBookmarks(ctx context.Context, userID uuid.UUID, in ListInput) (PageResult, error) {
	p, err := s.listParams(in)
	if err != nil {
		return PageResult{}, err
	}
	page, err := s.posts.ListBookmarks(ctx, userID, p)
	if err != nil {
		return PageResult{}, err
	}
	return toPageResult(page), nil
}

func (s *Service) Share(ctx context.Context, postID, actorID uuid.UUID) (ShareResult, error) {
	p, err := s.posts.Share(ctx, postID, actorID)
	if err != nil {
		return ShareResult{}, err
	}
	if p.AuthorID != actorID {
		s.syncAuthorReputation(ctx, p.AuthorID)
	}
	return ShareResult{
		Post:     p,
		ShareURL: s.sharePublicBaseURL + "/posts/" + p.ID.String(),
	}, nil
}

func (s *Service) Search(ctx context.Context, query string, in ListInput) (PageResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return PageResult{}, apperr.Validation(apperr.Required("q"))
	}
	p, err := s.listParams(in)
	if err != nil {
		return PageResult{}, err
	}
	page, err := s.posts.Search(ctx, query, p)
	if err != nil {
		return PageResult{}, err
	}
	return toPageResult(page), nil
}

func (s *Service) ListByHashtag(ctx context.Context, tag string, in ListInput) (PageResult, error) {
	tag, err := normalizeHashtag(tag)
	if err != nil {
		return PageResult{}, err
	}
	p, err := s.listParams(in)
	if err != nil {
		return PageResult{}, err
	}
	page, err := s.posts.ListByHashtag(ctx, tag, p)
	if err != nil {
		return PageResult{}, err
	}
	return toPageResult(page), nil
}

func (s *Service) Feed(ctx context.Context, in FeedInput) (PageResult, error) {
	limit := in.Limit
	if limit <= 0 {
		limit = s.defaultPageSize
	}
	if limit > s.maxPageSize {
		limit = s.maxPageSize
	}
	hasLoc := in.Lat != nil && in.Lng != nil
	if (in.Lat == nil) != (in.Lng == nil) {
		return PageResult{}, apperr.Validation(apperr.Field{
			Field: "lat/lng", Code: "INVALID", Message: "lat and lng must both be provided or both omitted",
		})
	}
	if hasLoc && (*in.Lat < -90 || *in.Lat > 90 || *in.Lng < -180 || *in.Lng > 180) {
		return PageResult{}, apperr.Validation(apperr.Field{
			Field: "lat/lng", Code: "INVALID", Message: "lat/lng out of range",
		})
	}

	since := time.Now().UTC().Add(-time.Duration(s.feed.LookbackDays) * 24 * time.Hour)
	followed := map[uuid.UUID]struct{}{}
	var authorIDs []uuid.UUID
	if s.follows != nil {
		ids, err := s.follows.FollowingIDs(ctx, in.ViewerID)
		if err != nil {
			return PageResult{}, err
		}
		authorIDs = ids
		for _, id := range ids {
			followed[id] = struct{}{}
		}
	}

	var candidates []domain.Post
	if len(authorIDs) > 0 {
		fromFollow, err := s.posts.ListByAuthorIDs(ctx, authorIDs, since, s.feed.CandidateLimit)
		if err != nil {
			return PageResult{}, err
		}
		candidates = append(candidates, fromFollow...)
	} else {
		// Newcomers / empty follow graph: global recent pool ranked by score.
		recent, err := s.posts.ListRecent(ctx, since, s.feed.CandidateLimit)
		if err != nil {
			return PageResult{}, err
		}
		candidates = append(candidates, recent...)
	}
	if hasLoc {
		nearby, err := s.posts.ListNearby(ctx, *in.Lat, *in.Lng, s.feed.NearbyRadiusKm, since, s.feed.CandidateLimit)
		if err != nil {
			return PageResult{}, err
		}
		candidates = append(candidates, nearby...)
	}

	seen := map[uuid.UUID]domain.Post{}
	for _, p := range candidates {
		seen[p.ID] = p
	}
	merged := make([]domain.Post, 0, len(seen))
	for _, p := range seen {
		merged = append(merged, p)
	}

	type scored struct {
		post  domain.Post
		score float64
	}
	now := time.Now().UTC()
	items := make([]scored, 0, len(merged))
	for _, p := range merged {
		_, isFollow := followed[p.AuthorID]
		items = append(items, scored{
			post:  p,
			score: s.rankScore(p, isFollow, in.Lat, in.Lng, now),
		})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].score == items[j].score {
			return items[i].post.ID.String() > items[j].post.ID.String()
		}
		return items[i].score > items[j].score
	})

	cursorScore, cursorID, hasCursor, err := decodeFeedCursor(in.Cursor)
	if err != nil {
		return PageResult{}, err
	}
	start := 0
	if hasCursor {
		for i, it := range items {
			if it.score < cursorScore || (it.score == cursorScore && it.post.ID.String() < cursorID.String()) {
				start = i
				break
			}
			if i == len(items)-1 {
				start = len(items)
			}
		}
		// skip items that are >= cursor (already seen)
		for start < len(items) {
			it := items[start]
			if it.score < cursorScore || (it.score == cursorScore && it.post.ID.String() < cursorID.String()) {
				break
			}
			start++
		}
	}
	end := start + limit
	hasMore := end < len(items)
	if end > len(items) {
		end = len(items)
	}
	pageItems := items[start:end]
	posts := make([]domain.Post, 0, len(pageItems))
	for _, it := range pageItems {
		posts = append(posts, it.post)
	}
	out := PageResult{Posts: posts}
	if hasMore && len(pageItems) > 0 {
		last := pageItems[len(pageItems)-1]
		out.NextCursor = encodeFeedCursor(last.score, last.post.ID)
	}
	return out, nil
}

// TrendingPosts ranks recent posts with the feed HN formula (no follow/geo boost).
func (s *Service) TrendingPosts(ctx context.Context, in TrendingListInput) (PageResult, error) {
	limit := in.Limit
	if limit <= 0 {
		limit = s.defaultPageSize
	}
	if limit > s.maxPageSize {
		limit = s.maxPageSize
	}
	since := time.Now().UTC().Add(-time.Duration(s.trending.LookbackDays) * 24 * time.Hour)
	candidates, err := s.posts.ListRecent(ctx, since, s.trending.CandidateLimit)
	if err != nil {
		return PageResult{}, err
	}

	type scored struct {
		post  domain.Post
		score float64
	}
	now := time.Now().UTC()
	items := make([]scored, 0, len(candidates))
	for _, p := range candidates {
		items = append(items, scored{
			post:  p,
			score: s.rankScore(p, false, nil, nil, now),
		})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].score == items[j].score {
			return items[i].post.ID.String() > items[j].post.ID.String()
		}
		return items[i].score > items[j].score
	})

	cursorScore, cursorID, hasCursor, err := decodeFeedCursor(in.Cursor)
	if err != nil {
		return PageResult{}, err
	}
	start := 0
	if hasCursor {
		for start < len(items) {
			it := items[start]
			if it.score < cursorScore || (it.score == cursorScore && it.post.ID.String() < cursorID.String()) {
				break
			}
			start++
		}
	}
	end := start + limit
	hasMore := end < len(items)
	if end > len(items) {
		end = len(items)
	}
	pageItems := items[start:end]
	posts := make([]domain.Post, 0, len(pageItems))
	for _, it := range pageItems {
		posts = append(posts, it.post)
	}
	out := PageResult{Posts: posts}
	if hasMore && len(pageItems) > 0 {
		last := pageItems[len(pageItems)-1]
		out.NextCursor = encodeFeedCursor(last.score, last.post.ID)
	}
	return out, nil
}

// TrendingCreators ranks authors by recent interaction-event score (reputation weights).
func (s *Service) TrendingCreators(ctx context.Context, in TrendingListInput) (TrendingCreatorsResult, error) {
	limit := in.Limit
	if limit <= 0 {
		limit = s.defaultPageSize
	}
	if limit > s.maxPageSize {
		limit = s.maxPageSize
	}
	since := time.Now().UTC().Add(-time.Duration(s.trending.LookbackDays) * 24 * time.Hour)
	events, err := s.posts.ListInteractionsSince(ctx, since, s.trending.CreatorEventLimit)
	if err != nil {
		return TrendingCreatorsResult{}, err
	}
	creators := rankTrendingCreators(events, time.Now().UTC(), s.repCfg, limit)
	return TrendingCreatorsResult{Creators: creators}, nil
}

func (s *Service) rankScore(p domain.Post, followed bool, viewerLat, viewerLng *float64, now time.Time) float64 {
	engagement := float64(p.LikeCount) +
		s.feed.CommentWeight*float64(p.CommentCount) +
		s.feed.ShareWeight*float64(p.ShareCount) +
		s.feed.BookmarkWeight*float64(p.BookmarkCount)
	ageHours := now.Sub(p.CreatedAt.UTC()).Hours()
	if ageHours < 0 {
		ageHours = 0
	}
	base := math.Pow(engagement+1, 0.8) / math.Pow(ageHours+2, s.feed.Gravity)
	followMul := 1.0
	if followed {
		followMul = 1 + s.feed.FollowBoost
	}
	geo := 1.0
	if viewerLat != nil && viewerLng != nil {
		dist := haversineKm(*viewerLat, *viewerLng, p.Lat, p.Lng)
		geo = 1.0 / (1.0 + dist/s.feed.GeoD0Km)
	}
	return base * followMul * geo
}

func haversineKm(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371.0
	toRad := math.Pi / 180
	dLat := (lat2 - lat1) * toRad
	dLng := (lng2 - lng1) * toRad
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*toRad)*math.Cos(lat2*toRad)*math.Sin(dLng/2)*math.Sin(dLng/2)
	return 2 * R * math.Asin(math.Sqrt(a))
}

func (s *Service) syncAuthorReputation(ctx context.Context, authorID uuid.UUID) {
	if s.reputation == nil || !s.reputation.Enabled() {
		return
	}
	since := time.Now().UTC().Add(-time.Duration(s.repCfg.LookbackDays) * 24 * time.Hour)
	events, err := s.posts.ListAuthorInteractions(ctx, authorID, since, s.repCfg.CandidateLimit)
	if err != nil {
		return
	}
	score, postsCreated, likesReceived := authorReputationFromEvents(events, time.Now().UTC(), s.repCfg)
	_ = s.reputation.SyncScore(ctx, authorID, score, postsCreated, likesReceived)
}

// Related returns other search hits for q (excluding the seed), then similar posts
// by caption/hashtag text match. Not the home feed.
func (s *Service) Related(ctx context.Context, in RelatedInput) (RelatedResult, error) {
	limit := in.Limit
	if limit <= 0 {
		limit = s.defaultPageSize
	}
	if limit > s.maxPageSize {
		limit = s.maxPageSize
	}
	if _, err := s.posts.ByID(ctx, in.SeedID); err != nil {
		return RelatedResult{}, err
	}
	q := strings.TrimSpace(in.Query)
	phase, searchCur, simCur, err := decodeRelatedCursor(in.Cursor)
	if err != nil {
		return RelatedResult{}, err
	}

	var posts []domain.Post
	source := "similar"
	next := ""

	if phase == relatedPhaseSearch || (phase == "" && q != "") {
		need := limit
		var listParams domain.ListParams
		listParams.Limit = need
		if searchCur != nil {
			listParams.Cursor = searchCur
		}
		page, err := s.posts.SearchExcluding(ctx, q, in.SeedID, listParams)
		if err != nil {
			return RelatedResult{}, err
		}
		posts = append(posts, page.Posts...)
		if page.NextCursor != nil {
			next = encodeRelatedSearchCursor(*page.NextCursor)
			source = "search"
		} else {
			// Search exhausted — fill remainder with similar on this page if needed,
			// or advance cursor into similar phase when page is full of search hits.
			remain := limit - len(posts)
			if remain > 0 {
				simPosts, simNext, err := s.relatedSimilarPage(ctx, in.SeedID, q, nil, remain)
				if err != nil {
					return RelatedResult{}, err
				}
				posts = append(posts, simPosts...)
				next = simNext
				if len(simPosts) > 0 && len(page.Posts) > 0 {
					source = "mixed"
				} else if len(simPosts) > 0 {
					source = "similar"
				} else {
					source = "search"
				}
			} else {
				// Full page of search; next request should start similar.
				seedText, err := s.posts.SeedText(ctx, in.SeedID)
				if err != nil {
					return RelatedResult{}, err
				}
				probe, err := s.posts.ListSimilar(ctx, seedText, in.SeedID, q, nil, 1)
				if err != nil {
					return RelatedResult{}, err
				}
				if len(probe.Posts) > 0 {
					next = encodeRelatedSimilarStart()
				}
				source = "search"
			}
		}
		return RelatedResult{Posts: posts, NextCursor: next, Source: source}, nil
	}

	// Similar-only phase (no q, or cursor already in similar).
	var cur *domain.SimilarCursor
	if phase == relatedPhaseSimilar {
		cur = simCur
	}
	simPosts, simNext, err := s.relatedSimilarPage(ctx, in.SeedID, q, cur, limit)
	if err != nil {
		return RelatedResult{}, err
	}
	return RelatedResult{Posts: simPosts, NextCursor: simNext, Source: "similar"}, nil
}

func (s *Service) relatedSimilarPage(ctx context.Context, seedID uuid.UUID, excludeQuery string, cursor *domain.SimilarCursor, limit int) ([]domain.Post, string, error) {
	seedText, err := s.posts.SeedText(ctx, seedID)
	if err != nil {
		return nil, "", err
	}
	page, err := s.posts.ListSimilar(ctx, seedText, seedID, excludeQuery, cursor, limit)
	if err != nil {
		return nil, "", err
	}
	posts := make([]domain.Post, 0, len(page.Posts))
	for _, sp := range page.Posts {
		posts = append(posts, sp.Post)
	}
	next := ""
	if page.HasMore && len(page.Posts) > 0 {
		last := page.Posts[len(page.Posts)-1]
		next = encodeRelatedSimilarCursor(last.Rank, last.Post.CreatedAt, last.Post.ID)
	}
	return posts, next, nil
}

const (
	relatedPhaseSearch  = "s"
	relatedPhaseSimilar = "m"
)

func encodeRelatedSearchCursor(c domain.PageCursor) string {
	raw := fmt.Sprintf("%s|%d|%s", relatedPhaseSearch, c.CreatedAt.UTC().UnixNano(), c.ID.String())
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func encodeRelatedSimilarStart() string {
	raw := relatedPhaseSimilar + "|start"
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func encodeRelatedSimilarCursor(rank float64, createdAt time.Time, id uuid.UUID) string {
	raw := fmt.Sprintf("%s|%g|%d|%s", relatedPhaseSimilar, rank, createdAt.UTC().UnixNano(), id.String())
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func decodeRelatedCursor(raw string) (phase string, searchCur *domain.PageCursor, simCur *domain.SimilarCursor, err error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil, nil, nil
	}
	b, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return "", nil, nil, apperr.Validation(apperr.Field{
			Field: "cursor", Code: "INVALID", Message: "cursor is invalid",
		})
	}
	parts := strings.Split(string(b), "|")
	if len(parts) < 2 {
		return "", nil, nil, apperr.Validation(apperr.Field{
			Field: "cursor", Code: "INVALID", Message: "cursor is invalid",
		})
	}
	switch parts[0] {
	case relatedPhaseSearch:
		if len(parts) != 3 {
			return "", nil, nil, apperr.Validation(apperr.Field{
				Field: "cursor", Code: "INVALID", Message: "cursor is invalid",
			})
		}
		ns, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return "", nil, nil, apperr.Validation(apperr.Field{
				Field: "cursor", Code: "INVALID", Message: "cursor is invalid",
			})
		}
		id, err := uuid.Parse(parts[2])
		if err != nil {
			return "", nil, nil, apperr.Validation(apperr.Field{
				Field: "cursor", Code: "INVALID", Message: "cursor is invalid",
			})
		}
		c := domain.PageCursor{CreatedAt: time.Unix(0, ns).UTC(), ID: id}
		return relatedPhaseSearch, &c, nil, nil
	case relatedPhaseSimilar:
		if parts[1] == "start" {
			return relatedPhaseSimilar, nil, nil, nil
		}
		if len(parts) != 4 {
			return "", nil, nil, apperr.Validation(apperr.Field{
				Field: "cursor", Code: "INVALID", Message: "cursor is invalid",
			})
		}
		rank, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			return "", nil, nil, apperr.Validation(apperr.Field{
				Field: "cursor", Code: "INVALID", Message: "cursor is invalid",
			})
		}
		ns, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			return "", nil, nil, apperr.Validation(apperr.Field{
				Field: "cursor", Code: "INVALID", Message: "cursor is invalid",
			})
		}
		id, err := uuid.Parse(parts[3])
		if err != nil {
			return "", nil, nil, apperr.Validation(apperr.Field{
				Field: "cursor", Code: "INVALID", Message: "cursor is invalid",
			})
		}
		c := domain.SimilarCursor{Rank: rank, CreatedAt: time.Unix(0, ns).UTC(), ID: id}
		return relatedPhaseSimilar, nil, &c, nil
	default:
		return "", nil, nil, apperr.Validation(apperr.Field{
			Field: "cursor", Code: "INVALID", Message: "cursor is invalid",
		})
	}
}

func (s *Service) validateCreate(in CreatePostInput) (domain.MediaMode, []domain.CreateMediaInput, []string, error) {
	var details []apperr.Field
	title := strings.TrimSpace(in.Title)
	if title == "" {
		details = append(details, apperr.Required("title"))
	} else if utf8.RuneCountInString(title) > 200 {
		details = append(details, apperr.Field{Field: "title", Code: "TOO_LONG", Message: "title must be at most 200 characters"})
	}
	if in.Lat < -90 || in.Lat > 90 {
		details = append(details, apperr.Field{Field: "lat", Code: "INVALID", Message: "lat must be between -90 and 90"})
	}
	if in.Lng < -180 || in.Lng > 180 {
		details = append(details, apperr.Field{Field: "lng", Code: "INVALID", Message: "lng must be between -180 and 180"})
	}
	if utf8.RuneCountInString(in.Caption) > 5000 {
		details = append(details, apperr.Field{Field: "caption", Code: "TOO_LONG", Message: "caption must be at most 5000 characters"})
	}

	mode := domain.MediaMode(strings.TrimSpace(string(in.MediaMode)))
	switch mode {
	case domain.ModeImages, domain.ModeImagesAudio, domain.ModeVideo:
	default:
		details = append(details, apperr.Field{Field: "media_mode", Code: "INVALID", Message: "media_mode must be images, images_audio, or video"})
	}

	if len(in.Media) == 0 {
		details = append(details, apperr.Required("media"))
	}

	mediaItems := make([]domain.CreateMediaInput, 0, len(in.Media))
	images, videos, audios := 0, 0, 0
	for i, m := range in.Media {
		kind := domain.MediaKind(strings.ToLower(strings.TrimSpace(string(m.Kind))))
		prefix := fmt.Sprintf("media[%d]", i)
		if m.CloudinaryPublicID == "" || m.URL == "" {
			details = append(details, apperr.Field{Field: prefix, Code: "INVALID", Message: "cloudinary_public_id and url are required"})
		}
		switch kind {
		case domain.KindImage:
			images++
		case domain.KindVideo:
			videos++
			if m.DurationMs <= 0 || m.DurationMs > s.maxAVDurationMs {
				details = append(details, apperr.Field{
					Field: prefix + ".duration_ms", Code: "INVALID",
					Message: fmt.Sprintf("video duration_ms must be 1..%d", s.maxAVDurationMs),
				})
			}
		case domain.KindAudio:
			audios++
			if m.DurationMs <= 0 || m.DurationMs > s.maxAVDurationMs {
				details = append(details, apperr.Field{
					Field: prefix + ".duration_ms", Code: "INVALID",
					Message: fmt.Sprintf("audio duration_ms must be 1..%d", s.maxAVDurationMs),
				})
			}
		default:
			details = append(details, apperr.Field{Field: prefix + ".kind", Code: "INVALID", Message: "kind must be image, video, or audio"})
		}
		mediaItems = append(mediaItems, domain.CreateMediaInput{
			Kind:               kind,
			CloudinaryPublicID: strings.TrimSpace(m.CloudinaryPublicID),
			URL:                strings.TrimSpace(m.URL),
			SortOrder:          m.SortOrder,
			DurationMs:         m.DurationMs,
			Width:              m.Width,
			Height:             m.Height,
			Mime:               strings.TrimSpace(m.Mime),
		})
	}

	switch mode {
	case domain.ModeImages:
		if images < 1 || images > s.maxImages || videos != 0 || audios != 0 {
			details = append(details, apperr.Field{
				Field: "media", Code: "INVALID",
				Message: fmt.Sprintf("images mode requires 1..%d images and no video/audio", s.maxImages),
			})
		}
	case domain.ModeImagesAudio:
		if images < 1 || images > s.maxImages || audios != 1 || videos != 0 {
			details = append(details, apperr.Field{
				Field: "media", Code: "INVALID",
				Message: fmt.Sprintf("images_audio mode requires 1..%d images and exactly one audio", s.maxImages),
			})
		}
	case domain.ModeVideo:
		if videos != 1 || images != 0 || audios != 0 {
			details = append(details, apperr.Field{
				Field: "media", Code: "INVALID",
				Message: "video mode requires exactly one video and no images/audio",
			})
		}
	}

	tags, tagErr := normalizeHashtags(in.Hashtags)
	if tagErr != nil {
		details = append(details, tagErr.Details...)
	}

	if len(details) > 0 {
		return "", nil, nil, apperr.Validation(details...)
	}
	return mode, mediaItems, tags, nil
}

func (s *Service) listParams(in ListInput) (domain.ListParams, error) {
	limit := in.Limit
	if limit <= 0 {
		limit = s.defaultPageSize
	}
	if limit > s.maxPageSize {
		limit = s.maxPageSize
	}
	cur, ok, err := decodeCursor(in.Cursor)
	if err != nil {
		return domain.ListParams{}, err
	}
	p := domain.ListParams{Limit: limit}
	if ok {
		p.Cursor = &cur
	}
	return p, nil
}

func toPageResult(page domain.Page) PageResult {
	out := PageResult{Posts: page.Posts}
	if out.Posts == nil {
		out.Posts = []domain.Post{}
	}
	if page.NextCursor != nil {
		out.NextCursor = encodeCursor(*page.NextCursor)
	}
	return out
}

func normalizeHashtags(raw []string) ([]string, *apperr.Error) {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(raw))
	var details []apperr.Field
	for i, t := range raw {
		tag, err := normalizeHashtag(t)
		if err != nil {
			details = append(details, apperr.Field{
				Field: fmt.Sprintf("hashtags[%d]", i), Code: "INVALID", Message: err.Error(),
			})
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		out = append(out, tag)
	}
	if len(out) > 30 {
		details = append(details, apperr.Field{Field: "hashtags", Code: "TOO_MANY", Message: "at most 30 hashtags"})
	}
	if len(details) > 0 {
		return nil, apperr.Validation(details...)
	}
	return out, nil
}

func normalizeHashtag(raw string) (string, error) {
	tag := strings.ToLower(strings.TrimSpace(raw))
	tag = strings.TrimPrefix(tag, "#")
	if tag == "" || !hashtagRe.MatchString(tag) {
		return "", fmt.Errorf("hashtag must be 1-50 chars [a-z0-9_]")
	}
	return tag, nil
}

func encodeCursor(c domain.PageCursor) string {
	raw := fmt.Sprintf("%d|%s", c.CreatedAt.UTC().UnixNano(), c.ID.String())
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func decodeCursor(raw string) (domain.PageCursor, bool, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return domain.PageCursor{}, false, nil
	}
	b, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return domain.PageCursor{}, false, apperr.Validation(apperr.Field{
			Field: "cursor", Code: "INVALID", Message: "cursor is invalid",
		})
	}
	parts := strings.SplitN(string(b), "|", 2)
	if len(parts) != 2 {
		return domain.PageCursor{}, false, apperr.Validation(apperr.Field{
			Field: "cursor", Code: "INVALID", Message: "cursor is invalid",
		})
	}
	ns, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return domain.PageCursor{}, false, apperr.Validation(apperr.Field{
			Field: "cursor", Code: "INVALID", Message: "cursor is invalid",
		})
	}
	id, err := uuid.Parse(parts[1])
	if err != nil {
		return domain.PageCursor{}, false, apperr.Validation(apperr.Field{
			Field: "cursor", Code: "INVALID", Message: "cursor is invalid",
		})
	}
	return domain.PageCursor{CreatedAt: time.Unix(0, ns).UTC(), ID: id}, true, nil
}

func encodeFeedCursor(score float64, id uuid.UUID) string {
	raw := fmt.Sprintf("%g|%s", score, id.String())
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func decodeFeedCursor(raw string) (float64, uuid.UUID, bool, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, uuid.Nil, false, nil
	}
	b, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return 0, uuid.Nil, false, apperr.Validation(apperr.Field{
			Field: "cursor", Code: "INVALID", Message: "cursor is invalid",
		})
	}
	parts := strings.SplitN(string(b), "|", 2)
	if len(parts) != 2 {
		return 0, uuid.Nil, false, apperr.Validation(apperr.Field{
			Field: "cursor", Code: "INVALID", Message: "cursor is invalid",
		})
	}
	score, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, uuid.Nil, false, apperr.Validation(apperr.Field{
			Field: "cursor", Code: "INVALID", Message: "cursor is invalid",
		})
	}
	id, err := uuid.Parse(parts[1])
	if err != nil {
		return 0, uuid.Nil, false, apperr.Validation(apperr.Field{
			Field: "cursor", Code: "INVALID", Message: "cursor is invalid",
		})
	}
	return score, id, true, nil
}
