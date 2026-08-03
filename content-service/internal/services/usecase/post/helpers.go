package post

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/aarondl/null/v8"
	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/content-service/drivers/constants"
	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
	"github.com/beka-birhanu/yetbota/content-service/drivers/geotypes"
	"github.com/beka-birhanu/yetbota/content-service/drivers/utils"
	domainPostphoto "github.com/beka-birhanu/yetbota/content-service/internal/domain/postphoto"
	"github.com/beka-birhanu/yetbota/content-service/internal/domain/storage"
	"github.com/google/uuid"
	"github.com/twpayne/go-geom"
	"golang.org/x/sync/errgroup"
)

func cloudinaryPublicIDFromURL(keyOrURL string) string {
	if keyOrURL == "" {
		return ""
	}
	if !strings.HasPrefix(keyOrURL, "http://") && !strings.HasPrefix(keyOrURL, "https://") {
		return keyOrURL
	}
	u, err := url.Parse(keyOrURL)
	if err != nil {
		return keyOrURL
	}
	path := strings.TrimPrefix(u.Path, "/")
	parts := strings.Split(path, "/")
	uploadIdx := -1
	for i, p := range parts {
		if p == "upload" {
			uploadIdx = i
			break
		}
	}
	if uploadIdx < 0 || uploadIdx+1 >= len(parts) {
		return path
	}
	rest := parts[uploadIdx+1:]
	if len(rest) > 0 && len(rest[0]) > 1 && rest[0][0] == 'v' {
		isVersion := true
		for _, c := range rest[0][1:] {
			if c < '0' || c > '9' {
				isVersion = false
				break
			}
		}
		if isVersion {
			rest = rest[1:]
		}
	}
	if len(rest) == 0 {
		return path
	}
	joined := strings.Join(rest, "/")
	if dot := strings.LastIndex(joined, "."); dot > 0 {
		joined = joined[:dot]
	}
	return joined
}

func postFromAddReq(req *AddRequest) *dbmodels.Post {
	var location geotypes.NullPoint
	if req.Latitude != 0 || req.Longitude != 0 {
		location = geotypes.NullPoint{Point: geom.NewPoint(geom.XY).MustSetCoords([]float64{req.Longitude, req.Latitude}), Valid: true}
	}

	return &dbmodels.Post{
		ID:             uuid.NewString(),
		Title:          req.Title,
		Description:    req.Description,
		Tags:           req.Tags,
		IsQuestion:     req.IsQuestion,
		Location:       location,
		Address:        null.NewString(req.Address, req.Address != ""),
		AttachedPostID: null.NewString(req.AttachedPostID, req.AttachedPostID != ""),
	}
}

// postPhotos fetches and assembles the ordered photos for a single post.
func (s *svc) postPhotos(ctx context.Context, postID string, res PhotoResolution) ([]*OrderedPhoto, error) {
	photos, err := s.postPhotoRepo.List(ctx, &domainPostphoto.Options{
		PostIDs:   []string{postID},
		LoadPhoto: true,
	}, &domainPostphoto.SortOptions{
		Field:     domainPostphoto.SortFieldPosition,
		Direction: domainPostphoto.SortDirectionAsc,
	})
	if err != nil {
		return nil, err
	}
	return s.assembleOrderedPhoto(ctx, photos, res)
}

// validateAttachedPost guards the optional attached_post_id reference. An empty
// attachedID is a no-op (no attachment / clearing). A non-empty value is only
// allowed on a question post, may not reference selfID, and must point at an
// existing post.
func (s *svc) validateAttachedPost(ctx context.Context, attachedID, selfID string, isQuestion bool) error {
	if attachedID == "" {
		return nil
	}
	if !isQuestion {
		return &toddlerr.Error{
			PublicStatusCode:  status.BadRequest,
			ServiceStatusCode: status.BadRequest,
			PublicMessage:     "attached_post_id is only allowed on question posts",
			ServiceMessage:    "attached_post_id set on a non-question post",
		}
	}
	if attachedID == selfID {
		return &toddlerr.Error{
			PublicStatusCode:  status.BadRequest,
			ServiceStatusCode: status.BadRequest,
			PublicMessage:     "a post cannot reference itself",
			ServiceMessage:    "attached_post_id equals the post id",
		}
	}
	if _, err := s.postRepo.Read(ctx, attachedID); err != nil {
		var te *toddlerr.Error
		if errors.As(err, &te) && te.PublicStatusCode == status.NotFound {
			return &toddlerr.Error{
				PublicStatusCode:  status.BadRequest,
				ServiceStatusCode: status.BadRequest,
				PublicMessage:     "attached_post_id references a post that does not exist",
				ServiceMessage:    fmt.Sprintf("attached post not found id: %s", attachedID),
			}
		}
		return err
	}
	return nil
}

