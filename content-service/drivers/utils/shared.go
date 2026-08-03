package utils

import (
	"fmt"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/content-service/drivers/constants"
	ctxRP "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
)

func IsAdmin(ctxSess *ctxRP.Context) bool {
	return ctxSess.UserSession.Role == constants.RoleAdmin
}

func EnsureVisible(ctxSess *ctxRP.Context, moderationStatus, kind, id string) error {
	if moderationStatus == constants.ModerationStatusVisible || IsAdmin(ctxSess) {
		return nil
	}
	return &toddlerr.Error{
		PublicStatusCode:  status.NotFound,
		ServiceStatusCode: status.NotFound,
		PublicMessage:     fmt.Sprintf("%s not found", kind),
		ServiceMessage:    fmt.Sprintf("%s %s has moderation_status %s", kind, id, moderationStatus),
	}
}

func AllowAccess(ctxSess *ctxRP.Context) error {
	if _, ok := constants.AllowedAccessMap[ctxSess.UserSession.Role]; ok {
		return nil
	}
	return nil
}

func AllowAdminOrCSAAccess(ctxSess *ctxRP.Context) error {
	_, ok := constants.AllowedCSAAccessMap[ctxSess.UserSession.Role]
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
