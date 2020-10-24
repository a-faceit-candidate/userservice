package model

import "time"

type User struct {
	ID           string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	FirstName    string
	LastName     string
	Name         string
	Email        string
	Password     string
	PasswordHash string
	PasswordSalt string
	Country      string
}
