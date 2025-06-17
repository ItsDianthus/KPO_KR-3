package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/ItsDianthus/shop/orders-service/internal/model"
	"github.com/ItsDianthus/shop/orders-service/internal/repo"
	"net/http"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

// CreateOrderHandler -- POST /orders?user_id=...&amount=...
func CreateOrderHandler(db *sql.DB, writer *kafka.Writer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		userID := r.URL.Query().Get("user_id")
		amountStr := r.URL.Query().Get("amount")
		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			http.Error(w, "Invalid amount", http.StatusBadRequest)
			return
		}

		tx, err := db.Begin()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// 1) создаём заказ
		orderID, err := repo.CreateOrder(tx, userID, amount)
		if err != nil {
			tx.Rollback()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// 2) пишем событие в outbox
		evt := model.OrderCreatedEvent{
			OrderID:   orderID,
			UserID:    userID,
			Amount:    amount,
			CreatedAt: time.Now(),
		}
		payload, err := json.Marshal(evt)
		if err != nil {
			tx.Rollback()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := repo.InsertOutbox(tx, "orders.created", payload); err != nil {
			tx.Rollback()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tx.Commit()
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "%d", orderID)
	}
}

// GetOrderOrListHandler -- GET /orders and GET /orders/{id}
func GetOrderOrListHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/orders" || path == "/orders/" {
			// список
			orders, err := repo.ListOrders(db)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(orders)
		} else {
			// конкретный заказ
			idStr := path[len("/orders/"):]
			id, err := strconv.Atoi(idStr)
			if err != nil {
				http.Error(w, "Invalid ID", http.StatusBadRequest)
				return
			}
			order, err := repo.GetOrderByID(db, id)
			if err == sql.ErrNoRows {
				http.NotFound(w, r)
				return
			} else if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(order)
		}
	}
}
