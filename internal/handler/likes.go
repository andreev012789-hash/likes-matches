package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/andreev012789-hash/likes-matches/internal/event"
	"github.com/andreev012789-hash/likes-matches/internal/repository"
)

type likeRequest struct {
	FromID int64 `json:"from_id"`
	ToID   int64 `json:"to_id"`
}

func (h *Handler) createLike(w http.ResponseWriter, r *http.Request) {
	var req likeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}

	if req.FromID <= 0 || req.ToID <= 0 {
		writeError(w, http.StatusBadRequest, "invalid ids")
		return
	}

	allowed, err := h.limiter.Allow(r.Context(), req.FromID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "rate limiter unavailable")
		return
	}

	if !allowed {
		writeError(w, http.StatusTooManyRequests, "too many likes")
		return
	}

	result, err := h.likes.Like(r.Context(), req.FromID, req.ToID)
	if err != nil {
		if errors.Is(err, repository.ErrSelfLike) {
			writeError(w, http.StatusBadRequest, "cannot like yourself")
			return
		}
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "profile not found")
			return
		}
		slog.Error("create like", "error", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return

	}
	h.publish(event.New(event.TypeLikeCreated, req.FromID, req.ToID))
	if result.Matched {
		h.publish(event.New(event.TypeMatchCreated, req.FromID, req.ToID))
	}
	writeJSON(w, http.StatusOK, result)

}

func (h *Handler) getMatches(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	matches, err := h.likes.Matches(r.Context(), id)
	if err != nil {
		slog.Error("get matches", "error", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, matches)
}

func (h *Handler) publish(e event.Event) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := h.publisher.Publish(ctx, e); err != nil {
		slog.Error("publish event", "type", e.Type, "event_id", e.ID, "error", err)
	}
}
