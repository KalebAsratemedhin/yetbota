package qa

import (
	"testing"

	"github.com/google/uuid"

	"github.com/yetbota/content-service/internal/domain/apperr"
)

func TestCreateQuestion_RequiresTitleAndBody(t *testing.T) {
	s := NewService(nil, 20, 50)
	_, err := s.CreateQuestion(t.Context(), CreateQuestionInput{
		AuthorID: uuid.New(), Title: "", Body: "x",
	})
	ae, ok := apperr.As(err)
	if !ok || ae.Code != apperr.CodeValidation {
		t.Fatalf("expected validation, got %v", err)
	}
	_, err = s.CreateQuestion(t.Context(), CreateQuestionInput{
		AuthorID: uuid.New(), Title: "Q", Body: "",
	})
	if _, ok := apperr.As(err); !ok {
		t.Fatalf("expected validation for empty body")
	}
}

func TestCreateQuestion_LatLngPair(t *testing.T) {
	s := NewService(nil, 20, 50)
	lat := 9.0
	_, err := s.CreateQuestion(t.Context(), CreateQuestionInput{
		AuthorID: uuid.New(), Title: "Q", Body: "B", Lat: &lat,
	})
	if _, ok := apperr.As(err); !ok {
		t.Fatalf("expected validation when only lat set")
	}
}
