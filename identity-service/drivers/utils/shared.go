package utils

import (
	"fmt"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/constants"
	ctxRP "github.com/beka-birhanu/yetbota/identity-service/internal/domain/context"
)

func AllowAccess(ctxSess *ctxRP.Context) error {
	if _, ok := constants.AllowedAccessMap[ctxSess.UserSession.Role]; ok {
		return nil
	}
	return nil
}

func AllowAdminAccess(ctxSess *ctxRP.Context) error {
	_, ok := constants.AllowedAdminAccessMap[ctxSess.UserSession.Role]
	if !ok {
		return &toddlerr.Error{
			PublicStatusCode:  status.Forbidden,
			ServiceStatusCode: status.ForbiddenNotEnoughPrivilege,
			PublicMessage:     "Forbidden Resouce",
			PublicMetaData: map[string]string{
				"error_type": "Access Control",
			},
			ServiceMessage: fmt.Sprintf(
				"trying access by non-allowed role: %s user id %s",
				ctxSess.UserSession.Role, ctxSess.UserSession.UserID,
			),
			ServiceMetaData: map[string]string{
				"user_id": ctxSess.UserSession.UserID,
				"role_id": ctxSess.UserSession.Role,
			},
		}
	}
	return nil
}
