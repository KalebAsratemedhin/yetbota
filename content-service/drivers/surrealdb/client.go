package surrealdb

import (
	"context"
	"fmt"

	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	surrealdb "github.com/surrealdb/surrealdb.go"
	"github.com/surrealdb/surrealdb.go/pkg/models"
)

type Client struct {
	db *surrealdb.DB
}

type Config struct {
	URL       string `yaml:"url" mapstructure:"url" validate:"required"`
	Username  string `yaml:"username" mapstructure:"username" validate:"required"`
	Password  string `yaml:"password" mapstructure:"password" validate:"required"`
	Namespace string `yaml:"namespace" mapstructure:"namespace" validate:"required"`
	Database  string `yaml:"database" mapstructure:"database" validate:"required"`
	Schema    string `yaml:"-" mapstructure:"-"`
}

func (c *Config) Validate() error {
	return validator.Validate.Struct(c)
}

func NewClient(ctx context.Context, c *Config) (*Client, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	db, err := surrealdb.FromEndpointURLString(ctx, c.URL)
	if err != nil {
		return nil, fmt.Errorf("surrealdb connect: %w", err)
	}

	if err := db.Use(ctx, c.Namespace, c.Database); err != nil {
		_ = db.Close(ctx)
		return nil, fmt.Errorf("surrealdb use ns/db: %w", err)
	}

	if _, err := db.SignIn(ctx, surrealdb.Auth{
		Username: c.Username,
		Password: c.Password,
	}); err != nil {
		_ = db.Close(ctx)
		return nil, fmt.Errorf("surrealdb signin: %w", err)
	}

	client := &Client{db: db}
	schema := c.Schema
	if schema == "" {
		schema = FollowSchema
	}
	if err := client.ensureSchema(ctx, schema); err != nil {
		_ = db.Close(ctx)
		return nil, fmt.Errorf("surrealdb ensure schema: %w", err)
	}

	return client, nil
}

const FollowSchema = `
DEFINE TABLE OVERWRITE user SCHEMALESS;
DEFINE TABLE OVERWRITE follows TYPE RELATION IN user OUT user;
DEFINE INDEX OVERWRITE uniq_follows ON follows FIELDS in, out UNIQUE;
`

const SimilaritySchema = `
DEFINE TABLE OVERWRITE post SCHEMALESS;
DEFINE TABLE OVERWRITE similar_to TYPE RELATION IN post OUT post;
DEFINE TABLE OVERWRITE duplicate_of TYPE RELATION IN post OUT post;
`

func (c *Client) ensureSchema(ctx context.Context, schema string) error {
	_, err := Query[any](ctx, c, schema, nil)
	return err
}

func (c *Client) Close(ctx context.Context) error {
	if c == nil || c.db == nil {
		return nil
	}
	return c.db.Close(ctx)
}

func (c *Client) DB() *surrealdb.DB {
	return c.db
}

func Query[T any](ctx context.Context, c *Client, query string, vars map[string]any) (T, error) {
	var zero T
	if err := ctx.Err(); err != nil {
		return zero, err
	}
	results, err := surrealdb.Query[T](ctx, c.db, query, vars)
	if err != nil {
		return zero, err
	}
	if results == nil || len(*results) == 0 {
		return zero, nil
	}
	last := (*results)[len(*results)-1]
	if last.Error != nil {
		return zero, last.Error
	}
	return last.Result, nil
}

func UserID(id string) models.RecordID {
	return models.NewRecordID("user", id)
}

func PostID(id string) models.RecordID {
	return models.NewRecordID("post", id)
}
