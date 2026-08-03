package storage

import "context"

type UploadRequest struct {
	BucketName  string
	FileInByte  []byte
	ContentType string
}

type UploadResponse struct {
	FileName    string
	Url         string
	ContentType string
}

type SignURLRequest struct {
	BucketName string
	FileName   string
}

type SignURLResponse struct {
	Url string
}

type DeleteRequest struct {
	BucketName string
	FileName   string
}

type Bucket interface {
	UploadFile(ctx context.Context, in *UploadRequest) (*UploadResponse, error)
	SignURL(ctx context.Context, in *SignURLRequest) (*SignURLResponse, error)
	RemoveFile(ctx context.Context, in *DeleteRequest) error
}
