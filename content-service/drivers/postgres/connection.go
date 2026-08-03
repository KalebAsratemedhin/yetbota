package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	_ "github.com/lib/pq"
)

const postgresDBMS = "postgres"

type Config struct {
	Host     string `yaml:"host" mapstructure:"host" validate:"required"`
	Port     string `yaml:"port" mapstructure:"port" validate:"required"`
	User     string `yaml:"user" mapstructure:"user" validate:"required"`
	Password string `yaml:"password" mapstructure:"password" validate:"required"`
	DB       string `yaml:"db" mapstructure:"db" validate:"required"`
	SslMode  string `yaml:"sslmode" mapstructure:"sslmode"`
}

func (c *Config) sslMode() string {
	if c.SslMode == "" {
		return "require"
	}
	return c.SslMode
}

func (c *Config) Validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return err
	}
	return nil
}

func (c *Config) getDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DB, c.sslMode())
}

func NewDB(c *Config) (*sql.DB, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	db, err := sql.Open(postgresDBMS, c.getDSN())
	if err != nil {
		return nil, err
	}

	var pingErr error
	for attempt := 1; attempt <= 30; attempt++ {
		pingErr = db.Ping()
		if pingErr == nil {
			return db, nil
		}
		if attempt < 30 {
			time.Sleep(2 * time.Second)
		}
	}
	_ = db.Close()
	return nil, pingErr
}
