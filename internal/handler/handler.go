package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/andreev012789-hash/likes-matches/internal/model"
)

type ProfileStore interface {
	CreateProfile(ctx context.Context, name string, age int, city string) (model.Profile, error)
	GetProfile(ctx context.Context, id int64) (model.Profile, error)
}

type Handler struct {
	profiles ProfileStore
	likes    LikeStore
}

type LikeStore interface {
	Like(ctx context.Context, fromID, toID int64) (model.LikeResult, error)
	Matches(ctx context.Context, userID int64) ([]model.Profile, error)
}

func New(profiles ProfileStore, likes LikeStore) *Handler {
	return &Handler{profiles: profiles, likes: likes}
}

func (h *Handler) Router() http.Handler {
	r := chi.NewRouter()
	r.Get("/health", health)
	r.Post("/profiles", h.CreateProfile)
	r.Get("/profiles/{id}", h.GetProfile)
	r.Post("/likes", h.createLike)
	r.Get("/profiles/{id}/matches", h.getMatches)
	return r
}

func health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("write response", "error", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
