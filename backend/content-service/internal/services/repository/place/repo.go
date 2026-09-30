package place

import (
	"context"
	"errors"
	"math"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yetbota/content-service/internal/db"
	"github.com/yetbota/content-service/internal/domain/apperr"
	domain "github.com/yetbota/content-service/internal/domain/place"
)

var _ domain.Repository = (*Repo)(nil)

type Repo struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func New(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool, q: db.New(pool)}
}

func (r *Repo) Create(ctx context.Context, in domain.CreateInput) (domain.Place, error) {
	row, err := r.q.CreatePlace(ctx, db.CreatePlaceParams{
		ID:        pgUUID(uuid.New()),
		Name:      in.Name,
		Lat:       in.Lat,
		Lng:       in.Lng,
		Summary:   "",
		SummaryModel: "",
		Embedding: in.Embedding,
	})
	if err != nil {
		return domain.Place{}, apperr.Wrap(err, apperr.CodeInternal, "create place")
	}
	return placeFromDB(row)
}

func (r *Repo) Get(ctx context.Context, id uuid.UUID) (domain.Place, error) {
	row, err := r.q.GetPlaceByID(ctx, pgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Place{}, apperr.ErrNotFound
		}
		return domain.Place{}, apperr.Wrap(err, apperr.CodeInternal, "get place")
	}
	return placeFromDB(row)
}

func (r *Repo) GetByPostID(ctx context.Context, postID uuid.UUID) (domain.Place, error) {
	row, err := r.q.GetPlaceByPostID(ctx, pgUUID(postID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Place{}, apperr.ErrNotFound
		}
		return domain.Place{}, apperr.Wrap(err, apperr.CodeInternal, "get place by post")
	}
	return placeFromDB(row)
}

func (r *Repo) List(ctx context.Context, p domain.ListParams) (domain.PlacePage, error) {
	params := db.ListPlacesParams{
		CursorCreatedAt: cursorTime(p.Cursor),
		CursorID:        cursorID(p.Cursor),
		ResultLimit:     int32(p.Limit + 1),
	}
	if q := p.Query; q != "" {
		params.Query = pgtype.Text{String: q, Valid: true}
	}
	rows, err := r.q.ListPlaces(ctx, params)
	if err != nil {
		return domain.PlacePage{}, apperr.Wrap(err, apperr.CodeInternal, "list places")
	}
	return placePage(rows, p.Limit)
}

