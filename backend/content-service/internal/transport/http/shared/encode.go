package shared

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/yetbota/content-service/internal/domain/apperr"
)

type errorBody struct {
	Error errorPayload `json:"error"`
}

type errorPayload struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details []apperr.Field `json:"details,omitempty"`
}

func EncodeJSONResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(response)
}

func EncodeError(ctx context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")

	payload := errorPayload{
		Code:    string(apperr.CodeInternal),
		Message: "internal error",
	}
	status := http.StatusInternalServerError

	if ae, ok := apperr.As(err); ok {
		payload.Code = string(ae.Code)
		payload.Message = ae.Message
		payload.Details = ae.Details
		status = apperr.HTTPStatus(ae.Code)
	}

	log := LoggerFromContext(ctx)
	attrs := []any{"err", err.Error(), "status", status, "code", payload.Code}
	if status >= 500 {
		log.Error("request_error", attrs...)
	} else {
		log.Warn("request_error", attrs...)
	}

	w.WriteHeader(status)
	if encErr := json.NewEncoder(w).Encode(errorBody{Error: payload}); encErr != nil {
		slog.Error("encode_error_body_failed", "err", encErr.Error())
	}
}
