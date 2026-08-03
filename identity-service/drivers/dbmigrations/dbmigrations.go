package dbmigrations

import (
	"database/sql"
	"fmt"

	"github.com/beka-birhanu/yetbota/identity-service/drivers/validator"
	"github.com/pressly/goose"
)

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

func RunDBMigrations(c *Config) (*sql.DB, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	dbGoose, err := goose.OpenDBWithDriver("postgres", c.getDSN())
	if err != nil {
		return nil, err
	}

	return dbGoose, nil
}
