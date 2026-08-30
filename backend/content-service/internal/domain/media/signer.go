package media

import "context"

type SignRequest struct {
	UserID   string
	Kind     string // image | video | audio
	Folder   string
	PublicID string
}

type SignResult struct {
	CloudName string `json:"cloud_name"`
	APIKey    string `json:"api_key"`
	Timestamp int64  `json:"timestamp"`
	Signature string `json:"signature"`
	Folder    string `json:"folder"`
	PublicID  string `json:"public_id"`
}

type Signer interface {
	Configured() bool
	SignUpload(ctx context.Context, in SignRequest) (SignResult, error)
}
