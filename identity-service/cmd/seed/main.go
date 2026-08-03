package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/google/uuid"

	"github.com/beka-birhanu/yetbota/identity-service/drivers/config"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/dbmodels"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/postgres"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/utils"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/validator"
)

func main() {
	username := flag.String("username", "", "admin username (required)")
	password := flag.String("password", "", "admin password (required)")
	firstName := flag.String("first-name", "Admin", "admin first name")
	lastName := flag.String("last-name", "User", "admin last name")
	flag.Parse()

	if *username == "" || *password == "" {
		fmt.Println("usage: seed -username <u> -password <p> [-first-name <f>] [-last-name <l>]")
		os.Exit(1)
	}

	validator.InitValidator()

	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Errorf("error load config: %v", err))
	}

	pgdb, err := postgres.NewDB(&postgres.Config{
		Host:     cfg.Postgres.Host,
		Port:     cfg.Postgres.Port,
		User:     cfg.Postgres.User,
		Password: cfg.Postgres.Password,
		DB:       cfg.Postgres.DB,
		SslMode:  cfg.Postgres.SslMode,
	})
	if err != nil {
		panic(fmt.Errorf("error connect postgres: %v", err))
	}
	defer func() {
		_ = pgdb.Close()
	}()

	ctx := context.Background()

	existing, err := dbmodels.Users(dbmodels.UserWhere.Username.EQ(*username)).One(ctx, pgdb)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		panic(fmt.Errorf("error checking existing user: %v", err))
	}

	if existing != nil {
		existing.Role = dbmodels.RolesADMIN
		existing.Status = dbmodels.UserStatusACTIVE
		if _, err := existing.Update(ctx, pgdb, boil.Whitelist(
			dbmodels.UserColumns.Role,
			dbmodels.UserColumns.Status,
		)); err != nil {
			panic(fmt.Errorf("error promoting user to admin: %v", err))
		}
		fmt.Printf("promoted existing user %q to ADMIN (id=%s)\n", *username, existing.ID)
		return
	}

	hashed, err := utils.NewHasher().Hash(*password)
	if err != nil {
		panic(fmt.Errorf("error hashing password: %v", err))
	}

	admin := &dbmodels.User{
		ID:        uuid.NewString(),
		FirstName: *firstName,
		LastName:  *lastName,
		Username:  *username,
		Password:  hashed,
		Role:      dbmodels.RolesADMIN,
		Status:    dbmodels.UserStatusACTIVE,
	}

	if err := admin.Insert(ctx, pgdb, boil.Infer()); err != nil {
		panic(fmt.Errorf("error inserting admin: %v", err))
	}

	fmt.Printf("created admin %q (id=%s)\n", *username, admin.ID)
}