func (s *svc) uploadPhotos(ctx context.Context, postID string, photos []*OrderedPhotoUpload) (*uploadPhotosResponse, error) {
	gctx, _ := errgroup.WithContext(ctx)

	res := &uploadPhotosResponse{
		photos:     make(dbmodels.PhotoSlice, len(photos)),
		postPhotos: make(dbmodels.PostPhotoSlice, len(photos)),
	}

	slices.SortFunc(photos, func(a, b *OrderedPhotoUpload) int {
		return a.Position - b.Position
	})

	for i, photo := range photos {
		gctx.Go(func() error {
			if int64(len(photo.Photo)) > constants.MaxUploadSize {
				err := &toddlerr.Error{
					PublicStatusCode:  status.BadRequest,
					ServiceStatusCode: status.BadRequest,
					PublicMessage:     fmt.Sprintf("Image exceeds maximum size of %dMB", constants.MaxUploadSize/constants.MB),
					ServiceMessage:    "image too large",
				}
				return err
			}

			mime, err := utils.ImageMimeType(photo.Photo)
			if err != nil {
				return err
			}

			uploadResp, err := s.bucket.UploadFile(ctx, &storage.UploadRequest{
				FileInByte:  photo.Photo,
				ContentType: mime,
			})
			if err != nil {
				return err
			}

			id := uuid.NewString()
			res.photos[i] = &dbmodels.Photo{
				ID:             id,
				BucketProvider: dbmodels.PhotoBucketCLOUDINARY,
				MimeType:       uploadResp.ContentType,
				URL:            uploadResp.Url,
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			}

			res.postPhotos[i] = &dbmodels.PostPhoto{
				PhotoID:  id,
				PostID:   postID,
				Position: photo.Position,
			}

			return nil
		})
	}

	if err := gctx.Wait(); err != nil {
		return nil, err
	}

	return res, nil
}

func (s *svc) deleteUploads(ctx context.Context, photos dbmodels.PhotoSlice) error {
	gctx, _ := errgroup.WithContext(ctx)

	for _, photo := range photos {
		key := cloudinaryPublicIDFromURL(photo.URL)
		gctx.Go(func() error {
			return s.bucket.RemoveFile(ctx, &storage.DeleteRequest{
				FileName: key,
			})
		})
	}

	if err := gctx.Wait(); err != nil {
		return err
	}

	return nil
}

func (s *svc) assembleOrderedPhoto(ctx context.Context, postPhotos dbmodels.PostPhotoSlice, resolution PhotoResolution) ([]*OrderedPhoto, error) {
	orderedPhotos := make([]*OrderedPhoto, len(postPhotos))
	for i, postPhoto := range postPhotos {
		var photo *dbmodels.Photo
		var err error
		if postPhoto.R != nil && postPhoto.R.Photo != nil {
			photo = postPhoto.R.Photo
		} else {
			photo, err = s.photoRepo.Read(ctx, postPhoto.PhotoID)
			if err != nil {
				return nil, err
			}
		}

		orderedPhotos[i] = &OrderedPhoto{
			ID:       photo.ID,
			PostID:   postPhoto.PostID,
			URL:      pickPhotoURL(photo, resolution),
			Position: postPhoto.Position,
		}
	}

	return orderedPhotos, nil
}

func pickPhotoURL(photo *dbmodels.Photo, res PhotoResolution) string {
	if photo == nil {
		return ""
	}
	switch res {
	case PhotoResolutionMobile:
		if photo.URLMobile.Valid && photo.URLMobile.String != "" {
			return photo.URLMobile.String
		}
		fallthrough
	case PhotoResolutionWeb:
		if photo.URLWeb.Valid && photo.URLWeb.String != "" {
			return photo.URLWeb.String
		}
		fallthrough
	default:
		return photo.URL
	}
}
