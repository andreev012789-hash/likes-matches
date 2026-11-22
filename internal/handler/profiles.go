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

type createProfileRequest struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
	City string `json:"city"`
}

func (h *Handler) CreateProfile(w http.ResponseWriter, r *http.Request) {
	var req createProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.City == "" {
		writeError(w, http.StatusBadRequest, "city is required")
		return
	}
	if req.Age < 18 {
		writeError(w, http.StatusBadRequest, "age must be at least 18")
		return
	}

	p, err := h.profiles.CreateProfile(r.Context(), req.Name, req.Age, req.City)
	if err != nil {
		slog.Error("create profile", "error", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusCreated, p)
}

func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	p, err := h.profiles.GetProfile(r.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "profile not found")
		return
	}
	if err != nil {
		slog.Error("get profile", "error", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, p)
}
