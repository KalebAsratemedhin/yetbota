package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/beka-birhanu/yetbota/content-service/drivers/utils"
	ctxRP "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
)

func computeChange(current, previous int64) (*int, string) {
	if previous == 0 {
		if current == 0 {
			pct := 0
			return &pct, "flat"
		}
		pct := 100
		return &pct, "up"
	}
	pct := int((float64(current-previous) / float64(previous)) * 100)
	switch {
	case pct > 0:
		return &pct, "up"
	case pct < 0:
		return &pct, "down"
	default:
		return &pct, "flat"
	}
}

func (s *svc) OverviewStats(ctx context.Context, ctxSess *ctxRP.Context) (*OverviewStatsResponse, error) {
	if err := utils.AllowAdminOrCSAAccess(ctxSess); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	totalUsers, err := s.repo.CountUsers(ctx)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	totalQuestions, err := s.repo.CountQuestions(ctx)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	totalLocations, err := s.repo.CountPostsWithLocation(ctx)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	userPeriod, err := s.repo.UserPeriodCount(ctx, 7*24*time.Hour)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	userPct, userDir := computeChange(userPeriod.Current, userPeriod.Previous)
	return &OverviewStatsResponse{
		TotalUsers:     &StatCard{Value: totalUsers, ChangePct: userPct, Direction: userDir},
		TotalQuestions: &StatCard{Value: totalQuestions, Direction: "flat"},
		TotalLocations: &StatCard{Value: totalLocations, Direction: "flat"},
	}, nil
}

func rangeDays(r string) (int, string, bool) {
	switch r {
	case "", "7d":
		return 7, "week", true
	case "30d":
		return 30, "month", true
	case "90d":
		return 90, "quarter", true
	default:
		return 0, "", false
	}
}

func growthLabel(day time.Time, days int) string {
	if days <= 7 {
		return day.Format("Mon")
	}
	return day.Format("Jan 2")
}

func humanCount(n int64) string {
	switch {
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	case n >= 1_000:
		return fmt.Sprintf("%.1fk", float64(n)/1_000)
	default:
		return fmt.Sprintf("%d", n)
	}
}

func (s *svc) OverviewGrowth(ctx context.Context, ctxSess *ctxRP.Context, req *OverviewGrowthRequest) (*OverviewGrowthResponse, error) {
	if err := utils.AllowAdminOrCSAAccess(ctxSess); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	if req.Metric != "users" {
		err := badRequest("unsupported metric; only 'users' is available")
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	days, period, ok := rangeDays(req.Range)
	if !ok {
		err := badRequest("invalid range; use 7d, 30d or 90d")
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	now := time.Now().UTC()
	start := now.Truncate(24*time.Hour).AddDate(0, 0, -(days - 1))

	baseline, err := s.repo.CountUsersBefore(ctx, start)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	daily, err := s.repo.UserGrowthDaily(ctx, start)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	byDay := make(map[string]int64, len(daily))
	for _, b := range daily {
		byDay[b.Day.UTC().Format("2006-01-02")] = b.Count
	}

	points := make([]GrowthPoint, 0, days)
	cumulative := baseline
	var windowNew int64
	for i := 0; i < days; i++ {
		day := start.AddDate(0, 0, i)
		add := byDay[day.Format("2006-01-02")]
		cumulative += add
		windowNew += add
		points = append(points, GrowthPoint{Label: growthLabel(day, days), Value: cumulative})
	}

	return &OverviewGrowthResponse{
		Metric:     "users",
		Range:      req.Range,
		Total:      cumulative,
		DeltaLabel: fmt.Sprintf("+%s this %s", humanCount(windowNew), period),
		Points:     points,
	}, nil
}
