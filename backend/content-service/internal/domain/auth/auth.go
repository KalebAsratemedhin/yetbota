package auth

import "context"

type Principal struct {
	UserID   string
	Username string
}

type Introspector interface {
	Introspect(ctx context.Context, accessToken string) (Principal, bool, error)
}
