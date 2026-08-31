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

	"github.com/google/uuid"

	"github.com/yetbota/content-service/internal/domain/apperr"
)

type ReputationClient struct {
	baseURL string
	token   string
	http    *http.Client
}

func NewReputationClient(baseURL, serviceToken string) *ReputationClient {
	return &ReputationClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   serviceToken,
		http:    &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *ReputationClient) Enabled() bool {
	return c != nil && c.baseURL != "" && c.token != ""
}

func (c *ReputationClient) Record(ctx context.Context, userID uuid.UUID, actorID *uuid.UUID, eventType, subjectType string, subjectID *uuid.UUID) error {
	if !c.Enabled() {
		return nil
	}
	body := map[string]any{
		"user_id":      userID.String(),
		"event_type":   eventType,
		"subject_type": subjectType,
	}
	if actorID != nil {
		body["actor_id"] = actorID.String()
	}
	if subjectID != nil {
		body["subject_id"] = subjectID.String()
	}
	raw, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/internal/reputation/events", bytes.NewReader(raw))
	if err != nil {
		return apperr.Wrap(err, apperr.CodeInternal, "reputation request")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Token", c.token)
	res, err := c.http.Do(req)
	if err != nil {
		return apperr.Wrap(err, apperr.CodeInternal, "reputation call")
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if res.StatusCode >= 300 {
		return apperr.Wrap(fmt.Errorf("status %d: %s", res.StatusCode, strings.TrimSpace(string(b))), apperr.CodeInternal, "reputation failed")
	}
	return nil
}
