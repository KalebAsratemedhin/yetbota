package utils

import (
	"golang.org/x/crypto/bcrypt"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/identity-service/internal/domain/auth"
)

type hasher struct{}

func NewHasher() auth.Hasher {
	return &hasher{}
}

// Hash implements [auth.Hasher].
func (h *hasher) Hash(text string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword(
		[]byte(text),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return "", &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "Something went wrong",
			ServiceMessage:    "BCrypt hashing failed: " + err.Error(),
		}
	}
	return string(hashed), nil
}

// Verify implements [auth.Hasher].
func (h *hasher) Verify(hashedText string, text string) error {
	return bcrypt.CompareHashAndPassword(
		[]byte(hashedText),
		[]byte(text),
	)
}