func (r *Repo) ListNearby(ctx context.Context, p domain.NearbyParams) ([]domain.Place, error) {
	minLat, maxLat, minLng, maxLng := bbox(p.Lat, p.Lng, p.RadiusM)
	limit := p.Limit
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.q.ListPlacesNearby(ctx, db.ListPlacesNearbyParams{
		MinLat:      minLat,
		MaxLat:      maxLat,
		MinLng:      minLng,
		MaxLng:      maxLng,
		ResultLimit: int32(limit * 3), // over-fetch then filter by Haversine
	})
	if err != nil {
		return nil, apperr.Wrap(err, apperr.CodeInternal, "list places nearby")
	}
	out := make([]domain.Place, 0, len(rows))
	for _, row := range rows {
		pl, err := placeFromDB(row)
		if err != nil {
			return nil, err
		}
		if haversineM(p.Lat, p.Lng, pl.Lat, pl.Lng) <= p.RadiusM {
			out = append(out, pl)
			if len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

func (r *Repo) AttachPost(ctx context.Context, in domain.AttachInput) (domain.Place, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Place{}, apperr.Wrap(err, apperr.CodeInternal, "begin attach")
	}
	defer tx.Rollback(ctx)
	q := r.q.WithTx(tx)

	prev, err := q.GetPlacePost(ctx, pgUUID(in.PostID))
	hasPrev := err == nil
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return domain.Place{}, apperr.Wrap(err, apperr.CodeInternal, "get place post")
	}

	_, err = q.AttachPostToPlace(ctx, db.AttachPostToPlaceParams{
		PlaceID:      pgUUID(in.PlaceID),
		PostID:       pgUUID(in.PostID),
		AttachScore:  in.AttachScore,
		AttachMethod: in.AttachMethod,
	})
	if err != nil {
		return domain.Place{}, apperr.Wrap(err, apperr.CodeInternal, "attach post")
	}

	if hasPrev {
		oldID, err := uuidFromPg(prev.PlaceID)
		if err != nil {
			return domain.Place{}, err
		}
		if oldID != in.PlaceID {
			if _, err := q.AdjustPlacePostCount(ctx, db.AdjustPlacePostCountParams{
				Delta: -1, ID: pgUUID(oldID),
			}); err != nil {
				return domain.Place{}, apperr.Wrap(err, apperr.CodeInternal, "decrement old place")
			}
			if _, err := q.AdjustPlacePostCount(ctx, db.AdjustPlacePostCountParams{
				Delta: 1, ID: pgUUID(in.PlaceID),
			}); err != nil {
				return domain.Place{}, apperr.Wrap(err, apperr.CodeInternal, "increment new place")
			}
		}
	} else {
		if _, err := q.AdjustPlacePostCount(ctx, db.AdjustPlacePostCountParams{
			Delta: 1, ID: pgUUID(in.PlaceID),
		}); err != nil {
			return domain.Place{}, apperr.Wrap(err, apperr.CodeInternal, "increment place")
		}
	}

	if len(in.Embedding) > 0 {
		if _, err := q.UpdatePlaceEmbedding(ctx, db.UpdatePlaceEmbeddingParams{
			ID: pgUUID(in.PlaceID), Embedding: in.Embedding,
		}); err != nil {
			return domain.Place{}, apperr.Wrap(err, apperr.CodeInternal, "update place embedding")
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Place{}, apperr.Wrap(err, apperr.CodeInternal, "commit attach")
	}
	return r.Get(ctx, in.PlaceID)
}

func (r *Repo) CreateAndAttach(ctx context.Context, in domain.CreateAndAttachInput) (domain.Place, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Place{}, apperr.Wrap(err, apperr.CodeInternal, "begin create-attach")
	}
	defer tx.Rollback(ctx)
	q := r.q.WithTx(tx)

	placeID := uuid.New()
	_, err = q.CreatePlace(ctx, db.CreatePlaceParams{
		ID: pgUUID(placeID), Name: in.Name, Lat: in.Lat, Lng: in.Lng,
		Summary: "", SummaryModel: "", Embedding: in.Embedding,
	})
	if err != nil {
		return domain.Place{}, apperr.Wrap(err, apperr.CodeInternal, "create place")
	}

	prev, err := q.GetPlacePost(ctx, pgUUID(in.PostID))
	hasPrev := err == nil
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return domain.Place{}, apperr.Wrap(err, apperr.CodeInternal, "get place post")
	}

	_, err = q.AttachPostToPlace(ctx, db.AttachPostToPlaceParams{
		PlaceID: pgUUID(placeID), PostID: pgUUID(in.PostID),
		AttachScore: in.AttachScore, AttachMethod: in.AttachMethod,
	})
	if err != nil {
		return domain.Place{}, apperr.Wrap(err, apperr.CodeInternal, "attach post")
	}

	if hasPrev {
		oldID, err := uuidFromPg(prev.PlaceID)
		if err != nil {
			return domain.Place{}, err
		}
		if _, err := q.AdjustPlacePostCount(ctx, db.AdjustPlacePostCountParams{
			Delta: -1, ID: pgUUID(oldID),
		}); err != nil {
			return domain.Place{}, apperr.Wrap(err, apperr.CodeInternal, "decrement old place")
		}
	}
	if _, err := q.AdjustPlacePostCount(ctx, db.AdjustPlacePostCountParams{
		Delta: 1, ID: pgUUID(placeID),
	}); err != nil {
		return domain.Place{}, apperr.Wrap(err, apperr.CodeInternal, "increment place")
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Place{}, apperr.Wrap(err, apperr.CodeInternal, "commit create-attach")
	}
	return r.Get(ctx, placeID)
}

