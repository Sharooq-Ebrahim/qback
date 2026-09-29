package models

import "time"

type Role string

const (
	RoleUser  Role = "user"
	RoleStaff Role = "staff"
	RoleAdmin Role = "admin"
)

type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	Role      Role      `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
