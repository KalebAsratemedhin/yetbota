package identity

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/yetbota/content-service/internal/domain/apperr"
	"github.com/yetbota/content-service/internal/domain/auth"
)

var _ auth.Introspector = (*Client)(nil)
var _ interface {
	AuthenticateToken(ctx context.Context, token string) (auth.Principal, error)
} = (*Client)(nil)

type Client struct {
	baseURL string
	http    *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) AuthenticateToken(ctx context.Context, token string) (auth.Principal, error) {
	p, ok, err := c.Introspect(ctx, token)
	if err != nil {
		return auth.Principal{}, err
	}
	if !ok {
		return auth.Principal{}, apperr.ErrUnauthorized
	}
	return p, nil
}

func (c *Client) Introspect(ctx context.Context, accessToken string) (auth.Principal, bool, error) {
	accessToken = strings.TrimSpace(accessToken)
	if accessToken == "" {
		return auth.Principal{}, false, nil
	}
	body, _ := json.Marshal(map[string]string{"token": accessToken})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/auth/introspect", bytes.NewReader(body))
	if err != nil {
		return auth.Principal{}, false, apperr.Wrap(err, apperr.CodeInternal, "introspect request")
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := c.http.Do(req)
	if err != nil {
		return auth.Principal{}, false, apperr.Wrap(err, apperr.CodeInternal, "introspect call")
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if res.StatusCode >= 300 {
		return auth.Principal{}, false, apperr.Wrap(
			fmt.Errorf("status %d: %s", res.StatusCode, strings.TrimSpace(string(raw))),
			apperr.CodeInternal,
			"introspect failed",
		)
	}
	var out struct {
		Active   bool   `json:"active"`
		UserID   string `json:"user_id"`
		Username string `json:"username"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return auth.Principal{}, false, apperr.Wrap(err, apperr.CodeInternal, "introspect decode")
	}
	if !out.Active || out.UserID == "" {
		return auth.Principal{}, false, nil
	}
	return auth.Principal{UserID: out.UserID, Username: out.Username}, true, nil
}
