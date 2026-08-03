package postphoto

import (
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
	"github.com/beka-birhanu/yetbota/content-service/internal/domain/postphoto"
)

func buildFilterMods(opts *postphoto.Options) []qm.QueryMod {
	if opts == nil || len(opts.PostIDs) == 0 {
		return nil
	}
	return []qm.QueryMod{dbmodels.PostPhotoWhere.PostID.IN(opts.PostIDs)}
}

func buildLoadMods(opts *postphoto.Options) []qm.QueryMod {
	if opts == nil || !opts.LoadPhoto {
		return nil
	}
	return []qm.QueryMod{qm.Load(dbmodels.PostPhotoRels.Photo)}
}

func buildQueryMods(opts *postphoto.Options) []qm.QueryMod {
	mods := buildFilterMods(opts)
	return append(mods, buildLoadMods(opts)...)
}
