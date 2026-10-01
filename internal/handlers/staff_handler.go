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

type StaffHandler struct {
	svc service.StaffService
}

func NewStaffHandler(svc service.StaffService) *StaffHandler {
	return &StaffHandler{
		svc: svc,
	}
}

func (h *StaffHandler) GetQueues(w http.ResponseWriter, r *http.Request) {
	ownerID := r.Context().Value(middleware.UserIDKey).(int)

	var venueID, serviceID *int
	var status *string

	if v := r.URL.Query().Get("venue_id"); v != "" {
		if id, err := strconv.Atoi(v); err == nil {
			venueID = &id
		}
	}
	if v := r.URL.Query().Get("service_id"); v != "" {
		if id, err := strconv.Atoi(v); err == nil {
			serviceID = &id
		}
	}
	if v := r.URL.Query().Get("status"); v != "" {
		status = &v
	}

	tickets, err := h.svc.GetStaffQueues(r.Context(), ownerID, venueID, serviceID, status)
	if err != nil {
		http.Error(w, `{"error": "failed to fetch queues"}`, http.StatusInternalServerError)
		return
	}

	if tickets == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tickets)
}

func (h *StaffHandler) CallNext(w http.ResponseWriter, r *http.Request) {
	ownerID := r.Context().Value(middleware.UserIDKey).(int)
	queueIDStr := r.PathValue("queue_id")
	queueID, err := strconv.Atoi(queueIDStr)
	if err != nil {
		http.Error(w, `{"error": "invalid queue id format"}`, http.StatusBadRequest)
		return
	}

	ticket, err := h.svc.CallNext(r.Context(), ownerID, queueID)
	if err != nil {
		if errors.Is(err, repository.ErrQueueTicketNotFound) {
			http.Error(w, `{"error": "no waiting tickets found or unauthorized"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error": "failed to call next ticket"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ticket)
}

func (h *StaffHandler) ServeTicket(w http.ResponseWriter, r *http.Request) {
	ownerID := r.Context().Value(middleware.UserIDKey).(int)
	ticketIDStr := r.PathValue("ticket_id")
	ticketID, err := strconv.Atoi(ticketIDStr)
	if err != nil {
		http.Error(w, `{"error": "invalid ticket id format"}`, http.StatusBadRequest)
		return
	}

	ticket, err := h.svc.Serve(r.Context(), ownerID, ticketID)
	if err != nil {
		if errors.Is(err, repository.ErrQueueTicketNotFound) {
			http.Error(w, `{"error": "ticket not found, unauthorized, or invalid state transition"}`, http.StatusConflict)
			return
		}
		http.Error(w, `{"error": "failed to serve ticket"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ticket)
}

func (h *StaffHandler) NoShowTicket(w http.ResponseWriter, r *http.Request) {
	ownerID := r.Context().Value(middleware.UserIDKey).(int)
	ticketIDStr := r.PathValue("ticket_id")
	ticketID, err := strconv.Atoi(ticketIDStr)
	if err != nil {
		http.Error(w, `{"error": "invalid ticket id format"}`, http.StatusBadRequest)
		return
	}

	ticket, err := h.svc.NoShow(r.Context(), ownerID, ticketID)
	if err != nil {
		if errors.Is(err, repository.ErrQueueTicketNotFound) {
			http.Error(w, `{"error": "ticket not found, unauthorized, or invalid state transition"}`, http.StatusConflict)
			return
		}
		http.Error(w, `{"error": "failed to mark ticket as no-show"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ticket)
}

func (h *StaffHandler) CancelTicket(w http.ResponseWriter, r *http.Request) {
	ownerID := r.Context().Value(middleware.UserIDKey).(int)
	ticketIDStr := r.PathValue("ticket_id")
	ticketID, err := strconv.Atoi(ticketIDStr)
	if err != nil {
		http.Error(w, `{"error": "invalid ticket id format"}`, http.StatusBadRequest)
		return
	}

	ticket, err := h.svc.Cancel(r.Context(), ownerID, ticketID)
	if err != nil {
		if errors.Is(err, repository.ErrQueueTicketNotFound) {
			http.Error(w, `{"error": "ticket not found, unauthorized, or invalid state transition"}`, http.StatusConflict)
			return
		}
		http.Error(w, `{"error": "failed to cancel ticket"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ticket)
}
