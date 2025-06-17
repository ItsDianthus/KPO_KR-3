package repo

import (
	"database/sql"
	"github.com/ItsDianthus/shop/orders-service/internal/model"
)

func CreateOrder(tx *sql.Tx, userID string, amount float64) (int, error) {
	var id int
	err := tx.QueryRow(
		`INSERT INTO orders(user_id, amount) VALUES($1, $2) RETURNING id`,
		userID, amount,
	).Scan(&id)
	return id, err
}

func InsertOutbox(tx *sql.Tx, topic string, payload []byte) error {
	_, err := tx.Exec(
		`INSERT INTO outbox(topic, payload) VALUES($1, $2)`,
		topic, payload,
	)
	return err
}

func GetOrderByID(db *sql.DB, id int) (*model.Order, error) {
	o := &model.Order{}
	err := db.QueryRow(
		`SELECT id, user_id, amount, status, created_at FROM orders WHERE id = $1`,
		id,
	).Scan(&o.ID, &o.UserID, &o.Amount, &o.Status, &o.CreatedAt)
	if err != nil {
		return nil, err
	}
	return o, nil
}

func ListOrders(db *sql.DB) ([]model.Order, error) {
	rows, err := db.Query(
		`SELECT id, user_id, amount, status, created_at FROM orders ORDER BY id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.Order
	for rows.Next() {
		var o model.Order
		rows.Scan(&o.ID, &o.UserID, &o.Amount, &o.Status, &o.CreatedAt)
		list = append(list, o)
	}
	return list, rows.Err()
}
