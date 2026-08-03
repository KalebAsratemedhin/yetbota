package shared

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	toddlerr "github.com/beka-birhanu/toddler/error"
	ctxYB "github.com/beka-birhanu/yetbota/identity-service/internal/domain/context"
	kithttp "github.com/go-kit/kit/transport/http"
)

type Envelope struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

func ServerOptions() []kithttp.ServerOption {
	return []kithttp.ServerOption{
		kithttp.ServerBefore(injectAppSession()),
		kithttp.ServerErrorEncoder(ErrorEncoder),
		kithttp.ServerFinalizer(finalizeRequest()),
	}
}

func ErrorEncoder(ctx context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")

	var te *toddlerr.Error
	if ok := todlerrAs(err, &te); ok && te != nil {
		httpStatus := int(te.PublicStatusCode) / 10
		if httpStatus < 100 || httpStatus > 599 {
			httpStatus = http.StatusInternalServerError
		}
		w.WriteHeader(httpStatus)
		_ = json.NewEncoder(w).Encode(Envelope{
			Success: false,
			Message: te.PublicMessage,
		})
		return
	}

	w.WriteHeader(http.StatusInternalServerError)
	_ = json.NewEncoder(w).Encode(Envelope{
		Success: false,
		Message: "something went wrong",
	})
}

func injectAppSession() func(ctx context.Context, r *http.Request) context.Context {
	return func(ctx context.Context, r *http.Request) context.Context {
		ctxSess := ctxYB.New()
		ctxSess.SetHeader(r.Header).SetMethod(r.Method).SetURL(r.URL.Path)
		return context.WithValue(ctx, ctxYB.AppSession, ctxSess)
	}
}

func finalizeRequest() func(ctx context.Context, code int, r *http.Request) {
	return func(ctx context.Context, code int, r *http.Request) {
		data := ctx.Value(ctxYB.AppSession)
		ctxSess, ok := data.(*ctxYB.Context)
		if !ok || ctxSess == nil {
			return
		}

		ctxSess.SetHeader(r.Header).SetMethod(r.Method).SetURL(r.URL.Path)
		ctxSess.SetResponseCode(strconv.Itoa(code))
		if ctxSess.Response == nil {
			ctxSess.Response = Envelope{Success: code < 400}
		}

		ctxSess.Lv4()
	}
}

func todlerrAs(err error, target **toddlerr.Error) bool {
	if err == nil {
		return false
	}
	if te, ok := err.(*toddlerr.Error); ok {
		*target = te
		return true
	}
	return false
}

