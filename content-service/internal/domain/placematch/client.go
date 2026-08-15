package placematch

import (
	"context"
)

type Request struct {
	Text             string
	ExcludeSourceID  string
	Latitude         *float64
	Longitude        *float64
	RadiusKm         *float64
}

type Result struct {
	PlaceID string
	Score   float64
}

type Client interface {
	Match(ctx context.Context, req *Request) (*Result, error)
}
