package place

import (
	"time"

	"github.com/aarondl/null/v8"
	"github.com/beka-birhanu/yetbota/content-service/drivers/geotypes"
)

type Place struct {
	ID                  string
	Name                string
	Description         string
	Address             null.String
	Location            geotypes.NullPoint
	CreatedBy           string
	Status              string
	DuplicateOf         null.String
	ConfidenceScore     int16
	LastActivityAt      null.Time
	LastConfirmedAt     null.Time
	NeedsCheck          bool
	SummaryText         null.String
	SummaryGeneratedAt  null.Time
	SummaryCitations    []byte
	CreatedAt           time.Time
	UpdatedAt           time.Time
}
