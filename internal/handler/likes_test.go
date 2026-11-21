package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/andreev012789-hash/likes-matches/internal/model"
	"github.com/andreev012789-hash/likes-matches/internal/repository"
)

type fakeLikeStore struct {
	likeFn func(ctx context.Context, fromID, toID int64) (model.LikeResult, error)
}

func (f *fakeLikeStore) Like(ctx context.Context, fromID, toID int64) (model.LikeResult, error) {
	return f.likeFn(ctx, fromID, toID)
}

func (f *fakeLikeStore) Matches(ctx context.Context, userID int64) ([]model.Profile, error) {
	return nil, nil
}

func TestCreateLike(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		likeResult model.LikeResult
		likeErr    error
		wantStatus int
		wantBody   string
	}{
		{
			name:       "self like",
			body:       `{"from_id":1,"to_id":1}`,
			likeErr:    repository.ErrSelfLike,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "profile not found",
			body:       `{"from_id":1,"to_id":999}`,
			likeErr:    repository.ErrNotFound,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "invalid json",
			body:       `{bad`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "zero id",
			body:       `{"from_id":0,"to_id":2}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "like without match",
			body:       `{"from_id":1,"to_id":2}`,
			wantStatus: http.StatusOK,
			wantBody:   `"matched":false`,
		},
		{
			name:       "like with match",
			body:       `{"from_id":2,"to_id":1}`,
			likeResult: model.LikeResult{Matched: true, Match: &model.Match{UserA: 1, UserB: 2}},
			wantStatus: http.StatusOK,
			wantBody:   `"matched":true`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeLikeStore{
				likeFn: func(ctx context.Context, fromID, toID int64) (model.LikeResult, error) {
					return tt.likeResult, tt.likeErr
				},
			}

			h := New(nil, store)

			req := httptest.NewRequest(http.MethodPost, "/likes", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			h.Router().ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			if !strings.Contains(rec.Body.String(), tt.wantBody) {
				t.Errorf("body = %q, want it to contain %q", rec.Body.String(), tt.wantBody)
			}
		})
	}
}
