package handler

import (
	"database/sql"
	"encoding/json"
	"github.com/ItsDianthus/shop/payments-service/internal/model"
	"github.com/ItsDianthus/shop/payments-service/internal/repo"
	"net/http"
	"strconv"
	"strings"
)

// POST /accounts?user_id=alice
func CreateAccountHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		user := r.URL.Query().Get("user_id")
		if user == "" {
			http.Error(w, "Missing user_id", http.StatusBadRequest)
			return
		}
		if err := repo.CreateAccount(db, user); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}

// GET  /accounts/{user_id}/balance
// POST /accounts/{user_id}/topup?amount=123.45
func AccountHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// /accounts/{rest...}
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) < 2 || parts[0] != "accounts" {
			http.NotFound(w, r)
			return
		}
		userID := parts[1]

		switch {
		case r.Method == http.MethodGet && len(parts) == 3 && parts[2] == "balance":
			balance, err := repo.GetBalance(db, userID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(model.Account{UserID: userID, Balance: balance})

		case r.Method == http.MethodPost && len(parts) == 3 && parts[2] == "topup":
			amtStr := r.URL.Query().Get("amount")
			amt, err := strconv.ParseFloat(amtStr, 64)
			if err != nil {
				http.Error(w, "Invalid amount", http.StatusBadRequest)
				return
			}
			if err := repo.TopUp(db, userID, amt); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)

		default:
			http.NotFound(w, r)
		}
	}
}
