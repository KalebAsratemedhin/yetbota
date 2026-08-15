package config

import "time"

type Configs struct {
	Debug        bool          `yaml:"debug" mapstructure:"debug"`
	App          *App          `yaml:"app" mapstructure:"app" validate:"required"`
	Rest         *Rest         `yaml:"rest" mapstructure:"rest" validate:"required"`
	Cors         *Cors         `yaml:"cors" mapstructure:"cors" validate:"required"`
	Postgres     *Postgres     `yaml:"database" mapstructure:"database" validate:"required"`
	Redis        *Redis        `yaml:"redis" mapstructure:"redis" validate:"required"`
	Jwt          *Jwt          `yaml:"jwt" mapstructure:"jwt" validate:"required"`
	Cloudinary   *Cloudinary   `yaml:"cloudinary" mapstructure:"cloudinary" validate:"required"`
	AIService      *AIService      `yaml:"ai_service" mapstructure:"ai_service" validate:"required"`
	RabbitMQ       *RabbitMQ       `yaml:"rabbitmq" mapstructure:"rabbitmq" validate:"required"`
	Feed           *Feed           `yaml:"feed" mapstructure:"feed" validate:"required"`
	AuthorRating   *AuthorRating   `yaml:"author_rating" mapstructure:"author_rating" validate:"required"`
	Moderation     *Moderation     `yaml:"moderation" mapstructure:"moderation" validate:"required"`
	SurrealDB      *SurrealDB      `yaml:"surrealdb" mapstructure:"surrealdb" validate:"required"`
	Truth          *Truth          `yaml:"truth" mapstructure:"truth"`
	Summary        *Summary        `yaml:"summary" mapstructure:"summary"`
	PlaceLinking   *PlaceLinking   `yaml:"place_linking" mapstructure:"place_linking"`
}

type Summary struct {
	DebounceMinutes        int `yaml:"debounce_minutes" mapstructure:"debounce_minutes"`
	MinContributions       int `yaml:"min_contributions" mapstructure:"min_contributions"`
	MaxSourceContributions int `yaml:"max_source_contributions" mapstructure:"max_source_contributions"`
}

type PlaceLinking struct {
	MinMatchScore float64 `yaml:"min_match_score" mapstructure:"min_match_score"`
	LinkRadiusKm  float64 `yaml:"link_radius_km" mapstructure:"link_radius_km"`
}

func (p *PlaceLinking) MinMatchScoreOrDefault() float64 {
	if p == nil || p.MinMatchScore <= 0 {
		return 0.75
	}
	return p.MinMatchScore
}

func (p *PlaceLinking) LinkRadiusKmOrDefault() float64 {
	if p == nil || p.LinkRadiusKm <= 0 {
		return 0.5
	}
	return p.LinkRadiusKm
}

func (s *Summary) DebounceSecondsOrDefault() int64 {
	mins := 15
	if s != nil && s.DebounceMinutes > 0 {
		mins = s.DebounceMinutes
	}
	return int64(mins) * 60
}

type SurrealDB struct {
	URL                string `yaml:"url" mapstructure:"url" validate:"required"`
	Username           string `yaml:"username" mapstructure:"username" validate:"required"`
	Password           string `yaml:"password" mapstructure:"password" validate:"required"`
	Namespace          string `yaml:"namespace" mapstructure:"namespace" validate:"required"`
	FollowDatabase     string `yaml:"follow_database" mapstructure:"follow_database" validate:"required"`
	SimilarityDatabase string `yaml:"similarity_database" mapstructure:"similarity_database" validate:"required"`
}

type Moderation struct {
	AutoHideThreshold  int `yaml:"auto_hide_threshold" mapstructure:"auto_hide_threshold" validate:"required"`
	RateLimitMax       int `yaml:"rate_limit_max" mapstructure:"rate_limit_max" validate:"required"`
	RateLimitWindowSec int `yaml:"rate_limit_window_sec" mapstructure:"rate_limit_window_sec" validate:"required"`
}

type Truth struct {
	DecayDays          int     `yaml:"decay_days" mapstructure:"decay_days"`
	VerificationBoost  float64 `yaml:"verification_boost" mapstructure:"verification_boost"`
	StaleMarkIntervalM int     `yaml:"stale_mark_interval_min" mapstructure:"stale_mark_interval_min"`
}

func (t *Truth) DecayDaysOrDefault() int {
	if t == nil || t.DecayDays <= 0 {
		return 90
	}
	return t.DecayDays
}

func (t *Truth) VerificationBoostOrDefault() float64 {
	if t == nil || t.VerificationBoost <= 0 {
		return 20
	}
	return t.VerificationBoost
}

func (t *Truth) StaleMarkIntervalOrDefault() time.Duration {
	mins := 60
	if t != nil && t.StaleMarkIntervalM > 0 {
		mins = t.StaleMarkIntervalM
	}
	return time.Duration(mins) * time.Minute
}

type RabbitMQ struct {
	URL string `yaml:"url" mapstructure:"url" validate:"required"`
}

type AIService struct {
	BaseURL string `yaml:"base_url" mapstructure:"base_url" validate:"required"`
	APIKey  string `yaml:"api_key" mapstructure:"api_key"`
	Timeout int    `yaml:"timeout" mapstructure:"timeout" validate:"required"`
}

type Cloudinary struct {
	CloudName string `yaml:"cloud_name" mapstructure:"cloud_name" validate:"required"`
	APIKey    string `yaml:"api_key"    mapstructure:"api_key"    validate:"required"`
	APISecret string `yaml:"api_secret" mapstructure:"api_secret" validate:"required"`
	Folder    string `yaml:"folder"     mapstructure:"folder"`
}

