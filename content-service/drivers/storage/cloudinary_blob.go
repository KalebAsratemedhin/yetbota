package storage

import (
	"bytes"
	"context"
	"fmt"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	blob "github.com/beka-birhanu/yetbota/content-service/internal/domain/storage"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/google/uuid"
)

type cloudinaryBlob struct {
	client *cloudinary.Cloudinary
	folder string
}

type CloudinaryConfig struct {
	CloudName string `validate:"required"`
	APIKey    string `validate:"required"`
	APISecret string `validate:"required"`
	Folder    string
}

func (c *CloudinaryConfig) Validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

func NewCloudinaryBlob(c *CloudinaryConfig) (blob.Bucket, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	client, err := cloudinary.NewFromParams(c.CloudName, c.APIKey, c.APISecret)
	if err != nil {
		return nil, &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "Something went wrong",
			ServiceMessage:    fmt.Sprintf("failed to init cloudinary client: %s", err),
		}
	}
	client.Config.URL.Secure = true

	return &cloudinaryBlob{client: client, folder: c.Folder}, nil
}

func (c *cloudinaryBlob) UploadFile(ctx context.Context, in *blob.UploadRequest) (*blob.UploadResponse, error) {
	publicID := uuid.NewString()
	res, err := c.client.Upload.Upload(ctx, bytes.NewReader(in.FileInByte), uploader.UploadParams{
		PublicID:     publicID,
		Folder:       c.folder,
		ResourceType: "image",
	})
	if err != nil {
		return nil, &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "Something went wrong",
			ServiceMessage:    fmt.Sprintf("failed to upload file: %s", err),
		}
	}
	if res.Error.Message != "" {
		return nil, &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "Something went wrong",
			ServiceMessage:    fmt.Sprintf("cloudinary upload error: %s", res.Error.Message),
		}
	}

	return &blob.UploadResponse{
		FileName:    res.PublicID,
		Url:         res.SecureURL,
		ContentType: in.ContentType,
	}, nil
}

func (c *cloudinaryBlob) SignURL(ctx context.Context, in *blob.SignURLRequest) (*blob.SignURLResponse, error) {
	img, err := c.client.Image(in.FileName)
	if err != nil {
		return nil, &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "Something went wrong",
			ServiceMessage:    fmt.Sprintf("failed to build cloudinary url: %s", err),
		}
	}
	url, err := img.String()
	if err != nil {
		return nil, &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "Something went wrong",
			ServiceMessage:    fmt.Sprintf("failed to serialize cloudinary url: %s", err),
		}
	}
	return &blob.SignURLResponse{Url: url}, nil
}

func (c *cloudinaryBlob) RemoveFile(ctx context.Context, in *blob.DeleteRequest) error {
	res, err := c.client.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID:     in.FileName,
		ResourceType: "image",
	})
	if err != nil {
		return &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "Something went wrong",
			ServiceMessage:    fmt.Sprintf("failed to delete file %q: %s", in.FileName, err),
		}
	}
	if res.Error.Message != "" {
		return &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "Something went wrong",
			ServiceMessage:    fmt.Sprintf("cloudinary destroy error: %s", res.Error.Message),
		}
	}
	return nil
}
