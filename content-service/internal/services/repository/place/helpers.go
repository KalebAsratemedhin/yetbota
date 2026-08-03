package place

import (
	"fmt"
	"strings"

	domainPlace "github.com/beka-birhanu/yetbota/content-service/internal/domain/place"
)

func listPaging(opts *domainPlace.ListOptions) (page int, pageSize int) {
	page = 1
	pageSize = 20
	if opts == nil {
		return page, pageSize
	}
	if opts.Page > 0 {
		page = opts.Page
	}
	if opts.PageSize > 0 && opts.PageSize <= 100 {
		pageSize = opts.PageSize
	}
	return page, pageSize
}

func buildWhereClause(opts *domainPlace.ListOptions, startArg int) (string, []any, int) {
	if opts == nil {
		return "", nil, startArg
	}

	clauses := make([]string, 0, 5)
	args := make([]any, 0, 6)
	next := startArg

	if opts.Search != "" {
		clauses = append(clauses, fmt.Sprintf("(name ILIKE $%d OR description ILIKE $%d OR address ILIKE $%d)", next, next, next))
		args = append(args, "%"+strings.TrimSpace(opts.Search)+"%")
		next++
	}
	if opts.CreatedBy != "" {
		clauses = append(clauses, fmt.Sprintf("created_by = $%d", next))
		args = append(args, opts.CreatedBy)
		next++
	}
	if opts.Status != nil && *opts.Status != "" {
		clauses = append(clauses, fmt.Sprintf("status = $%d", next))
		args = append(args, *opts.Status)
		next++
	}
	if opts.NearLat != nil && opts.NearLon != nil && opts.RadiusKm != nil {
		clauses = append(clauses, fmt.Sprintf(
			"location IS NOT NULL AND ST_DWithin(location::geography, ST_SetSRID(ST_MakePoint($%d, $%d), 4326)::geography, $%d)",
			next, next+1, next+2,
		))
		args = append(args, *opts.NearLon, *opts.NearLat, *opts.RadiusKm*1000)
		next += 3
	}
	if opts.NeedsCheck != nil {
		clauses = append(clauses, fmt.Sprintf("needs_check = $%d", next))
		args = append(args, *opts.NeedsCheck)
		next++
	}

	if len(clauses) == 0 {
		return "", args, next
	}
	return "WHERE " + strings.Join(clauses, " AND "), args, next
}
