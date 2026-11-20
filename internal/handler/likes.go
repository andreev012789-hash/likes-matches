package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

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
		writeError(w, http.StatusInternalServerError, "internal error")
		return

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