func (r *Repo) ListPostIDs(ctx context.Context, placeID uuid.UUID, limit int, cursor *domain.AttachCursor) (domain.PostIDPage, error) {
	params := db.ListPlacePostIDsParams{
		PlaceID:     pgUUID(placeID),
		ResultLimit: int32(limit + 1),
	}
	if cursor != nil {
		params.CursorAttachedAt = pgtype.Timestamptz{Time: cursor.AttachedAt, Valid: true}
		params.CursorID = pgUUID(cursor.PostID)
	}
	rows, err := r.q.ListPlacePostIDs(ctx, params)
	if err != nil {
		return domain.PostIDPage{}, apperr.Wrap(err, apperr.CodeInternal, "list place posts")
	}
	ids := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		id, err := uuidFromPg(row.PostID)
		if err != nil {
			return domain.PostIDPage{}, err
		}
		ids = append(ids, id)
	}
	var next *domain.AttachCursor
	if len(ids) > limit {
		ids = ids[:limit]
		last := rows[limit-1]
		pid, err := uuidFromPg(last.PostID)
		if err != nil {
			return domain.PostIDPage{}, err
		}
		next = &domain.AttachCursor{AttachedAt: last.AttachedAt.Time, PostID: pid}
	}
	return domain.PostIDPage{PostIDs: ids, NextCursor: next}, nil
}

func (r *Repo) UpdateSummary(ctx context.Context, in domain.UpdateSummaryInput) (domain.Place, error) {
	row, err := r.q.UpdatePlaceSummary(ctx, db.UpdatePlaceSummaryParams{
		ID: pgUUID(in.PlaceID), Summary: in.Summary, SummaryModel: in.SummaryModel,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Place{}, apperr.ErrNotFound
		}
		return domain.Place{}, apperr.Wrap(err, apperr.CodeInternal, "update place summary")
	}
	return placeFromDB(row)
}

func (r *Repo) AddEntry(ctx context.Context, in domain.AddEntryInput) (domain.Entry, error) {
	_, err := r.q.FindSimilarPlaceEntry(ctx, db.FindSimilarPlaceEntryParams{
		PlaceID: pgUUID(in.PlaceID), Category: in.Category, Body: in.Body,
	})
	if err == nil {
		return domain.Entry{}, apperr.Validation(apperr.Field{
			Field: "body", Code: "DUPLICATE", Message: "similar entry already exists",
		})
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return domain.Entry{}, apperr.Wrap(err, apperr.CodeInternal, "check entry dup")
	}

	params := db.CreatePlaceEntryParams{
		ID: pgUUID(uuid.New()), PlaceID: pgUUID(in.PlaceID),
		Category: in.Category, Body: in.Body, Source: in.Source,
	}
	if in.SourcePostID != nil {
		params.SourcePostID = pgUUID(*in.SourcePostID)
	}
	if in.AuthorID != nil {
		params.AuthorID = pgUUID(*in.AuthorID)
	}
	row, err := r.q.CreatePlaceEntry(ctx, params)
	if err != nil {
		return domain.Entry{}, apperr.Wrap(err, apperr.CodeInternal, "create place entry")
	}
	return entryFromDB(row)
}

func (r *Repo) ListEntries(ctx context.Context, placeID uuid.UUID, p domain.EntryListParams) (domain.EntryPage, error) {
	params := db.ListPlaceEntriesParams{
		PlaceID:         pgUUID(placeID),
		CursorCreatedAt: cursorTime(p.Cursor),
		CursorID:        cursorID(p.Cursor),
		ResultLimit:     int32(p.Limit + 1),
	}
	if p.Category != "" {
		params.Category = pgtype.Text{String: p.Category, Valid: true}
	}
	rows, err := r.q.ListPlaceEntries(ctx, params)
	if err != nil {
		return domain.EntryPage{}, apperr.Wrap(err, apperr.CodeInternal, "list place entries")
	}
	entries := make([]domain.Entry, 0, len(rows))
	for _, row := range rows {
		e, err := entryFromDB(row)
		if err != nil {
			return domain.EntryPage{}, err
		}
		entries = append(entries, e)
	}
	var next *domain.PageCursor
	if len(entries) > p.Limit {
		entries = entries[:p.Limit]
		last := entries[p.Limit-1]
		next = &domain.PageCursor{CreatedAt: last.CreatedAt, ID: last.ID}
	}
	return domain.EntryPage{Entries: entries, NextCursor: next}, nil
}

