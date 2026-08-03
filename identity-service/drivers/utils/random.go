package utils

import (
	"math/rand"
	"time"

	"github.com/oklog/ulid"
)

func GenerateThreadID() string {
	t := time.Now()
	entropy := rand.New(rand.NewSource(t.UnixNano()))
	uniqueID := ulid.MustNew(ulid.Timestamp(t), entropy)
	return uniqueID.String()
}
