package models

import "time"

type Service struct {
	ID              int        `json:"id"`
	VenueID         int        `json:"venue_id"`
	Name            string     `json:"name"`
	Description     *string    `json:"description,omitempty"`
	DurationMinutes int        `json:"duration_minutes"`
	Price           *float64   `json:"price,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}
