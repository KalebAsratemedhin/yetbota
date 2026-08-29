package social

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/yetbota/identity-service/internal/domain/apperr"
	domainfollow "github.com/yetbota/identity-service/internal/domain/follow"
	domainuser "github.com/yetbota/identity-service/internal/domain/user"
)

const (
	searchMinQueryLen = 2
	searchMaxResults  = 20
	listDefaultLimit  = 20
	listMaxLimit      = 50
)

type Service struct {
	users   domainuser.Repository
	follows domainfollow.Store
}

func NewService(users domainuser.Repository, follows domainfollow.Store) *Service {
	return &Service{users: users, follows: follows}
}

type PublicProfile struct {
	UserID    string
	Username  string
	FullName  string
	Phone     string
	Address   string
	Bio       string
	AvatarURL string
}

type ProfilePage struct {
	Users      []PublicProfile
	NextCursor string
	Total      int64
}

type ListInput struct {
	Limit  int
	Cursor string
}

func toPublic(u domainuser.User) PublicProfile {
	return PublicProfile{
		UserID:    u.ID.String(),
		Username:  u.Username,
		FullName:  u.FullName,
		Phone:     u.Phone,
		Address:   u.Address,
		Bio:       u.Bio,
		AvatarURL: u.AvatarURL,
	}
}

func (s *Service) GetPublicProfile(ctx context.Context, userID uuid.UUID) (PublicProfile, error) {
	u, err := s.users.ByID(ctx, userID)
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			return PublicProfile{}, apperr.ErrNotFound
		}
		return PublicProfile{}, err
	}
	return toPublic(u), nil
}

func (s *Service) Search(ctx context.Context, viewerID uuid.UUID, query string) ([]PublicProfile, error) {
	q := strings.TrimSpace(query)
	if utf8.RuneCountInString(q) < searchMinQueryLen {
		return nil, apperr.Validation(apperr.Field{
			Field: "q", Code: "TOO_SHORT", Message: "query must be at least 2 characters",
		})
	}
	rows, err := s.users.Search(ctx, q, viewerID, searchMaxResults)
	if err != nil {
		return nil, err
	}
	out := make([]PublicProfile, 0, len(rows))
	for _, u := range rows {
		out = append(out, toPublic(u))
	}
	return out, nil
}

func (s *Service) Follow(ctx context.Context, followerID, followeeID uuid.UUID) error {
	if followerID == followeeID {
		return apperr.Validation(apperr.Field{
			Field: "id", Code: "INVALID", Message: "cannot follow yourself",
		})
	}
	if _, err := s.users.ByID(ctx, followeeID); err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			return apperr.ErrNotFound
		}
		return err
	}
	return s.follows.Follow(ctx, followerID, followeeID)
}

func (s *Service) Unfollow(ctx context.Context, followerID, followeeID uuid.UUID) error {
	if followerID == followeeID {
		return apperr.Validation(apperr.Field{
			Field: "id", Code: "INVALID", Message: "cannot unfollow yourself",
		})
	}
	return s.follows.Unfollow(ctx, followerID, followeeID)
}

func (s *Service) Followers(ctx context.Context, userID uuid.UUID, in ListInput) (ProfilePage, error) {
	return s.listHydrated(ctx, userID, in, true)
}

func (s *Service) Following(ctx context.Context, userID uuid.UUID, in ListInput) (ProfilePage, error) {
	return s.listHydrated(ctx, userID, in, false)
}

func (s *Service) listHydrated(ctx context.Context, userID uuid.UUID, in ListInput, followers bool) (ProfilePage, error) {
	if _, err := s.users.ByID(ctx, userID); err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			return ProfilePage{}, apperr.ErrNotFound
		}
		return ProfilePage{}, err
	}

	params := domainfollow.ListParams{
		Limit:  normalizeListLimit(in.Limit),
		Cursor: strings.TrimSpace(in.Cursor),
	}

	var (
		page  domainfollow.EdgePage
		total int64
		err   error
	)
	if followers {
		page, err = s.follows.Followers(ctx, userID, params)
		if err != nil {
			return ProfilePage{}, err
		}
		total, err = s.follows.CountFollowers(ctx, userID)
	} else {
		page, err = s.follows.Following(ctx, userID, params)
		if err != nil {
			return ProfilePage{}, err
		}
		total, err = s.follows.CountFollowing(ctx, userID)
	}
	if err != nil {
		return ProfilePage{}, err
	}

	users, err := s.hydrateOrdered(ctx, page.UserIDs)
	if err != nil {
		return ProfilePage{}, err
	}
	return ProfilePage{
		Users:      users,
		NextCursor: page.NextCursor,
		Total:      total,
	}, nil
}

// hydrateOrdered batch-loads users and returns them in the same order as ids.
// Missing users (deleted accounts with stale edges) are skipped.
func (s *Service) hydrateOrdered(ctx context.Context, ids []uuid.UUID) ([]PublicProfile, error) {
	if len(ids) == 0 {
		return []PublicProfile{}, nil
	}
	rows, err := s.users.ByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	byID := make(map[uuid.UUID]domainuser.User, len(rows))
	for _, u := range rows {
		byID[u.ID] = u
	}
	out := make([]PublicProfile, 0, len(ids))
	for _, id := range ids {
		u, ok := byID[id]
		if !ok {
			continue
		}
		out = append(out, toPublic(u))
	}
	return out, nil
}

func normalizeListLimit(n int) int {
	if n <= 0 {
		return listDefaultLimit
	}
	if n > listMaxLimit {
		return listMaxLimit
	}
	return n
}
