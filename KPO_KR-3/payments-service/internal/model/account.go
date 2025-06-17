package model

import "time"

type Account struct {
	UserID  string  `json:"user_id"`
	Balance float64 `json:"balance"`
}

type PaymentProcessedEvent struct {
	OrderID     int       `json:"order_id"`
	UserID      string    `json:"user_id"`
	Amount      float64   `json:"amount"`
	Status      string    `json:"status"`
	ProcessedAt time.Time `json:"processed_at"`
}

type OrderCreatedEvent struct {
	OrderID   int       `json:"order_id"`
	UserID    string    `json:"user_id"`
	Amount    float64   `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
}
