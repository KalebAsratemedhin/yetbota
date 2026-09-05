package qa

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/google/uuid"

	"github.com/yetbota/content-service/internal/domain/apperr"
	svcendpoint "github.com/yetbota/content-service/internal/services/endpoint"
	"github.com/yetbota/content-service/internal/transport/http/shared"
)

type Endpoints struct {
	CreateQuestion endpoint.Endpoint
	GetQuestion    endpoint.Endpoint
	ListQuestions  endpoint.Endpoint
	CloseQuestion  endpoint.Endpoint
	QuestionVote   endpoint.Endpoint
	AddAnswer      endpoint.Endpoint
	ListAnswers    endpoint.Endpoint
	AnswerVote     endpoint.Endpoint
}

func Mount(mux *http.ServeMux, eps Endpoints, authn shared.TokenAuthenticator) {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(shared.EncodeError),
	}
	auth := func(ep endpoint.Endpoint, dec kithttp.DecodeRequestFunc) http.Handler {
		return shared.RequireAuth(authn, kithttp.NewServer(ep, dec, shared.EncodeJSONResponse, opts...))
	}

	mux.Handle("POST /questions", auth(eps.CreateQuestion, decodeCreateQuestion))
	mux.Handle("GET /questions", auth(eps.ListQuestions, decodeListQuestions))
	mux.Handle("GET /questions/{id}", auth(eps.GetQuestion, decodeQuestionID))
	mux.Handle("POST /questions/{id}/close", auth(eps.CloseQuestion, decodeQuestionID))
	mux.Handle("POST /questions/{id}/vote", auth(eps.QuestionVote, decodeQuestionVote))
	mux.Handle("POST /questions/{id}/answers", auth(eps.AddAnswer, decodeAddAnswer))
	mux.Handle("GET /questions/{id}/answers", auth(eps.ListAnswers, decodeListAnswers))
	mux.Handle("POST /answers/{id}/vote", auth(eps.AnswerVote, decodeAnswerVote))
}

func decodeCreateQuestion(ctx context.Context, r *http.Request) (interface{}, error) {
	p, err := shared.PrincipalFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	var req svcendpoint.CreateQuestionRequest
	if err := decodeJSON(r, &req); err != nil {
		return nil, err
	}
	req.Principal = p
	return req, nil
}

func decodeListQuestions(ctx context.Context, r *http.Request) (interface{}, error) {
	p, err := shared.PrincipalFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	limit, err := svcendpoint.ParseListLimit(r.URL.Query().Get("limit"))
	if err != nil {
		return nil, err
	}
	var postID *uuid.UUID
	if raw := r.URL.Query().Get("post_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			return nil, apperr.Validation(apperr.Field{
				Field: "post_id", Code: "INVALID", Message: "post_id must be a UUID",
			})
		}
		postID = &id
	}
	return svcendpoint.ListQuestionsRequest{
		Principal: p, PostID: postID, Limit: limit, Cursor: r.URL.Query().Get("cursor"),
	}, nil
}

func decodeQuestionID(ctx context.Context, r *http.Request) (interface{}, error) {
	p, err := shared.PrincipalFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	id, err := parseID(r.PathValue("id"), "id")
	if err != nil {
		return nil, err
	}
	return svcendpoint.QuestionIDRequest{Principal: p, QuestionID: id}, nil
}

func decodeQuestionVote(ctx context.Context, r *http.Request) (interface{}, error) {
	p, err := shared.PrincipalFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	id, err := parseID(r.PathValue("id"), "id")
	if err != nil {
		return nil, err
	}
	var body struct {
		Vote *string `json:"vote"`
	}
	raw, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return nil, apperr.Validation(apperr.Field{Field: "body", Code: "INVALID", Message: "invalid JSON body"})
	}
	if len(bytes.TrimSpace(raw)) > 0 {
		if err := json.Unmarshal(raw, &body); err != nil {
			return nil, apperr.Validation(apperr.Field{Field: "body", Code: "INVALID", Message: "invalid JSON body"})
		}
	}
	return svcendpoint.QuestionVoteRequest{Principal: p, QuestionID: id, Vote: body.Vote}, nil
}

func decodeAddAnswer(ctx context.Context, r *http.Request) (interface{}, error) {
	p, err := shared.PrincipalFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	id, err := parseID(r.PathValue("id"), "id")
	if err != nil {
		return nil, err
	}
	var body struct {
		Body string `json:"body"`
	}
	if err := decodeJSON(r, &body); err != nil {
		return nil, err
	}
	return svcendpoint.AddAnswerRequest{Principal: p, QuestionID: id, Body: body.Body}, nil
}

func decodeListAnswers(ctx context.Context, r *http.Request) (interface{}, error) {
	p, err := shared.PrincipalFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	id, err := parseID(r.PathValue("id"), "id")
	if err != nil {
		return nil, err
	}
	limit, err := svcendpoint.ParseListLimit(r.URL.Query().Get("limit"))
	if err != nil {
		return nil, err
	}
	return svcendpoint.ListAnswersRequest{
		Principal: p, QuestionID: id, Limit: limit, Cursor: r.URL.Query().Get("cursor"),
	}, nil
}

func decodeAnswerVote(ctx context.Context, r *http.Request) (interface{}, error) {
	p, err := shared.PrincipalFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	id, err := parseID(r.PathValue("id"), "id")
	if err != nil {
		return nil, err
	}
	var body struct {
		Vote *string `json:"vote"`
	}
	raw, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return nil, apperr.Validation(apperr.Field{Field: "body", Code: "INVALID", Message: "invalid JSON body"})
	}
	if len(bytes.TrimSpace(raw)) > 0 {
		if err := json.Unmarshal(raw, &body); err != nil {
			return nil, apperr.Validation(apperr.Field{Field: "body", Code: "INVALID", Message: "invalid JSON body"})
		}
	}
	return svcendpoint.AnswerVoteRequest{Principal: p, AnswerID: id, Vote: body.Vote}, nil
}

func parseID(raw, field string) (uuid.UUID, error) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, apperr.Validation(apperr.Field{
			Field: field, Code: "INVALID", Message: field + " must be a UUID",
		})
	}
	return id, nil
}

func decodeJSON(r *http.Request, dest interface{}) error {
	defer r.Body.Close()
	dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	if err := dec.Decode(dest); err != nil {
		return apperr.Validation(apperr.Field{Field: "body", Code: "INVALID", Message: "invalid JSON body"})
	}
	return nil
}
