package models

import "time"

type UserSessions struct {
	ID           int       `gorm:"AUTO_INCREMENT" json:"id"`
	RefreshToken string    `json:"refresh_token"`
	UserAgent    string    `json:"user_agent"`
	Revoked      uint      `json:"revoked"`
	UserID       uint      `json:"user_id"`
	ExpiredAt    int64     `json:"expired_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
