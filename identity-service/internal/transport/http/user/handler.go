package user

import (
	"net/http"

	driverMW "github.com/beka-birhanu/yetbota/identity-service/drivers/middleware"
	domainAuth "github.com/beka-birhanu/yetbota/identity-service/internal/domain/auth"
	"github.com/beka-birhanu/yetbota/identity-service/internal/services/endpoint"
	"github.com/beka-birhanu/yetbota/identity-service/internal/transport/http/shared"
	kithttp "github.com/go-kit/kit/transport/http"
)

type Config struct {
	E              *endpoint.Endpoints
	SessionManager domainAuth.SessionManager
}

func NewHandler(cfg *Config) (http.Handler, error) {
	mux := http.NewServeMux()

	if cfg != nil && cfg.E != nil {
		listEndpoint := cfg.E.UserList
		updateEndpoint := cfg.E.UserUpdate
		updateSelfEndpoint := cfg.E.UserUpdateSelf
		deleteEndpoint := cfg.E.UserDelete
		deleteSelfEndpoint := cfg.E.UserDeleteSelf
		uploadProfileEndpoint := cfg.E.UserUploadProfile
		followEndpoint := cfg.E.UserFollow
		unfollowEndpoint := cfg.E.UserUnfollow
		readMeEndpoint := cfg.E.UserRead
		if cfg.SessionManager != nil {
			readMeEndpoint = driverMW.AuthMiddleware(cfg.SessionManager)(readMeEndpoint)
			updateEndpoint = driverMW.AuthMiddleware(cfg.SessionManager)(updateEndpoint)
			updateSelfEndpoint = driverMW.AuthMiddleware(cfg.SessionManager)(updateSelfEndpoint)
			deleteEndpoint = driverMW.AuthMiddleware(cfg.SessionManager)(deleteEndpoint)
			deleteSelfEndpoint = driverMW.AuthMiddleware(cfg.SessionManager)(deleteSelfEndpoint)
			uploadProfileEndpoint = driverMW.AuthMiddleware(cfg.SessionManager)(uploadProfileEndpoint)
			followEndpoint = driverMW.AuthMiddleware(cfg.SessionManager)(followEndpoint)
			unfollowEndpoint = driverMW.AuthMiddleware(cfg.SessionManager)(unfollowEndpoint)
		}
		listServer := kithttp.NewServer(
			listEndpoint,
			decodeUserListHTTP,
			encodeUserListHTTP,
			shared.ServerOptions()...,
		)
		readMeServer := kithttp.NewServer(
			readMeEndpoint,
			decodeUserReadMeHTTP,
			encodeUserReadHTTP,
			shared.ServerOptions()...,
		)
		readPublicServer := kithttp.NewServer(
			cfg.E.UserReadPublic,
			decodeUserReadPublicHTTP,
			encodeUserReadPublicHTTP,
			shared.ServerOptions()...,
		)
		registerServer := kithttp.NewServer(
			cfg.E.UserRegister,
			decodeUserRegisterHTTP,
			encodeUserRegisterHTTP,
			shared.ServerOptions()...,
		)
		updateServer := kithttp.NewServer(
			updateEndpoint,
			decodeUserUpdateHTTP,
			encodeUserUpdateHTTP,
			shared.ServerOptions()...,
		)
		updateSelfServer := kithttp.NewServer(
			updateSelfEndpoint,
			decodeUserUpdateSelfHTTP,
			encodeUserUpdateSelfHTTP,
			shared.ServerOptions()...,
		)
		deleteServer := kithttp.NewServer(
			deleteEndpoint,
			decodeUserDeleteHTTP,
			encodeUserDeleteHTTP,
			shared.ServerOptions()...,
		)
		deleteSelfServer := kithttp.NewServer(
			deleteSelfEndpoint,
			decodeUserDeleteSelfHTTP,
			encodeUserDeleteSelfHTTP,
			shared.ServerOptions()...,
		)
		uploadProfileServer := kithttp.NewServer(
			uploadProfileEndpoint,
			decodeUserUploadProfileHTTP,
			encodeUserUploadProfileHTTP,
			shared.ServerOptions()...,
		)
		followServer := kithttp.NewServer(
			followEndpoint,
			decodeUserFollowHTTP,
			encodeUserFollowHTTP,
			shared.ServerOptions()...,
		)
		unfollowServer := kithttp.NewServer(
			unfollowEndpoint,
			decodeUserUnfollowHTTP,
			encodeUserUnfollowHTTP,
			shared.ServerOptions()...,
		)

		mux.Handle("GET /", listServer)
		mux.Handle("GET /me", readMeServer)
		mux.Handle("GET /{id}", readPublicServer)
		mux.Handle("POST /register", registerServer)
		mux.Handle("PATCH /{id}", updateServer)
		mux.Handle("PATCH /self", updateSelfServer)
		mux.Handle("DELETE /{id}", deleteServer)
		mux.Handle("DELETE /self", deleteSelfServer)
		mux.Handle("POST /profile", uploadProfileServer)
		mux.Handle("POST /follow", followServer)
		mux.Handle("POST /unfollow", unfollowServer)
	}

	return mux, nil
}

