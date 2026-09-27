package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"qback/internal/middleware"
	"qback/internal/repository"
	"qback/internal/service"
)

type TokenHandler struct {
	svc service.QueueService
}

func NewTokenHandler(svc service.QueueService) *TokenHandler {
	return &TokenHandler{
		svc: svc,
	}
}

type JoinQueueRequest struct {
	VenueID   int `json:"venue_id"`
	ServiceID int `json:"service_id"`
}

func (h *TokenHandler) JoinQueue(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req JoinQueueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.VenueID <= 0 || req.ServiceID <= 0 {
		http.Error(w, `{"error": "venue_id and service_id are required"}`, http.StatusBadRequest)
		return
	}

	ticket, err := h.svc.JoinQueue(r.Context(), userID, req.VenueID, req.ServiceID)
	if err != nil {
		if errors.Is(err, service.ErrServiceVenueMismatch) || errors.Is(err, repository.ErrServiceNotFound) {
			http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}
		http.Error(w, `{"error": "failed to join queue"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ticket)
}

func (h *TokenHandler) GetActive(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	tickets, err := h.svc.GetActiveTickets(r.Context(), userID)
	if err != nil {
		http.Error(w, `{"error": "failed to retrieve active tickets"}`, http.StatusInternalServerError)
		return
	}

	if tickets == nil {
		tickets = make([]*service.TicketDetails, 0)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tickets)
}

func (h *TokenHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	tickets, err := h.svc.GetTicketHistory(r.Context(), userID)
	if err != nil {
		http.Error(w, `{"error": "failed to retrieve ticket history"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if tickets == nil {
		w.Write([]byte("[]"))
		return
	}
	json.NewEncoder(w).Encode(tickets)
}

func (h *TokenHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	idStr := r.PathValue("id")
	ticketID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error": "invalid ticket id format"}`, http.StatusBadRequest)
		return
	}

	ticket, err := h.svc.GetTicketDetails(r.Context(), ticketID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrQueueTicketNotFound) {
			http.Error(w, `{"error": "ticket not found"}`, http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrUnauthorizedTicket) {
			http.Error(w, `{"error": "unauthorized access to ticket"}`, http.StatusForbidden)
			return
		}
		http.Error(w, `{"error": "failed to retrieve ticket details"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ticket)
}

func (h *TokenHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	idStr := r.PathValue("id")
	ticketID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error": "invalid ticket id format"}`, http.StatusBadRequest)
		return
	}

	err = h.svc.CancelTicket(r.Context(), ticketID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrQueueTicketNotFound) {
			http.Error(w, `{"error": "ticket not found"}`, http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrUnauthorizedTicket) {
			http.Error(w, `{"error": "unauthorized access to ticket"}`, http.StatusForbidden)
			return
		}
		if errors.Is(err, service.ErrTicketNotWaiting) {
			http.Error(w, `{"error": "ticket is not in waiting state"}`, http.StatusBadRequest)
			return
		}
		http.Error(w, `{"error": "failed to cancel ticket"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "ticket cancelled successfully"}`))
}
