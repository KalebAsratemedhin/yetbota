package media

import (
	"context"
	"io"
)

// UploadedImage is the result of a successful cloud upload.
type UploadedImage struct {
	URL      string
	PublicID string
}

// ImageStore uploads and deletes images (Cloudinary in production).
type ImageStore interface {
	Configured() bool
	Upload(ctx context.Context, folder, publicID string, r io.Reader, filename string) (UploadedImage, error)
	Destroy(ctx context.Context, publicID string) error
}
