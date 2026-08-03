package follow

import (
	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	domainFollow "github.com/beka-birhanu/yetbota/identity-service/internal/domain/follow"
)

func paginationParams(p *domainFollow.Pagination) (skip, limit int) {
	if p == nil || p.Limit <= 0 || p.Page <= 0 {
		return 0, 100
	}
	return (p.Page - 1) * p.Limit, p.Limit
}

func countFromRows(rows []map[string]any) int64 {
	if len(rows) == 0 {
		return 0
	}
	switch n := rows[0]["cnt"].(type) {
	case int64:
		return n
	case int:
		return int64(n)
	case float64:
		return int64(n)
	default:
		return 0
	}
}

func fromGraphError(err error) error {
	return &toddlerr.Error{
		PublicStatusCode:  status.ServerError,
		ServiceStatusCode: status.ServerErrorDatabase,
		PublicMessage:     "database error",
		ServiceMessage:    err.Error(),
	}
}