type App struct {
	Name    string `yaml:"name" mapstructure:"name" validate:"required"`
	Version string `yaml:"version" mapstructure:"version" validate:"required"`
}

type Rest struct {
	Port         int `yaml:"port" mapstructure:"port" validate:"required"`
	ReadTimeout  int `yaml:"read_timeout" mapstructure:"read_timeout" validate:"required"`
	WriteTimeout int `yaml:"write_timeout" mapstructure:"write_timeout" validate:"required"`
	IdleTimeout  int `yaml:"idle_timeout" mapstructure:"idle_timeout" validate:"required"`
}

type Cors struct {
	Hosts []string `yaml:"hosts" mapstructure:"hosts" validate:"required"`
}

type Postgres struct {
	DB       string `yaml:"db" mapstructure:"db" validate:"required"`
	Host     string `yaml:"host" mapstructure:"host" validate:"required"`
	Password string `yaml:"password" mapstructure:"password" validate:"required"`
	Port     string `yaml:"port" mapstructure:"port" validate:"required"`
	User     string `yaml:"user" mapstructure:"user" validate:"required"`
	SslMode  string `yaml:"sslmode" mapstructure:"sslmode"`
}

type Redis struct {
	Address  string `yaml:"address"  mapstructure:"address"  validate:"required"`
	Password string `yaml:"password" mapstructure:"password" validate:"required"`

	DB           int `yaml:"db"            mapstructure:"db"`
	PoolSize     int `yaml:"pool_size"     mapstructure:"pool_size"     validate:"required"`
	MinIdleConns int `yaml:"min_idle_conns" mapstructure:"min_idle_conns"`
	MaxIdleConns int `yaml:"max_idle_conns" mapstructure:"max_idle_conns" validate:"required"`
	MaxRetries   int `yaml:"max_retries"   mapstructure:"max_retries"   validate:"required"`

	DialTimeout     int `yaml:"dial_timeout"      mapstructure:"dial_timeout"      validate:"required"`
	ReadTimeout     int `yaml:"read_timeout"      mapstructure:"read_timeout"      validate:"required"`
	WriteTimeout    int `yaml:"write_timeout"     mapstructure:"write_timeout"     validate:"required"`
	PoolTimeout     int `yaml:"pool_timeout"      mapstructure:"pool_timeout"      validate:"required"`
	ConnMaxIdleTime int `yaml:"conn_max_idle_time" mapstructure:"conn_max_idle_time" validate:"required"`
	ConnMaxLifetime int `yaml:"conn_max_lifetime" mapstructure:"conn_max_lifetime" validate:"required"`

	TLS bool `yaml:"tls" mapstructure:"tls"`
}

type Jwt struct {
	Algorithm    string    `yaml:"algorithm" mapstructure:"algorithm" validate:"required"`
	AccessToken  *JwtToken `yaml:"access_token" mapstructure:"access_token" validate:"required"`
	RefreshToken *JwtToken `yaml:"refresh_token" mapstructure:"refresh_token" validate:"required"`
}

type JwtToken struct {
	Expiration int    `yaml:"expiration" mapstructure:"expiration" validate:"required"`
	Secret     string `yaml:"secret" mapstructure:"secret" validate:"required"`
}

type Feed struct {
	HalfLifeHours        float64 `yaml:"half_life_hours"    mapstructure:"half_life_hours"`
	FeedSize             int     `yaml:"feed_size"          mapstructure:"feed_size"`
	ColdStartN           int     `yaml:"cold_start_n"       mapstructure:"cold_start_n"`
	ScoreChangeDelta     float64 `yaml:"score_change_delta" mapstructure:"score_change_delta"`
	SeedBonus            float64 `yaml:"seed_bonus"          mapstructure:"seed_bonus"`
	QScale               float64 `yaml:"q_scale"             mapstructure:"q_scale"`
	Epoch                int64   `yaml:"epoch"               mapstructure:"epoch"`
	ScoreTTLHours        float64 `yaml:"score_ttl_hours"    mapstructure:"score_ttl_hours"`
	StaleLimit           int     `yaml:"stale_limit"        mapstructure:"stale_limit"`
	RefillThreshold      int     `yaml:"refill_threshold"   mapstructure:"refill_threshold"`
	FanOutLimit          int     `yaml:"fan_out_limit"      mapstructure:"fan_out_limit"`
	SeenCacheTTL         int64   `yaml:"seen_cache_ttl"         mapstructure:"seen_cache_ttl"`
	MinFeedScore         float64 `yaml:"min_feed_score"         mapstructure:"min_feed_score"`
	CelebrityThreshold   int64   `yaml:"celebrity_threshold"    mapstructure:"celebrity_threshold"`
	MaxCelebrityFeedSize int     `yaml:"max_celebrity_feed_size" mapstructure:"max_celebrity_feed_size"`
}

type Badge struct {
	Name     string `yaml:"name"      mapstructure:"name"`
	MinScore int64  `yaml:"min_score" mapstructure:"min_score"`
}

type AuthorRating struct {
	StabilizingWindowHours int     `yaml:"stabilizing_window_hours" mapstructure:"stabilizing_window_hours" validate:"required"`
	ConsumerGroup          string  `yaml:"consumer_group"           mapstructure:"consumer_group"           validate:"required"`
	PollIntervalSec        int     `yaml:"poll_interval_sec"        mapstructure:"poll_interval_sec"        validate:"required"`
	BatchSize              int64   `yaml:"batch_size"               mapstructure:"batch_size"               validate:"required"`
	ContinueAfterIter      int     `yaml:"continue_after_iter"      mapstructure:"continue_after_iter"      validate:"required"`
	Badges                 []Badge `yaml:"badges"                   mapstructure:"badges"`
}
