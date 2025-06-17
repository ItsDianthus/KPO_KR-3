package model

import "time"

// Account представляет баланс пользователя
type Account struct {
	UserID  string  `json:"user_id"`
	Balance float64 `json:"balance"`
}

// PaymentProcessedEvent для топика "payments.processed"
type PaymentProcessedEvent struct {
	OrderID     int       `json:"order_id"`
	UserID      string    `json:"user_id"`
	Amount      float64   `json:"amount"`
	Status      string    `json:"status"` // "success" или "failed"
	ProcessedAt time.Time `json:"processed_at"`
}

// OrderCreatedEvent нужен для десериализации входящих сообщений
type OrderCreatedEvent struct {
	OrderID   int       `json:"order_id"`
	UserID    string    `json:"user_id"`
	Amount    float64   `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
}
