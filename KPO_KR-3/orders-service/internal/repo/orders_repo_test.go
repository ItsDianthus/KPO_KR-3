// orders-service/internal/repo/orders_repo_test.go
package repo

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestCreateOrder(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error initializing mock: %v", err)
	}
	defer db.Close()

	// Ожидаем Begin → QueryRow → Commit
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(
		`INSERT INTO orders(user_id, amount) VALUES($1, $2) RETURNING id`,
	)).
		WithArgs("alice", 42.5).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(7))
	mock.ExpectCommit()

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}

	id, err := CreateOrder(tx, "alice", 42.5)
	if err != nil {
		t.Fatalf("CreateOrder error: %v", err)
	}
	if id != 7 {
		t.Errorf("expected id=7, got %d", id)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("commit tx: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestInsertOutbox(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error initializing mock: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(
		`INSERT INTO outbox(topic, payload) VALUES($1, $2)`,
	)).
		WithArgs("orders.created", []byte(`{"foo":"bar"}`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}

	err = InsertOutbox(tx, "orders.created", []byte(`{"foo":"bar"}`))
	if err != nil {
		t.Fatalf("InsertOutbox error: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("commit tx: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestGetOrderByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error initializing mock: %v", err)
	}
	defer db.Close()

	// Подготовим фиктивное время для сравнения
	now := time.Now().Truncate(time.Second)

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT id, user_id, amount, status, created_at FROM orders WHERE id = $1`,
	)).
		WithArgs(5).
		WillReturnRows(sqlmock.NewRows(
			[]string{"id", "user_id", "amount", "status", "created_at"},
		).AddRow(5, "bob", 99.9, "pending", now))

	order, err := GetOrderByID(db, 5)
	if err != nil {
		t.Fatalf("GetOrderByID error: %v", err)
	}
	if order.ID != 5 || order.UserID != "bob" || order.Amount != 99.9 || order.Status != "pending" {
		t.Errorf("unexpected order: %+v", order)
	}
	if !order.CreatedAt.Equal(now) {
		t.Errorf("expected CreatedAt=%v, got %v", now, order.CreatedAt)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestListOrders(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error initializing mock: %v", err)
	}
	defer db.Close()

	now1 := time.Now().Add(-time.Hour).Truncate(time.Second)
	now2 := time.Now().Truncate(time.Second)

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT id, user_id, amount, status, created_at FROM orders ORDER BY id`,
	)).
		WillReturnRows(sqlmock.NewRows(
			[]string{"id", "user_id", "amount", "status", "created_at"},
		).
			AddRow(1, "alice", 10.0, "pending", now1).
			AddRow(2, "eve", 20.5, "shipped", now2),
		)

	orders, err := ListOrders(db)
	if err != nil {
		t.Fatalf("ListOrders error: %v", err)
	}
	if len(orders) != 2 {
		t.Fatalf("expected 2 orders, got %d", len(orders))
	}

	if orders[0].ID != 1 || orders[0].UserID != "alice" || orders[0].Amount != 10.0 || orders[0].Status != "pending" {
		t.Errorf("order[0] mismatch: %+v", orders[0])
	}
	if orders[1].ID != 2 || orders[1].UserID != "eve" || orders[1].Amount != 20.5 || orders[1].Status != "shipped" {
		t.Errorf("order[1] mismatch: %+v", orders[1])
	}
	if !orders[0].CreatedAt.Equal(now1) || !orders[1].CreatedAt.Equal(now2) {
		t.Errorf("CreatedAt timestamps mismatch")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
