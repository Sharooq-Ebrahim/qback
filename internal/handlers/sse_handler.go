package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"qback/internal/middleware"
	"qback/internal/sse"
)


type SSEHandler struct {
	broker *sse.Broker
}


func NewSSEHandler(broker *sse.Broker) *SSEHandler {
	return &SSEHandler{broker: broker}
}


func (h *SSEHandler) StreamQueueEvents(w http.ResponseWriter, r *http.Request) {

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	ownerID := r.Context().Value(middleware.UserIDKey).(int)

	var serviceID *int
	if v := r.URL.Query().Get("service_id"); v != "" {
		if id, err := strconv.Atoi(v); err == nil {
			serviceID = &id
		}
	}


	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch, unsub := h.broker.Subscribe(ownerID, serviceID)
	defer unsub()


	fmt.Fprintf(w, "event: connected\ndata: {}\n\n")
	flusher.Flush()

	for {
		select {
		case <-r.Context().Done():
			// Client disconnected 
			return
		case event, ok := <-ch:
			if !ok {
				return
			}
			data, err := json.Marshal(event)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, data)
			flusher.Flush()
		}
	}
}
