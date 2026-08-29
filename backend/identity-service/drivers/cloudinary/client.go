package cloudinary

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/yetbota/identity-service/drivers/config"
	"github.com/yetbota/identity-service/internal/domain/apperr"
	"github.com/yetbota/identity-service/internal/domain/media"
)

var _ media.ImageStore = (*Client)(nil)

type Client struct {
	cloudName string
	apiKey    string
	apiSecret string
	folder    string
	http      *http.Client
}

func New(cfg config.CloudinaryConfig) *Client {
	folder := strings.Trim(cfg.Folder, "/")
	if folder == "" {
		folder = "yetbota/avatars"
	}
	return &Client{
		cloudName: cfg.CloudName,
		apiKey:    cfg.APIKey,
		apiSecret: cfg.APISecret,
		folder:    folder,
		http:      &http.Client{Timeout: 60 * time.Second},
	}
}

func (c *Client) Configured() bool {
	return c != nil && c.cloudName != "" && c.apiKey != "" && c.apiSecret != ""
}

func (c *Client) Upload(ctx context.Context, folder, publicID string, r io.Reader, filename string) (media.UploadedImage, error) {
	if !c.Configured() {
		return media.UploadedImage{}, apperr.New(apperr.CodeInternal, "cloudinary is not configured")
	}
	if folder == "" {
		folder = c.folder
	}
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	params := map[string]string{
		"folder":    folder,
		"public_id": publicID,
		"timestamp": ts,
		"overwrite": "true",
	}
	sig := sign(params, c.apiSecret)

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	_ = w.WriteField("api_key", c.apiKey)
	_ = w.WriteField("timestamp", ts)
	_ = w.WriteField("signature", sig)
	_ = w.WriteField("folder", folder)
	_ = w.WriteField("public_id", publicID)
	_ = w.WriteField("overwrite", "true")
	part, err := w.CreateFormFile("file", filename)
	if err != nil {
		return media.UploadedImage{}, apperr.Wrap(err, apperr.CodeInternal, "cloudinary form")
	}
	if _, err := io.Copy(part, r); err != nil {
		return media.UploadedImage{}, apperr.Wrap(err, apperr.CodeInternal, "cloudinary read file")
	}
	if err := w.Close(); err != nil {
		return media.UploadedImage{}, apperr.Wrap(err, apperr.CodeInternal, "cloudinary form close")
	}

	endpoint := fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/image/upload", url.PathEscape(c.cloudName))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, &body)
	if err != nil {
		return media.UploadedImage{}, apperr.Wrap(err, apperr.CodeInternal, "cloudinary request")
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	res, err := c.http.Do(req)
	if err != nil {
		return media.UploadedImage{}, apperr.Wrap(err, apperr.CodeInternal, "cloudinary upload")
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if res.StatusCode >= 300 {
		return media.UploadedImage{}, apperr.Wrap(
			fmt.Errorf("status %d: %s", res.StatusCode, strings.TrimSpace(string(raw))),
			apperr.CodeInternal,
			"cloudinary upload failed",
		)
	}
	var out struct {
		SecureURL string `json:"secure_url"`
		PublicID  string `json:"public_id"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return media.UploadedImage{}, apperr.Wrap(err, apperr.CodeInternal, "cloudinary decode")
	}
	if out.SecureURL == "" || out.PublicID == "" {
		return media.UploadedImage{}, apperr.New(apperr.CodeInternal, "cloudinary empty response")
	}
	return media.UploadedImage{URL: out.SecureURL, PublicID: out.PublicID}, nil
}

func (c *Client) Destroy(ctx context.Context, publicID string) error {
	if !c.Configured() {
		return apperr.New(apperr.CodeInternal, "cloudinary is not configured")
	}
	publicID = strings.TrimSpace(publicID)
	if publicID == "" {
		return nil
	}
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	params := map[string]string{
		"public_id": publicID,
		"timestamp": ts,
	}
	sig := sign(params, c.apiSecret)

	form := url.Values{}
	form.Set("api_key", c.apiKey)
	form.Set("timestamp", ts)
	form.Set("signature", sig)
	form.Set("public_id", publicID)

	endpoint := fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/image/destroy", url.PathEscape(c.cloudName))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return apperr.Wrap(err, apperr.CodeInternal, "cloudinary destroy request")
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.http.Do(req)
	if err != nil {
		return apperr.Wrap(err, apperr.CodeInternal, "cloudinary destroy")
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if res.StatusCode >= 300 {
		return apperr.Wrap(
			fmt.Errorf("status %d: %s", res.StatusCode, strings.TrimSpace(string(raw))),
			apperr.CodeInternal,
			"cloudinary destroy failed",
		)
	}
	return nil
}

func sign(params map[string]string, secret string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+params[k])
	}
	payload := strings.Join(parts, "&") + secret
	sum := sha1.Sum([]byte(payload))
	return hex.EncodeToString(sum[:])
}
