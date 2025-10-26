package models

import "time"

type Notification struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	Message    string    `json:"message"`
	Channel    string    `json:"channel"`   // "email" or "telegram"
	Recipient  string    `json:"recipient"` // Email или Telegram ID
	SendAt     time.Time `json:"send_at"`
	CreatedAt  time.Time `json:"created_at"`
	Status     string    `json:"status"` // "pending", "sent", "failed", "canceled"
	RetryCount int       `json:"retry_count"`
}

type CreateNotificationRequest struct {
	UserID    string    `json:"user_id" validate:"required"`
	Message   string    `json:"message" validate:"required"`
	Channel   string    `json:"channel" validate:"required,oneof=email telegram"`
	Recipient string    `json:"recipient" validate:"required"`
	SendAt    time.Time `json:"send_at" validate:"required"`
}
