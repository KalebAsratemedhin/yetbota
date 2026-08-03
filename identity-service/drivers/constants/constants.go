package constants

import (
	"github.com/beka-birhanu/yetbota/identity-service/drivers/dbmodels"
)

const (
	DefaultPaginationLength = 15
)

const (
	MB                 = 1 << (10 * 2)
	MaxUploadSize      = 10 * MB
	URLExpiration      = 30
	MaxImageResolution = 480
	MaxImageSize       = MaxImageResolution * MaxImageResolution * 3
)

const (
	MigrationFolder = "migrations"
)

var SkipAuth = map[string]struct{}{}

var AllowedAccessMap = map[string]struct{}{
	dbmodels.RolesADMIN: {},
	dbmodels.RolesUSER:  {},
}

var AllowedAdminAccessMap = map[string]struct{}{
	dbmodels.RolesADMIN: {},
}
