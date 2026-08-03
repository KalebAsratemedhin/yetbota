package screening

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
)

type Client interface {
	Check(ctx context.Context, text, kind string) (*Result, error)
}

type Result struct {
	OK         bool               `json:"ok"`
	Reason     string             `json:"reason"`
	Categories map[string]float64 `json:"categories"`
}

type client struct {
	baseURL    string
	httpClient *http.Client
}

type Config struct {
	BaseURL   string `validate:"required"`
	TimeoutMS int    `validate:"required"`
}

func NewClient(cfg *Config) (Client, error) {
	if err := validator.Validate.Struct(cfg); err != nil {
		return nil, err
	}
	timeout := time.Duration(cfg.TimeoutMS) * time.Millisecond
	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	return &client{
		baseURL: cfg.BaseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

type checkRequest struct {
	Text string `json:"text"`
	Kind string `json:"kind"`
}

func (c *client) Check(ctx context.Context, text, kind string) (*Result, error) {
	body, err := json.Marshal(checkRequest{Text: text, Kind: kind})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/screening/check", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, screeningUnavailable(err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusServiceUnavailable || resp.StatusCode >= 500 {
		return nil, screeningUnavailable(fmt.Errorf("status %d: %s", resp.StatusCode, string(raw)))
	}

	var result Result
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("decode screening response: %w", err)
	}
	return &result, nil
}

func screeningUnavailable(err error) error {
	return &toddlerr.Error{
		PublicStatusCode:  status.ServerError,
		ServiceStatusCode: status.ServerError,
		PublicMessage:     "Content screening is temporarily unavailable. Please try again.",
		ServiceMessage:    fmt.Sprintf("screening client: %v", err),
	}
}
