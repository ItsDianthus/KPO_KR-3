package repo

import "database/sql"

// CREATE TABLE accounts...
func CreateAccount(db *sql.DB, userID string) error {
	_, err := db.Exec(`INSERT INTO accounts(user_id) VALUES($1)`, userID)
	return err
}

func GetBalance(db *sql.DB, userID string) (float64, error) {
	var bal float64
	err := db.QueryRow(`SELECT balance FROM accounts WHERE user_id = $1`, userID).Scan(&bal)
	return bal, err
}

func TopUp(db *sql.DB, userID string, amount float64) error {
	_, err := db.Exec(`UPDATE accounts SET balance = balance + $1 WHERE user_id = $2`, amount, userID)
	return err
}

// Inbox operations
func InsertInbox(tx *sql.Tx, messageID string, payload []byte) error {
	_, err := tx.Exec(`
        INSERT INTO inbox(message_id, payload)
        VALUES($1,$2)
        ON CONFLICT (message_id) DO NOTHING
    `, messageID, payload)
	return err
}

func MarkInboxProcessed(tx *sql.Tx, messageID string) error {
	_, err := tx.Exec(`UPDATE inbox SET processed = true WHERE message_id = $1`, messageID)
	return err
}

// Outbox operations
func InsertOutbox(tx *sql.Tx, topic string, payload []byte) error {
	_, err := tx.Exec(`INSERT INTO outbox(topic, payload) VALUES($1, $2)`, topic, payload)
	return err
}

func ListUnprocessedOutbox(db *sql.DB) (*sql.Rows, error) {
	return db.Query(`SELECT id, topic, payload FROM outbox WHERE processed = false`)
}

func MarkOutboxProcessed(db *sql.DB, id int) error {
	_, err := db.Exec(`UPDATE outbox SET processed = true WHERE id = $1`, id)
	return err
}
