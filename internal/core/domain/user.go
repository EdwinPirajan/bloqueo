package domain

import "time"

type User struct {
	ID             int       `json:"id,omitempty"`
	Name           string    `json:"name"`
	Client         string    `json:"client,omitempty"`
	Active         bool      `json:"active"`
	LastConnection time.Time `json:"last_connection"`
}
