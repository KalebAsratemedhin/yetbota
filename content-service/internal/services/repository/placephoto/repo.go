package placephoto

import (
	"context"
	"database/sql"
	"fmt"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	domainPlacePhoto "github.com/beka-birhanu/yetbota/content-service/internal/domain/placephoto"
)

type repository struct {
	db *sql.DB
}

type dbExec interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func (r *repository) executor(tx *sql.Tx) dbExec {
	if tx != nil {
		return tx
	}
	return r.db
}

type Config struct {
	DB *sql.DB `validate:"required"`
}

func (c *Config) Validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

func NewRepo(c *Config) (domainPlacePhoto.Repository, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return &repository{db: c.DB}, nil
}

func (r *repository) Add(ctx context.Context, tx *sql.Tx, entity *domainPlacePhoto.PlacePhoto) error {
	return r.AddBulk(ctx, tx, []*domainPlacePhoto.PlacePhoto{entity})
}

func (r *repository) AddBulk(ctx context.Context, tx *sql.Tx, entities []*domainPlacePhoto.PlacePhoto) error {
	exec := r.executor(tx)
	for _, entity := range entities {
		if entity == nil {
			continue
		}
		if _, err := exec.ExecContext(ctx, `
			INSERT INTO place_photos (place_id, photo_id, position)
			VALUES ($1, $2, $3)
		`, entity.PlaceID, entity.PhotoID, entity.Position); err != nil {
			return toddlerr.FromDBError(err, "place_photos")
		}
	}
	return nil
}

func (r *repository) Update(ctx context.Context, tx *sql.Tx, entity *domainPlacePhoto.PlacePhoto) error {
	res, err := r.executor(tx).ExecContext(ctx, `
		UPDATE place_photos
		SET photo_id = $1
		WHERE place_id = $2 AND position = $3
	`, entity.PhotoID, entity.PlaceID, entity.Position)
	if err != nil {
		return toddlerr.FromDBError(err, "place_photos")
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if ra == 0 {
		return &toddlerr.Error{
			PublicStatusCode:  status.NotFound,
			ServiceStatusCode: status.NotFound,
			PublicMessage:     "Place photo not found",
			ServiceMessage:    fmt.Sprintf("place photo not found place_id: %s position: %d", entity.PlaceID, entity.Position),
		}
	}
	return nil
}

func (r *repository) Delete(ctx context.Context, tx *sql.Tx, placeID string, position int) error {
	_, err := r.executor(tx).ExecContext(ctx, `DELETE FROM place_photos WHERE place_id = $1 AND position = $2`, placeID, position)
	if err != nil {
		return toddlerr.FromDBError(err, "place_photos")
	}
	return nil
}

func (r *repository) List(ctx context.Context, opts *domainPlacePhoto.Options, sort *domainPlacePhoto.SortOptions) ([]*domainPlacePhoto.PlacePhoto, error) {
	if opts == nil || len(opts.PlaceIDs) == 0 {
		return nil, nil
	}

	args := []any{opts.PlaceIDs}
	where := "pp.place_id = ANY($1)"
	if opts.PositionEq != nil {
		args = append(args, *opts.PositionEq)
		where = fmt.Sprintf("%s AND pp.position = $%d", where, len(args))
	}

	order := "pp.position ASC"
	if sort != nil && sort.Field != "" && sort.Direction != "" {
		order = fmt.Sprintf("pp.%s %s", string(sort.Field), string(sort.Direction))
	}

	selectCols := "pp.place_id, pp.photo_id, pp.position"
	join := ""
	if opts.LoadPhoto {
		selectCols += ", p.id, p.url, p.url_mobile, p.url_web, p.created_at, p.updated_at"
		join = "JOIN photos p ON p.id = pp.photo_id"
	}

	query := fmt.Sprintf(`
		SELECT %s
		FROM place_photos pp
		%s
		WHERE %s
		ORDER BY %s
	`, selectCols, join, where, order)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, toddlerr.FromDBError(err, "place_photos")
	}
	defer rows.Close()

	out := make([]*domainPlacePhoto.PlacePhoto, 0)
	for rows.Next() {
		pp := &domainPlacePhoto.PlacePhoto{}
		if opts.LoadPhoto {
			pp.PhotoData = &domainPlacePhoto.Photo{}
			if err := rows.Scan(
				&pp.PlaceID, &pp.PhotoID, &pp.Position,
				&pp.PhotoData.ID, &pp.PhotoData.URL, &pp.PhotoData.URLMobile, &pp.PhotoData.URLWeb, &pp.PhotoData.CreatedAt, &pp.PhotoData.UpdatedAt,
			); err != nil {
				return nil, toddlerr.FromDBError(err, "place_photos")
			}
		} else {
			if err := rows.Scan(&pp.PlaceID, &pp.PhotoID, &pp.Position); err != nil {
				return nil, toddlerr.FromDBError(err, "place_photos")
			}
		}
		out = append(out, pp)
	}
	if err := rows.Err(); err != nil {
		return nil, toddlerr.FromDBError(err, "place_photos")
	}
	return out, nil
}
