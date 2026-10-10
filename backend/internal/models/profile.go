package models

import "time"

// Profile groups skills and saved jobs for application tracking.
type Profile struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Skills    []string  `json:"skills"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
