package model

import "time"

// Order обозначает запись в таблице orders
type Order struct {
	ID        int       `json:"id"`
	UserID    string    `json:"user_id"`
	Amount    float64   `json:"amount"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// OrderCreatedEvent — полезная нагрузка для топика "orders.created"
type OrderCreatedEvent struct {
	OrderID   int       `json:"order_id"`
	UserID    string    `json:"user_id"`
	Amount    float64   `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
}
