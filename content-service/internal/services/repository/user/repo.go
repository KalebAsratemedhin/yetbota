package user

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	domainUser "github.com/beka-birhanu/yetbota/content-service/internal/domain/user"
)

type repository struct {
	db *sql.DB
}

type Config struct {
	DB *sql.DB `validate:"required"`
}

func (c *Config) Validate() error {
	return validator.Validate.Struct(c)
}

func NewRepo(c *Config) (domainUser.Repository, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return &repository{db: c.DB}, nil
}

func (r *repository) Read(ctx context.Context, id string) (*dbmodels.User, error) {
	user, err := dbmodels.FindUser(ctx, r.db, id)
	if err == sql.ErrNoRows {
		return nil, &toddlerr.Error{
			PublicStatusCode:  http.StatusNotFound,
			ServiceStatusCode: http.StatusNotFound,
			PublicMessage:     "user not found",
			ServiceMessage:    fmt.Sprintf("user repo: %s not found", id),
		}
	}
	if err != nil {
		return nil, &toddlerr.Error{
			PublicStatusCode:  http.StatusInternalServerError,
			ServiceStatusCode: http.StatusInternalServerError,
			PublicMessage:     "something went wrong",
			ServiceMessage:    fmt.Sprintf("user repo: read failed: %v", err),
		}
	}
	return user, nil
}

func (r *repository) GetDeviceTokens(ctx context.Context, userID string) ([]string, error) {
	devices, err := dbmodels.UserDevices(
		qm.Select(dbmodels.UserDeviceColumns.Token),
		dbmodels.UserDeviceWhere.UserID.EQ(userID),
		dbmodels.UserDeviceWhere.Token.IsNotNull(),
	).All(ctx, r.db)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, &toddlerr.Error{
			PublicStatusCode:  http.StatusInternalServerError,
			ServiceStatusCode: http.StatusInternalServerError,
			PublicMessage:     "something went wrong",
			ServiceMessage:    fmt.Sprintf("user repo: get device tokens failed: %v", err),
		}
	}
	tokens := make([]string, 0, len(devices))
	for _, d := range devices {
		if d.Token.Valid {
			tokens = append(tokens, d.Token.String)
		}
	}
	return tokens, nil
}

func (r *repository) Update(ctx context.Context, model *dbmodels.User) error {
	_, err := model.Update(ctx, r.db, boil.Infer())
	if err != nil {
		return &toddlerr.Error{
			PublicStatusCode:  http.StatusInternalServerError,
			ServiceStatusCode: http.StatusInternalServerError,
			PublicMessage:     "something went wrong",
			ServiceMessage:    fmt.Sprintf("user repo: update failed: %v", err),
		}
	}
	return nil
}
