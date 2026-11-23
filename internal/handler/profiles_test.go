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

type fakeProfileStore struct {
	createFn func(ctx context.Context, name string, age int, city string) (model.Profile, error)
	getFn    func(ctx context.Context, id int64) (model.Profile, error)
}

func (f *fakeProfileStore) CreateProfile(ctx context.Context, name string, age int, city string) (model.Profile, error) {
	return f.createFn(ctx, name, age, city)
}

func (f *fakeProfileStore) GetProfile(ctx context.Context, id int64) (model.Profile, error) {
	return f.getFn(ctx, id)
}

func TestGetProfile(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		profile    model.Profile
		getErr     error
		wantStatus int
		wantBody   string
	}{
		{
			name:       "found",
			path:       "/profiles/1",
			profile:    model.Profile{ID: 1, Name: "Anna"},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"Anna"`,
		},
		{
			name:       "not found",
			path:       "/profiles/999",
			getErr:     repository.ErrNotFound,
			wantStatus: http.StatusNotFound,
			wantBody:   "profile not found",
		},
		{
			name:       "invalid id",
			path:       "/profiles/abc",
			wantStatus: http.StatusBadRequest,
			wantBody:   "invalid id",
		},
		{
			name:       "store error",
			path:       "/profiles/1",
			getErr:     errors.New("db is down"),
			wantStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeProfileStore{
				getFn: func(ctx context.Context, id int64) (model.Profile, error) {
					return tt.profile, tt.getErr
				},
			}

			h := New(store, nil)

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

func TestCreateProfile(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		profile    model.Profile
		createErr  error
		wantStatus int
		wantBody   string
	}{
		{
			name:       "created",
			body:       `{"name":"Anna","age":25,"city":"Moscow"}`,
			profile:    model.Profile{ID: 1, Name: "Anna"},
			wantStatus: http.StatusCreated,
			wantBody:   `"name":"Anna"`,
		},
		{
			name:       "invalid json",
			body:       `{bad`,
			wantStatus: http.StatusBadRequest,
			wantBody:   "invalid json",
		},
		{
			name:       "empty name",
			body:       `{"name":"","age":25,"city":"Moscow"}`,
			wantStatus: http.StatusBadRequest,
			wantBody:   "name is required",
		},
		{
			name:       "empty city",
			body:       `{"name":"Anna","age":25,"city":""}`,
			wantStatus: http.StatusBadRequest,
			wantBody:   "city is required",
		},
		{
			name:       "age under 18",
			body:       `{"name":"Anna","age":17,"city":"Moscow"}`,
			wantStatus: http.StatusBadRequest,
			wantBody:   "age must be at least 18",
		},
		{
			name:       "store error",
			body:       `{"name":"Anna","age":25,"city":"Moscow"}`,
			createErr:  errors.New("db is down"),
			wantStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeProfileStore{
				createFn: func(ctx context.Context, name string, age int, city string) (model.Profile, error) {
					return tt.profile, tt.createErr
				},
			}

			h := New(store, nil)

			req := httptest.NewRequest(http.MethodPost, "/profiles", strings.NewReader(tt.body))
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
