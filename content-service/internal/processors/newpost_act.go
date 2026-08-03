package processors

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/beka-birhanu/yetbota/content-service/drivers/constants"
	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
	"github.com/beka-birhanu/yetbota/content-service/drivers/utils"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	domainNotification "github.com/beka-birhanu/yetbota/content-service/internal/domain/notification"
	domainPhoto "github.com/beka-birhanu/yetbota/content-service/internal/domain/photo"
	domainPost "github.com/beka-birhanu/yetbota/content-service/internal/domain/post"
	domainPostphoto "github.com/beka-birhanu/yetbota/content-service/internal/domain/postphoto"
	domainStorage "github.com/beka-birhanu/yetbota/content-service/internal/domain/storage"
	domainUser "github.com/beka-birhanu/yetbota/content-service/internal/domain/user"
	"github.com/google/uuid"
)

type newPostActivity struct {
	postPhotoRepo      domainPostphoto.Repository
	photoRepo          domainPhoto.Repository
	bucket             domainStorage.Bucket
	postRepo           domainPost.Repository
	notificationRepo   domainNotification.Repository
	notificationSender domainNotification.Sender
	userRepo           domainUser.Repository
}

type newPostActConfig struct {
	PostPhotoRepo      domainPostphoto.Repository `validate:"required"`
	PhotoRepo          domainPhoto.Repository     `validate:"required"`
	Bucket             domainStorage.Bucket       `validate:"required"`
	PostRepo           domainPost.Repository      `validate:"required"`
	NotificationRepo   domainNotification.Repository
	NotificationSender domainNotification.Sender
	UserRepo           domainUser.Repository
}

func (c *newPostActConfig) validate() error {
	return validator.Validate.Struct(c)
}

func newNewPostActivity(cfg *newPostActConfig) (*newPostActivity, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &newPostActivity{
		postPhotoRepo:      cfg.PostPhotoRepo,
		photoRepo:          cfg.PhotoRepo,
		bucket:             cfg.Bucket,
		postRepo:           cfg.PostRepo,
		notificationRepo:   cfg.NotificationRepo,
		notificationSender: cfg.NotificationSender,
		userRepo:           cfg.UserRepo,
	}, nil
}

// MarkPostDuplicate hides a post flagged as a duplicate by the ai-service and
// notifies its author. Best-effort: notification failures do not fail the activity.
func (a *newPostActivity) MarkPostDuplicate(ctx context.Context, postID string) error {
	if err := a.postRepo.UpdateModerationStatus(ctx, nil, postID, constants.ModerationStatusDuplicate); err != nil {
		return err
	}
	post, err := a.postRepo.Read(ctx, postID)
	if err != nil {
		log.Printf("newpost: failed to read post %s for duplicate notification: %v", postID, err)
		return nil
	}
	a.notifyDuplicate(ctx, post.UserID)
	return nil
}

func (a *newPostActivity) notifyDuplicate(ctx context.Context, userID string) {
	if userID == "" {
		return
	}
	const title = "Post not published"
	const body = "Your post looks like a duplicate of an existing post, so it wasn't published."

	if a.notificationRepo != nil {
		n := &dbmodels.Notification{
			ID:     uuid.New().String(),
			UserID: userID,
			Title:  title,
			Body:   body,
			SentAt: time.Now(),
		}
		if err := a.notificationRepo.Add(ctx, nil, n); err != nil {
			log.Printf("newpost: failed to persist duplicate notification for user %s: %v", userID, err)
		}
	}

	if a.notificationSender == nil || a.userRepo == nil {
		return
	}
	tokens, err := a.userRepo.GetDeviceTokens(ctx, userID)
	if err != nil {
		log.Printf("newpost: failed to fetch device tokens for user %s: %v", userID, err)
		return
	}
	if len(tokens) == 0 {
		return
	}
	if err := a.notificationSender.Send(ctx, &domainNotification.NotificationData{
		Title:     title,
		Body:      body,
		Receivers: tokens,
	}); err != nil {
		log.Printf("newpost: failed to send duplicate notification to user %s: %v", userID, err)
	}
}

// FetchPostPhotoIDs returns IDs of all photos attached to a post, ordered by position.
func (a *newPostActivity) FetchPostPhotoIDs(ctx context.Context, postID string) ([]string, error) {
	postPhotos, err := a.postPhotoRepo.List(ctx,
		&domainPostphoto.Options{PostIDs: []string{postID}},
		&domainPostphoto.SortOptions{
			Field:     domainPostphoto.SortFieldPosition,
			Direction: domainPostphoto.SortDirectionAsc,
		},
	)
	if err != nil {
		return nil, err
	}
	ids := make([]string, len(postPhotos))
	for i, pp := range postPhotos {
		ids[i] = pp.PhotoID
	}
	return ids, nil
}

// ProcessPhoto downloads the original, generates mobile/web compressed variants, uploads them,
// and updates the photo DB record. Idempotent: already-processed photos are skipped.
func (a *newPostActivity) ProcessPhoto(ctx context.Context, photoID string) error {
	photo, err := a.photoRepo.Read(ctx, photoID)
	if err != nil {
		return err
	}

	if photo.URLMobile.Valid && photo.URLWeb.Valid {
		return nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, photo.URL, nil)
	if err != nil {
		return fmt.Errorf("build request for photo %s: %w", photoID, err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("download photo %s: %w", photoID, err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Println("failed to close photo download response body", "photoID", photoID, "error", err)
		}
	}()

	original, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read photo body %s: %w", photoID, err)
	}

	img, _, err := image.Decode(bytes.NewReader(original))
	if err != nil {
		return fmt.Errorf("decode photo %s: %w", photoID, err)
	}

	w := uint(img.Bounds().Dx())
	h := uint(img.Bounds().Dy())
	updated := false

	mobileURL := photo.URLMobile.String
	webURL := photo.URLWeb.String

	if !photo.URLMobile.Valid && (w > constants.MobileImageResolution || h > constants.MobileImageResolution) {
		mobileURL, err = a.compressAndUpload(ctx, original, constants.MobileImageResolution)
		if err != nil {
			return fmt.Errorf("upload mobile variant for photo %s: %w", photoID, err)
		}
		updated = true
	}

	if !photo.URLWeb.Valid && (w > constants.WebImageResolution || h > constants.WebImageResolution) {
		webURL, err = a.compressAndUpload(ctx, original, constants.WebImageResolution)
		if err != nil {
			return fmt.Errorf("upload web variant for photo %s: %w", photoID, err)
		}
		updated = true
	}

	if updated {
		photo.URLMobile = null.StringFrom(mobileURL)
		photo.URLWeb = null.StringFrom(webURL)
		photo.UpdatedAt = time.Now()
		return a.photoRepo.Update(ctx, nil, photo)
	}
	return nil
}

func (a *newPostActivity) compressAndUpload(ctx context.Context, original []byte, maxDim uint) (string, error) {
	compressed, mime, err := utils.CompressToMaxDim(original, maxDim)
	if err != nil {
		return "", err
	}

	uploadResp, err := a.bucket.UploadFile(ctx, &domainStorage.UploadRequest{
		FileInByte:  compressed,
		ContentType: mime,
	})
	if err != nil {
		return "", err
	}

	return uploadResp.Url, nil
}
