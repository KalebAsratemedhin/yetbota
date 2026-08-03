package admin

import (
	domainAdmin "github.com/beka-birhanu/yetbota/content-service/internal/domain/admin"
)

type StatCard struct {
	Value     int64
	ChangePct *int
	Direction string
}

type OverviewStatsResponse struct {
	TotalUsers     *StatCard
	TotalQuestions *StatCard
	TotalLocations *StatCard
}

type OverviewGrowthRequest struct {
	Metric string
	Range  string
}

type GrowthPoint struct {
	Label string
	Value int64
}

type OverviewGrowthResponse struct {
	Metric     string
	Range      string
	Total      int64
	DeltaLabel string
	Points     []GrowthPoint
}

type UserStatsResponse struct {
	NewlyJoinedToday  int64
	NewlyJoinedChange *int
	HighReputation    int64
	Threshold         int64
}

type SystemAuditRequest struct {
	ActionType string
	Actor      string
	From       string
	To         string
	Page       int
	PageSize   int
}

type SystemAuditResponse struct {
	Entries  []*domainAdmin.AuditEntry
	Total    int64
	Page     int
	PageSize int
}
