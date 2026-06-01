package domain

import "github.com/google/uuid"

type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	Title        string    `json:"title"`
	RefreshToken string    `json:"refresh_token"`
	AccessToken  string    `json:"access_token"`
	Role         string    `json:"role"`
}

type OnlyUser struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
	Title string    `json:"title"`
	Role  string    `json:"role"`
}
