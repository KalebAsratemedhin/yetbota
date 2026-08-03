package screening

import "context"

type Request struct {
	Kind    string
	Text    string
	UserID  string
	PlaceID string
}

type Result struct {
	OK     bool
	Reason string
}

type Client interface {
	Screen(ctx context.Context, req *Request) (*Result, error)
}
