package validate

import (
	"strings"
	"unicode/utf8"

	"github.com/yetbota/identity-service/internal/domain/apperr"
)

const (
	PhoneMinLen = 7
	PhoneMaxLen = 15
)

// PhoneLength allows empty (cleared) phone, otherwise requires length 7–15.
func PhoneLength(phone string) error {
	p := strings.TrimSpace(phone)
	if p == "" {
		return nil
	}
	n := utf8.RuneCountInString(p)
	if n < PhoneMinLen || n > PhoneMaxLen {
		return apperr.Validation(apperr.Field{
			Field:   "phone",
			Code:    "INVALID_LENGTH",
			Message: "phone must be between 7 and 15 characters",
		})
	}
	return nil
}
