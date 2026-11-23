package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/andreev012789-hash/likes-matches/internal/model"
	"github.com/andreev012789-hash/likes-matches/internal/repository"
)

type fakeLikeStore struct {
	likeFn    func(ctx context.Context, fromID, toID int64) (model.LikeResult, error)
	matchesFn func(ctx context.Context, userID int64) ([]model.Profile, error)
}

type fakeLimiter struct {
	allow bool
	err   error
}

func (f *fakeLikeStore) Like(ctx context.Context, fromID, toID int64) (model.LikeResult, error) {
	return f.likeFn(ctx, fromID, toID)
}

func (f *fakeLikeStore) Matches(ctx context.Context, userID int64) ([]model.Profile, error) {
	return f.matchesFn(ctx, userID)
}

func (f fakeLimiter) Allow(ctx context.Context, userID int64) (bool, error) {
	return f.allow, f.err
}

func TestCreateLike(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		likeResult model.LikeResult
		likeErr    error
		wantStatus int
		wantBody   string
		blocked    bool
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
		{
			name:       "rate limit exceeded",
			body:       `{"from_id":1,"to_id":2}`,
			blocked:    true,
			wantStatus: http.StatusTooManyRequests,
			wantBody:   `"too many likes"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeLikeStore{
				likeFn: func(ctx context.Context, fromID, toID int64) (model.LikeResult, error) {
					return tt.likeResult, tt.likeErr
				},
			}

			h := New(nil, store, fakeLimiter{allow: !tt.blocked})

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

func TestGetMatches(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		matches    []model.Profile
		matchesErr error
		wantStatus int
		wantBody   string
	}{
		{
			name:       "has matches",
			path:       "/profiles/1/matches",
			matches:    []model.Profile{{ID: 2, Name: "Boris"}},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"Boris"`,
		},
		{
			name:       "no matches",
			path:       "/profiles/1/matches",
			matches:    []model.Profile{},
			wantStatus: http.StatusOK,
			wantBody:   `[]`,
		},
		{
			name:       "invalid id",
			path:       "/profiles/abc/matches",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "store error",
			path:       "/profiles/1/matches",
			matchesErr: errors.New("db is down"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeLikeStore{
				matchesFn: func(ctx context.Context, userID int64) ([]model.Profile, error) {
					return tt.matches, tt.matchesErr
				},
			}

			h := New(nil, store, fakeLimiter{allow: true})

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
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
