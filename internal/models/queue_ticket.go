package models

import "time"

type QueueTicket struct {
	ID           int        `json:"id"`
	UserID       int        `json:"user_id"`
	VenueID      int        `json:"venue_id"`
	ServiceID    int        `json:"service_id"`
	TicketNumber string     `json:"ticket_number"`
	Status       string     `json:"status"`
	JoinedAt     time.Time  `json:"joined_at"`
	ServedAt     *time.Time `json:"served_at,omitempty"`
	CancelledAt  *time.Time `json:"cancelled_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}
