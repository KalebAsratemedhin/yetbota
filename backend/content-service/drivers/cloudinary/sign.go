package cloudinary

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/yetbota/content-service/drivers/config"
	"github.com/yetbota/content-service/internal/domain/apperr"
	"github.com/yetbota/content-service/internal/domain/media"
)

var _ media.Signer = (*Signer)(nil)

type Signer struct {
	cloudName string
	apiKey    string
	apiSecret string
	folder    string
}

func NewSigner(cfg config.CloudinaryConfig) *Signer {
	folder := strings.Trim(cfg.Folder, "/")
	if folder == "" {
		folder = "yetbota/posts"
	}
	return &Signer{
		cloudName: cfg.CloudName,
		apiKey:    cfg.APIKey,
		apiSecret: cfg.APISecret,
		folder:    folder,
	}
}

func (s *Signer) Configured() bool {
	return s != nil && s.cloudName != "" && s.apiKey != "" && s.apiSecret != ""
}

func (s *Signer) SignUpload(_ context.Context, in media.SignRequest) (media.SignResult, error) {
	if !s.Configured() {
		return media.SignResult{}, apperr.New(apperr.CodeInternal, "cloudinary is not configured")
	}
	kind := strings.ToLower(strings.TrimSpace(in.Kind))
	switch kind {
	case "image", "video", "audio":
	default:
		return media.SignResult{}, apperr.Validation(apperr.Field{
			Field: "kind", Code: "INVALID", Message: "kind must be image, video, or audio",
		})
	}
	ts := time.Now().Unix()
	publicID := strings.TrimSpace(in.PublicID)
	if publicID == "" {
		publicID = uuid.NewString()
	}
	folder := s.folder
	if uid := strings.TrimSpace(in.UserID); uid != "" {
		folder = s.folder + "/" + uid
	}
	params := map[string]string{
		"folder":    folder,
		"public_id": publicID,
		"timestamp": strconv.FormatInt(ts, 10),
	}
	return media.SignResult{
		CloudName: s.cloudName,
		APIKey:    s.apiKey,
		Timestamp: ts,
		Signature: sign(params, s.apiSecret),
		Folder:    folder,
		PublicID:  publicID,
	}, nil
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
	sum := sha1.Sum([]byte(strings.Join(parts, "&") + secret))
	return hex.EncodeToString(sum[:])
}

// ResourceType maps kind → Cloudinary upload resource_type for clients.
func ResourceType(kind string) string {
	switch strings.ToLower(kind) {
	case "video", "audio":
		return "video"
	default:
		return "image"
	}
}
