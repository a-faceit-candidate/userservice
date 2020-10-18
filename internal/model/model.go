package model

import "time"

type User struct {
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time
	Name      string
	Email     string
	Country   string
}
