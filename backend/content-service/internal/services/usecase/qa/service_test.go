package qa

import (
	"testing"

	"github.com/google/uuid"

	"github.com/yetbota/content-service/internal/domain/apperr"
)

func TestCreateQuestion_RequiresBody(t *testing.T) {
	s := NewService(nil, 20, 50)
	_, err := s.CreateQuestion(t.Context(), CreateQuestionInput{
		AuthorID: uuid.New(), Body: "",
	})
	if _, ok := apperr.As(err); !ok {
		t.Fatalf("expected validation for empty body, got %v", err)
	}
}

func TestCreateQuestion_LatLngPair(t *testing.T) {
	s := NewService(nil, 20, 50)
	lat := 9.0
	_, err := s.CreateQuestion(t.Context(), CreateQuestionInput{
		AuthorID: uuid.New(), Body: "Is this open late?", Lat: &lat,
	})
	if _, ok := apperr.As(err); !ok {
		t.Fatalf("expected validation when only lat set")
	}
}
