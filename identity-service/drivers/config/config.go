package config

type Configs struct {
	Debug      bool        `yaml:"debug" mapstructure:"debug"`
	App        *App        `yaml:"app" mapstructure:"app" validate:"required"`
	Rest       *Rest       `yaml:"rest" mapstructure:"rest" validate:"required"`
	Cors       *Cors       `yaml:"cors" mapstructure:"cors" validate:"required"`
	Postgres   *Postgres   `yaml:"database" mapstructure:"database" validate:"required"`
	Redis      *Redis      `yaml:"redis" mapstructure:"redis" validate:"required"`
	Jwt        *Jwt        `yaml:"jwt" mapstructure:"jwt" validate:"required"`
	Cloudinary *Cloudinary `yaml:"cloudinary" mapstructure:"cloudinary" validate:"required"`
	SurrealDB  *SurrealDB  `yaml:"surrealdb" mapstructure:"surrealdb" validate:"required"`
}

type SurrealDB struct {
	URL       string `yaml:"url" mapstructure:"url" validate:"required"`
	Username  string `yaml:"username" mapstructure:"username" validate:"required"`
	Password  string `yaml:"password" mapstructure:"password" validate:"required"`
	Namespace string `yaml:"namespace" mapstructure:"namespace" validate:"required"`
	Database  string `yaml:"database" mapstructure:"database" validate:"required"`
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
	Address      string `yaml:"address" mapstructure:"address" validate:"required"`
	Password     string `yaml:"password" mapstructure:"password"`
	PoolTimeout  string `yaml:"poolTimeout" mapstructure:"poolTimeout" validate:"required"`
	IdleTimeout  string `yaml:"idleTimeout" mapstructure:"idleTimeout" validate:"required"`
	ReadTimeout  string `yaml:"readTimeout" mapstructure:"readTimeout" validate:"required"`
	WriteTimeout string `yaml:"writeTimeout" mapstructure:"writeTimeout" validate:"required"`
	TLS          bool   `yaml:"tls" mapstructure:"tls"`
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
