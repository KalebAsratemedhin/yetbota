package place

import (
	"context"
	"encoding/json"
	"time"

	"github.com/aarondl/null/v8"
	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	ctxRP "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
)

func (s *svc) UpdateSummary(ctx context.Context, ctxSess *ctxRP.Context, req *UpdateSummaryRequest) (*UpdateSummaryResponse, error) {
	if err := req.Validate(); err != nil {
		if ctxSess != nil {
			ctxSess.SetErrorMessage(err.Error())
		}
		return nil, err
	}
	place, err := s.placeRepo.Read(ctx, req.PlaceID)
	if err != nil {
		if ctxSess != nil {
			ctxSess.SetErrorMessage(err.Error())
		}
		return nil, err
	}
	citationsJSON, err := json.Marshal(req.Citations)
	if err != nil {
		err = &toddlerr.Error{
			PublicStatusCode:  status.BadRequest,
			ServiceStatusCode: status.BadRequest,
			PublicMessage:     "invalid citations",
			ServiceMessage:    err.Error(),
		}
		if ctxSess != nil {
			ctxSess.SetErrorMessage(err.Error())
		}
		return nil, err
	}
	now := time.Now().UTC()
	if err := s.placeRepo.UpdateSummary(ctx, nil, place.ID, req.SummaryText, now, citationsJSON); err != nil {
		if ctxSess != nil {
			ctxSess.SetErrorMessage(err.Error())
		}
		return nil, err
	}
	place.SummaryText = null.StringFrom(req.SummaryText)
	place.SummaryGeneratedAt = null.TimeFrom(now)
	place.SummaryCitations = citationsJSON
	return &UpdateSummaryResponse{Place: place}, nil
}
