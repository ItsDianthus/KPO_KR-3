package repo

import (
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestCreateAccount(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(
		`INSERT INTO accounts(user_id) VALUES($1)`,
	)).
		WithArgs("alice").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = CreateAccount(db, "alice")
	if err != nil {
		t.Errorf("CreateAccount error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestGetBalance(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT balance FROM accounts WHERE user_id = $1`,
	)).
		WithArgs("bob").
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(123.45))

	bal, err := GetBalance(db, "bob")
	if err != nil {
		t.Errorf("GetBalance error: %v", err)
	}
	if bal != 123.45 {
		t.Errorf("expected balance=123.45, got %f", bal)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestTopUp(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE accounts SET balance = balance + $1 WHERE user_id = $2`,
	)).
		WithArgs(50.0, "eve").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = TopUp(db, "eve", 50.0)
	if err != nil {
		t.Errorf("TopUp error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestInsertInbox(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(
		`INSERT INTO inbox(message_id, payload)
        VALUES($1,$2)
        ON CONFLICT (message_id) DO NOTHING`,
	)).
		WithArgs("msg-1", []byte(`{"foo":"bar"}`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Begin: %v", err)
	}

	err = InsertInbox(tx, "msg-1", []byte(`{"foo":"bar"}`))
	if err != nil {
		t.Errorf("InsertInbox error: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestMarkInboxProcessed(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE inbox SET processed = true WHERE message_id = $1`,
	)).
		WithArgs("msg-1").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Begin: %v", err)
	}

	err = MarkInboxProcessed(tx, "msg-1")
	if err != nil {
		t.Errorf("MarkInboxProcessed error: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestInsertOutbox(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(
		`INSERT INTO outbox(topic, payload) VALUES($1, $2)`,
	)).
		WithArgs("payments.processed", []byte(`{"ok":true}`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Begin: %v", err)
	}

	err = InsertOutbox(tx, "payments.processed", []byte(`{"ok":true}`))
	if err != nil {
		t.Errorf("InsertOutbox error: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestListUnprocessedOutbox(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "topic", "payload"}).
		AddRow(1, "t1", []byte(`{"a":1}`)).
		AddRow(2, "t2", []byte(`{"b":2}`))

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT id, topic, payload FROM outbox WHERE processed = false`,
	)).WillReturnRows(rows)

	sqlRows, err := ListUnprocessedOutbox(db)
	if err != nil {
		t.Fatalf("ListUnprocessedOutbox error: %v", err)
	}
	defer sqlRows.Close()

	var got []struct {
		id      int
		topic   string
		payload []byte
	}
	for sqlRows.Next() {
		var id int
		var topic string
		var payload []byte
		if err := sqlRows.Scan(&id, &topic, &payload); err != nil {
			t.Errorf("scan error: %v", err)
		}
		got = append(got, struct {
			id      int
			topic   string
			payload []byte
		}{id, topic, payload})
	}
	if len(got) != 2 {
		t.Errorf("expected 2 rows, got %d", len(got))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestMarkOutboxProcessed(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE outbox SET processed = true WHERE id = $1`,
	)).
		WithArgs(42).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = MarkOutboxProcessed(db, 42)
	if err != nil {
		t.Errorf("MarkOutboxProcessed error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
