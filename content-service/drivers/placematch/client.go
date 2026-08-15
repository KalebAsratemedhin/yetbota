package placematch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	domainPlaceMatch "github.com/beka-birhanu/yetbota/content-service/internal/domain/placematch"
)

type client struct {
	baseURL    string
	httpClient *http.Client
}

type Config struct {
	BaseURL   string `validate:"required"`
	TimeoutMS int    `validate:"required"`
}

func NewClient(cfg *Config) (domainPlaceMatch.Client, error) {
	if err := validator.Validate.Struct(cfg); err != nil {
		return nil, err
	}
	timeout := time.Duration(cfg.TimeoutMS) * time.Millisecond
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return &client{
		baseURL: cfg.BaseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

type matchRequest struct {
	Text            string   `json:"text"`
	ExcludeSourceID string   `json:"exclude_source_id,omitempty"`
	Latitude        *float64 `json:"latitude,omitempty"`
	Longitude       *float64 `json:"longitude,omitempty"`
	RadiusKm        *float64 `json:"radius_km,omitempty"`
}

type matchResponse struct {
	PlaceID *string `json:"place_id"`
	Score   float64 `json:"score"`
}

func (c *client) Match(ctx context.Context, req *domainPlaceMatch.Request) (*domainPlaceMatch.Result, error) {
	if req == nil || req.Text == "" {
		return &domainPlaceMatch.Result{}, nil
	}
	body, err := json.Marshal(matchRequest{
		Text:            req.Text,
		ExcludeSourceID: req.ExcludeSourceID,
		Latitude:        req.Latitude,
		Longitude:       req.Longitude,
		RadiusKm:        req.RadiusKm,
	})
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/places/match", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, matchUnavailable(err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusServiceUnavailable || resp.StatusCode >= 500 {
		return nil, matchUnavailable(fmt.Errorf("status %d: %s", resp.StatusCode, string(raw)))
	}
	if resp.StatusCode >= 400 {
		return nil, matchUnavailable(fmt.Errorf("status %d: %s", resp.StatusCode, string(raw)))
	}

	var out matchResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("decode place match response: %w", err)
	}
	result := &domainPlaceMatch.Result{Score: out.Score}
	if out.PlaceID != nil && *out.PlaceID != "" {
		result.PlaceID = *out.PlaceID
	}
	return result, nil
}

func matchUnavailable(err error) error {
	return &toddlerr.Error{
		PublicStatusCode:  status.ServerError,
		ServiceStatusCode: status.ServerError,
		PublicMessage:     "Place matching is temporarily unavailable.",
		ServiceMessage:    fmt.Sprintf("placematch client: %v", err),
	}
}
