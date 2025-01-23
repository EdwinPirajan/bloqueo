package domain

import "time"

type User struct {
	ID             int
	Name           string
	Active         bool
	LastConnection time.Time
}
