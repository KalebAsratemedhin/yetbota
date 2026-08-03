package moderation

import (
	"context"
	"database/sql"
	"time"
)

const (
	ContentTypePost    = "POST"
	ContentTypeComment = "COMMENT"

	ReasonSpam      = "SPAM"
	ReasonOffensive = "OFFENSIVE"
	ReasonIncorrect = "INCORRECT"
	ReasonOther     = "OTHER"

	CaseStatusPending  = "PENDING"
	CaseStatusResolved = "RESOLVED"
	CaseStatusRejected = "REJECTED"

	ResolutionDeleted   = "DELETED"
	ResolutionDismissed = "DISMISSED"

	ActionDelete  = "DELETE"
	ActionDismiss = "DISMISS"
	ActionBan     = "BAN"
	ActionUnhide  = "UNHIDE"
)

type Report struct {
	ID          string
	ContentType string
	ContentID   string
	ReporterID  string
	Reason      string
	Details     string
	CreatedAt   time.Time
}

type ModerationCase struct {
	ID              string
	ContentType     string
	ContentID       string
	ReportCount     int
	Status          string
	Severity        int
	AutoHidden      bool
	FirstReportedAt time.Time
	LastReportedAt  time.Time
	ResolvedBy      string
	ResolvedAt      *time.Time
	Resolution      string
	Version         int
}

type ModerationAction struct {
	ID         string
	CaseID     string
	AdminID    string
	Action     string
	TargetType string
	TargetID   string
	Note       string
	CreatedAt  time.Time
}

type CaseListOptions struct {
	Status      string
	Reason      string
	ContentType string
	Page        int
	PageSize    int
}

type Repository interface {
	CreateReport(ctx context.Context, tx *sql.Tx, r *Report) (bool, error)
	UpsertCaseOnReport(ctx context.Context, tx *sql.Tx, id, contentType, contentID string, reportedAt time.Time) (*ModerationCase, error)
	MarkCaseAutoHidden(ctx context.Context, tx *sql.Tx, caseID string) error
	ListCases(ctx context.Context, opts *CaseListOptions) ([]*ModerationCase, error)
	CountCases(ctx context.Context, opts *CaseListOptions) (int64, error)
	GetCase(ctx context.Context, id string) (*ModerationCase, error)
	ListReports(ctx context.Context, contentType, contentID string) ([]*Report, error)
	ResolveCase(ctx context.Context, tx *sql.Tx, caseID string, expectedVersion int, status, resolution, resolvedBy string) (bool, error)
	CreateAction(ctx context.Context, tx *sql.Tx, a *ModerationAction) error
}

type RateLimiter interface {
	Allow(ctx context.Context, key string, max int, window time.Duration) (bool, error)
}

type UserBanner interface {
	Ban(ctx context.Context, userID, reason, actorID string) error
}
