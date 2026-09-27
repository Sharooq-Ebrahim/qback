package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"qback/internal/repository"
	"qback/internal/service"
)

type VenueHandler struct {
	svc service.VenueService
}

func NewVenueHandler(svc service.VenueService) *VenueHandler {
	return &VenueHandler{
		svc: svc,
	}
}

func (h *VenueHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")

	var result interface{}
	var err error

	if search != "" {
		result, err = h.svc.Search(r.Context(), search)
	} else {
		result, err = h.svc.GetAll(r.Context())
	}

	if err != nil {
		if errors.Is(err, service.ErrEmptySearchQuery) {
			http.Error(w, `{"error": "search query cannot be empty"}`, http.StatusBadRequest)
			return
		}
		http.Error(w, `{"error": "failed to retrieve venues"}`, http.StatusInternalServerError)
		return
	}

	if result == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *VenueHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error": "invalid venue id format"}`, http.StatusBadRequest)
		return
	}

	venueObj, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrInvalidVenueID) {
			http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}
		if errors.Is(err, repository.ErrVenueNotFound) {
			http.Error(w, `{"error": "venue not found"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error": "failed to retrieve venue"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(venueObj)
}
