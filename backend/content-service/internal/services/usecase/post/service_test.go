package post

import (
	"testing"

	"github.com/google/uuid"

	domain "github.com/yetbota/content-service/internal/domain/post"
)

func TestValidateCreateModes(t *testing.T) {
	s := NewService(nil, nil, nil, nil, FeedConfig{}, ReputationConfig{}, 10, 10, 20, 50, "http://localhost:3000")
	author := uuid.New()
	img := domain.CreateMediaInput{Kind: domain.KindImage, CloudinaryPublicID: "a", URL: "https://x/a"}
	vid := domain.CreateMediaInput{Kind: domain.KindVideo, CloudinaryPublicID: "v", URL: "https://x/v", DurationMs: 5000}
	aud := domain.CreateMediaInput{Kind: domain.KindAudio, CloudinaryPublicID: "s", URL: "https://x/s", DurationMs: 5000}

	ok := []CreatePostInput{
		{AuthorID: author, Lat: 9, Lng: 38, MediaMode: domain.ModeImages, Media: []domain.CreateMediaInput{img}},
		{AuthorID: author, Lat: 9, Lng: 38, MediaMode: domain.ModeImagesAudio, Media: []domain.CreateMediaInput{img, aud}},
		{AuthorID: author, Lat: 9, Lng: 38, MediaMode: domain.ModeVideo, Media: []domain.CreateMediaInput{vid}},
	}
	for i, in := range ok {
		if _, _, _, err := s.validateCreate(in); err != nil {
			t.Fatalf("case %d: unexpected error: %v", i, err)
		}
	}

	bad := []CreatePostInput{
		{AuthorID: author, Lat: 9, Lng: 38, MediaMode: domain.ModeImages, Media: nil},
		{AuthorID: author, Lat: 9, Lng: 38, MediaMode: domain.ModeImages, Media: []domain.CreateMediaInput{vid}},
		{AuthorID: author, Lat: 9, Lng: 38, MediaMode: domain.ModeImagesAudio, Media: []domain.CreateMediaInput{img}},
		{AuthorID: author, Lat: 9, Lng: 38, MediaMode: domain.ModeVideo, Media: []domain.CreateMediaInput{img, vid}},
		{AuthorID: author, Lat: 9, Lng: 38, MediaMode: domain.ModeVideo, Media: []domain.CreateMediaInput{
			{Kind: domain.KindVideo, CloudinaryPublicID: "v", URL: "https://x/v", DurationMs: 20000},
		}},
		{AuthorID: author, Lat: 99, Lng: 38, MediaMode: domain.ModeImages, Media: []domain.CreateMediaInput{img}},
	}
	for i, in := range bad {
		if _, _, _, err := s.validateCreate(in); err == nil {
			t.Fatalf("case %d: expected validation error", i)
		}
	}
}
