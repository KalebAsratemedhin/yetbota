package place

import (
	"context"
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
	ctxRP "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	domainContribution "github.com/beka-birhanu/yetbota/content-service/internal/domain/contribution"
	domainContributionVote "github.com/beka-birhanu/yetbota/content-service/internal/domain/contributionvote"
	domainPlace "github.com/beka-birhanu/yetbota/content-service/internal/domain/place"
	domainPlacePhoto "github.com/beka-birhanu/yetbota/content-service/internal/domain/placephoto"
	domainScreening "github.com/beka-birhanu/yetbota/content-service/internal/domain/screening"
	domainStorage "github.com/beka-birhanu/yetbota/content-service/internal/domain/storage"
	"github.com/google/uuid"
	"github.com/twpayne/go-geom"
	"golang.org/x/sync/errgroup"
)

func placeFromAddReq(req *AddRequest, userID string) *domainPlace.Place {
	var location geotypes.NullPoint
	if req.Latitude != 0 || req.Longitude != 0 {
		location = geotypes.NullPoint{
			Point: geom.NewPoint(geom.XY).MustSetCoords([]float64{req.Longitude, req.Latitude}),
			Valid: true,
		}
	}
	now := time.Now()
	return &domainPlace.Place{
		ID:              uuid.NewString(),
		Name:            strings.TrimSpace(req.Name),
		Description:     strings.TrimSpace(req.Description),
		Address:         null.NewString(strings.TrimSpace(req.Address), strings.TrimSpace(req.Address) != ""),
		Location:        location,
		CreatedBy:       userID,
		Status:          constants.PlaceStatusPending,
		ConfidenceScore: 0,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

func (s *svc) uploadPhotos(ctx context.Context, placeID string, photos []*OrderedPhotoUpload) (*uploadPhotosResponse, error) {
	gctx, _ := errgroup.WithContext(ctx)
	res := &uploadPhotosResponse{
		photos:      make(dbmodels.PhotoSlice, len(photos)),
		placePhotos: make([]*domainPlacePhoto.PlacePhoto, len(photos)),
	}

	slices.SortFunc(photos, func(a, b *OrderedPhotoUpload) int { return a.Position - b.Position })
	for i, p := range photos {
		i := i
		p := p
		gctx.Go(func() error {
			if int64(len(p.Photo)) > constants.MaxUploadSize {
				return &toddlerr.Error{
					PublicStatusCode:  status.BadRequest,
					ServiceStatusCode: status.BadRequest,
					PublicMessage:     fmt.Sprintf("Image exceeds maximum size of %dMB", constants.MaxUploadSize/constants.MB),
					ServiceMessage:    "image too large",
				}
			}

			mime, err := utils.ImageMimeType(p.Photo)
			if err != nil {
				return err
			}
			uploadResp, err := s.bucket.UploadFile(ctx, &domainStorage.UploadRequest{
				FileInByte:  p.Photo,
				ContentType: mime,
			})
			if err != nil {
				return err
			}
			id := uuid.NewString()
			now := time.Now()
			res.photos[i] = &dbmodels.Photo{
				ID:             id,
				BucketProvider: dbmodels.PhotoBucketCLOUDINARY,
				MimeType:       uploadResp.ContentType,
				URL:            uploadResp.Url,
				CreatedAt:      now,
				UpdatedAt:      now,
			}
			res.placePhotos[i] = &domainPlacePhoto.PlacePhoto{
				PlaceID:  placeID,
				PhotoID:  id,
				Position: p.Position,
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
		photo := photo
		gctx.Go(func() error {
			return s.bucket.RemoveFile(ctx, &domainStorage.DeleteRequest{
				FileName: cloudinaryPublicIDFromURL(photo.URL),
			})
		})
	}
	return gctx.Wait()
}

func (s *svc) assembleOrderedPhoto(ctx context.Context, placePhotos []*domainPlacePhoto.PlacePhoto, resolution PhotoResolution) ([]*OrderedPhoto, error) {
	ordered := make([]*OrderedPhoto, len(placePhotos))
	for i, pp := range placePhotos {
		var photo *domainPlacePhoto.Photo
		if pp.PhotoData != nil {
			photo = pp.PhotoData
		} else {
			p, err := s.photoRepo.Read(ctx, pp.PhotoID)
			if err != nil {
				return nil, err
			}
			photo = &domainPlacePhoto.Photo{
				ID:        p.ID,
				URL:       p.URL,
				URLMobile: p.URLMobile,
				URLWeb:    p.URLWeb,
				CreatedAt: p.CreatedAt,
				UpdatedAt: p.UpdatedAt,
			}
		}
		ordered[i] = &OrderedPhoto{
			ID:       photo.ID,
			PlaceID:  pp.PlaceID,
			URL:      pickPhotoURL(photo, resolution),
			Position: pp.Position,
		}
	}
	return ordered, nil
}

func pickPhotoURL(photo *domainPlacePhoto.Photo, res PhotoResolution) string {
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

func lastFreshnessTime(place *domainPlace.Place) *time.Time {
	if place == nil {
		return nil
	}
	var latest *time.Time
	consider := func(t null.Time) {
		if !t.Valid {
			return
		}
		tt := t.Time
		if latest == nil || tt.After(*latest) {
			latest = &tt
		}
	}
	consider(place.LastConfirmedAt)
	consider(place.LastActivityAt)
	if latest == nil {
		t := place.CreatedAt
		latest = &t
	}
	return latest
}

func buildQuestionThreads(grouped map[string][]*ContributionView) []*QuestionThread {
	questions := grouped["question"]
	answers := grouped["answer"]
	byParent := make(map[string][]*ContributionView)
	for _, a := range answers {
		if a == nil || a.Contribution == nil || !a.Contribution.ParentID.Valid {
			continue
		}
		pid := a.Contribution.ParentID.String
		byParent[pid] = append(byParent[pid], a)
	}
	out := make([]*QuestionThread, 0, len(questions))
	for _, q := range questions {
		if q == nil || q.Contribution == nil {
			continue
		}
		thread := &QuestionThread{
			Question: q,
			Answers:  byParent[q.Contribution.ID],
		}
		for _, a := range thread.Answers {
			if a != nil && a.Contribution != nil && a.Contribution.AcceptedAt.Valid {
				thread.AcceptedAnswer = a
				thread.IsResolved = true
				break
			}
		}
		out = append(out, thread)
	}
	return out
}

func computeConfidence(
	contributions []*domainContribution.Contribution,
	recentConfirmCount int64,
	lastFreshness *time.Time,
	decayDays int,
	verificationBoost float64,
) int16 {
	score := 20.0
	for _, c := range contributions {
		if c == nil {
			continue
		}
		weight := 1.0
		switch c.Kind {
		case "important":
			weight = 2.0
		case "change":
			weight = 1.5
		case "answer":
			weight = 0.75
		}
		net := float64(c.Upvotes - c.Downvotes)
		score += weight + net*0.5
	}
	if recentConfirmCount > 0 {
		boost := float64(recentConfirmCount) * (verificationBoost / 5.0)
		if boost > verificationBoost {
			boost = verificationBoost
		}
		score += boost
	}
	if lastFreshness != nil && decayDays > 0 {
		ageDays := time.Since(*lastFreshness).Hours() / 24
		if ageDays > float64(decayDays) {
			score -= (ageDays - float64(decayDays)) * 0.5
		}
	}
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	return int16(score)
}

func screenOrReject(ctx context.Context, client domainScreening.Client, req *domainScreening.Request) error {
	if client == nil {
		return nil
	}
	resp, err := client.Screen(ctx, req)
	if err != nil {
		return err
	}
	if resp == nil || resp.OK {
		return nil
	}
	return &toddlerr.Error{
		PublicStatusCode:  status.BadRequest,
		ServiceStatusCode: status.BadRequest,
		PublicMessage:     resp.Reason,
		ServiceMessage:    fmt.Sprintf("screening rejected content: %s", resp.Reason),
	}
}

func isAdmin(ctxSess *ctxRP.Context) bool {
	if ctxSess == nil {
		return false
	}
	return ctxSess.UserSession.Role == constants.RoleAdmin
}

func votesByContribution(votes map[string]*domainContributionVote.Vote, contributionID string) *string {
	if votes == nil {
		return nil
	}
	vote, ok := votes[contributionID]
	if !ok || vote == nil {
		return nil
	}
	vt := vote.VoteType
	return &vt
}
