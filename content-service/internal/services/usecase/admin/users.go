package admin

import (
	"context"
	"time"

	"github.com/beka-birhanu/yetbota/content-service/drivers/utils"
	ctxRP "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
)

func (s *svc) UserStats(ctx context.Context, ctxSess *ctxRP.Context) (*UserStatsResponse, error) {
	if err := utils.AllowAdminOrCSAAccess(ctxSess); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	todayStart := time.Now().UTC().Truncate(24 * time.Hour)
	yesterdayStart := todayStart.AddDate(0, 0, -1)

	today, err := s.repo.CountUsersSince(ctx, todayStart)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	beforeToday, err := s.repo.CountUsersBefore(ctx, todayStart)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	beforeYesterday, err := s.repo.CountUsersBefore(ctx, yesterdayStart)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	highRep, err := s.repo.HighReputationCount(ctx, s.reputationThreshold)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	pct, _ := computeChange(today, beforeToday-beforeYesterday)
	return &UserStatsResponse{
		NewlyJoinedToday:  today,
		NewlyJoinedChange: pct,
		HighReputation:    highRep,
		Threshold:         s.reputationThreshold,
	}, nil
}
