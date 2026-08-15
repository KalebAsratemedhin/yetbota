package constants

const (
	DefaultPaginationLength = 15
	MaxPaginationLength     = 20
)

const (
	MB                    = 1 << (10 * 2)
	MaxUploadSize         = 20 * MB
	URLExpiration         = 30
	MaxImageResolution    = 3840
	WebImageResolution    = 1080
	MobileImageResolution = 750
)

const (
	MigrationFolder = "migrations"
)

const (
	RoleAdmin = "ADMIN"
	RoleUser  = "USER"
)

const (
	ModerationStatusVisible   = "VISIBLE"
	ModerationStatusHidden    = "HIDDEN"
	ModerationStatusRemoved   = "REMOVED"
	ModerationStatusDuplicate = "DUPLICATE"
)

const (
	ContributionOriginUser          = "USER"
	ContributionOriginPostExtracted = "POST_EXTRACTED"
)

const (
	PlaceStatusPending   = "PENDING"
	PlaceStatusVisible   = "VISIBLE"
	PlaceStatusHidden    = "HIDDEN"
	PlaceStatusRemoved   = "REMOVED"
	PlaceStatusDuplicate = "DUPLICATE"
)

const (
	VerdictDuplicate = "duplicate"
)

const (
	PostScoringStreamKey = "POST_SCORING_STREAM"
)

const (
	FeedFanOutBatchSize   = 500
	FanOutBatchTTLSeconds = 3600
)

var SkipAuthHTTP = map[string]struct{}{
	"GET /posts/":        {},
	"GET /posts/{id}":    {},
	"GET /comments/":     {},
	"GET /comments/{id}": {},
	"GET /places/":       {},
	"GET /places/{id}":   {},
	"GET /places/{id}/contributions": {},
	"GET /places/{id}/posts":         {},
	"GET /questions/":    {},
	"GET /questions/near": {},
	"GET /place-feed/":   {},
}

var AllowedAccessMap = map[string]struct{}{
	RoleAdmin: {},
	RoleUser:  {},
}

var AllowedAdminAccessMap = map[string]struct{}{
	RoleAdmin: {},
}

var AllowedCSAAccessMap = map[string]struct{}{
	RoleAdmin: {},
}
