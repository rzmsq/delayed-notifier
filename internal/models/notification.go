package models

import "time"

type NotificationStatus string

const (
	StatusPending   NotificationStatus = "PENDING"
	StatusSent      NotificationStatus = "SENT"
	StatusFailed    NotificationStatus = "FAILED"
	StatusCancelled NotificationStatus = "CANCELLED"
)

type Notification struct {
	ID        string             `json:"id"`
	Channel   string             `json:"channel"`
	Status    NotificationStatus `json:"status"`
	SendAt    time.Time          `json:"send_at"`
	Recipient string             `json:"recipient"`
	Message   string             `json:"message" validate:"required"`
}
