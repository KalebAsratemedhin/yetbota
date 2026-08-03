package feed

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	feedSvc "github.com/beka-birhanu/yetbota/content-service/internal/services/usecase/feed"
)

func decodeFeedGetHTTP(_ context.Context, r *http.Request) (any, error) {
	q := r.URL.Query()
	pageSize, _ := strconv.Atoi(q.Get("page_size"))
	cursor := q.Get("cursor")
	return &feedSvc.ListRequest{Cursor: cursor, PageSize: pageSize}, nil
}

func decodeFeedMarkViewedHTTP(_ context.Context, r *http.Request) (any, error) {
	var body struct {
		PostIDs []string `json:"post_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}
	return &feedSvc.MarkViewedRequest{PostIDs: body.PostIDs}, nil
}
