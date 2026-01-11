package entity

import "time"

// Token is a struct which holds all information about a token
type Token struct {
	UserID    string     `json:"userId"`
	Token     string     `json:"token"`
	CreatedAt *time.Time `json:"createdAt,omitempty"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
	ExpiresAt time.Time  `json:"expiresAt"`
	RevokedAt *time.Time `json:"revokedAt,omitempty"`
}