func placePage(rows []db.Place, limit int) (domain.PlacePage, error) {
	places := make([]domain.Place, 0, len(rows))
	for _, row := range rows {
		pl, err := placeFromDB(row)
		if err != nil {
			return domain.PlacePage{}, err
		}
		places = append(places, pl)
	}
	var next *domain.PageCursor
	if len(places) > limit {
		places = places[:limit]
		last := places[limit-1]
		next = &domain.PageCursor{CreatedAt: last.CreatedAt, ID: last.ID}
	}
	return domain.PlacePage{Places: places, NextCursor: next}, nil
}

func placeFromDB(row db.Place) (domain.Place, error) {
	id, err := uuidFromPg(row.ID)
	if err != nil {
		return domain.Place{}, err
	}
	p := domain.Place{
		ID: id, Name: row.Name, Lat: row.Lat, Lng: row.Lng,
		Summary: row.Summary, SummaryModel: row.SummaryModel,
		PostCount: row.PostCount, Embedding: row.Embedding,
		CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time,
	}
	if row.SummaryUpdatedAt.Valid {
		t := row.SummaryUpdatedAt.Time
		p.SummaryUpdatedAt = &t
	}
	return p, nil
}

func entryFromDB(row db.PlaceEntry) (domain.Entry, error) {
	id, err := uuidFromPg(row.ID)
	if err != nil {
		return domain.Entry{}, err
	}
	pid, err := uuidFromPg(row.PlaceID)
	if err != nil {
		return domain.Entry{}, err
	}
	e := domain.Entry{
		ID: id, PlaceID: pid, Category: row.Category, Body: row.Body,
		Source: row.Source, CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time,
	}
	if row.SourcePostID.Valid {
		sid, err := uuidFromPg(row.SourcePostID)
		if err != nil {
			return domain.Entry{}, err
		}
		e.SourcePostID = &sid
	}
	if row.AuthorID.Valid {
		aid, err := uuidFromPg(row.AuthorID)
		if err != nil {
			return domain.Entry{}, err
		}
		e.AuthorID = &aid
	}
	return e, nil
}

func pgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func uuidFromPg(v pgtype.UUID) (uuid.UUID, error) {
	if !v.Valid {
		return uuid.Nil, apperr.New(apperr.CodeInternal, "invalid uuid")
	}
	return uuid.UUID(v.Bytes), nil
}

func cursorTime(c *domain.PageCursor) pgtype.Timestamptz {
	if c == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: c.CreatedAt, Valid: true}
}

func cursorID(c *domain.PageCursor) pgtype.UUID {
	if c == nil {
		return pgtype.UUID{Valid: false}
	}
	return pgUUID(c.ID)
}

func bbox(lat, lng, radiusM float64) (minLat, maxLat, minLng, maxLng float64) {
	const metersPerDegLat = 111_320.0
	dLat := radiusM / metersPerDegLat
	cosLat := math.Cos(lat * math.Pi / 180)
	if cosLat < 0.01 {
		cosLat = 0.01
	}
	dLng := radiusM / (metersPerDegLat * cosLat)
	return lat - dLat, lat + dLat, lng - dLng, lng + dLng
}

func haversineM(lat1, lng1, lat2, lng2 float64) float64 {
	const r = 6_371_000.0
	toRad := math.Pi / 180
	dLat := (lat2 - lat1) * toRad
	dLng := (lng2 - lng1) * toRad
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*toRad)*math.Cos(lat2*toRad)*math.Sin(dLng/2)*math.Sin(dLng/2)
	return 2 * r * math.Asin(math.Sqrt(a))
}
