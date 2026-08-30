package post

import (
	"context"
	"encoding/base64"
	"fmt"
	"regexp"
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

type Service struct {
	posts              domain.Repository
	signer             media.Signer
	maxImages          int
	maxAVDurationMs    int
	defaultPageSize    int
	maxPageSize        int
	sharePublicBaseURL string
}

func NewService(
	posts domain.Repository,
	signer media.Signer,
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
	return &Service{
		posts:              posts,
		signer:             signer,
		maxImages:          maxImages,
		maxAVDurationMs:    maxAVDurationSec * 1000,
		defaultPageSize:    defaultPageSize,
		maxPageSize:        maxPageSize,
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
	return s.posts.Create(ctx, domain.CreateInput{
		AuthorID:  in.AuthorID,
		Caption:   strings.TrimSpace(in.Caption),
		Lat:       in.Lat,
		Lng:       in.Lng,
		MediaMode: mode,
		Media:     mediaItems,
		Hashtags:  tags,
	})
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
	return s.posts.SetReaction(ctx, postID, userID, reaction)
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
	return s.posts.AddComment(ctx, postID, authorID, body, parentID)
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
	return s.posts.AddBookmark(ctx, userID, postID)
}

func (s *Service) RemoveBookmark(ctx context.Context, userID, postID uuid.UUID) (domain.Post, error) {
	return s.posts.RemoveBookmark(ctx, userID, postID)
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

func (s *Service) Share(ctx context.Context, postID uuid.UUID) (ShareResult, error) {
	p, err := s.posts.Share(ctx, postID)
	if err != nil {
		return ShareResult{}, err
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

func (s *Service) validateCreate(in CreatePostInput) (domain.MediaMode, []domain.CreateMediaInput, []string, error) {
	var details []apperr.Field
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
