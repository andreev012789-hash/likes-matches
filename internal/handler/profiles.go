package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
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
	if req.Name == "" || req.City == "" || req.Age < 18 {
		writeError(w, http.StatusBadRequest, "name and city are required, age must be at least 18")
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
