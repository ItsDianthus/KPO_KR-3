package worker

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

// RunOutboxWorker читает неотправленные записи из outbox и отправляет их в Kafka
func RunOutboxWorker(db *sql.DB, writer *kafka.Writer) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		rows, err := db.Query(`SELECT id, topic, payload FROM outbox WHERE processed = false`)
		if err != nil {
			log.Println("outbox worker query:", err)
			continue
		}

		for rows.Next() {
			var id int
			var topic string
			var payload []byte
			if err := rows.Scan(&id, &topic, &payload); err != nil {
				log.Println("outbox worker scan:", err)
				continue
			}

			msg := kafka.Message{
				Value: payload,
			}
			if err := writer.WriteMessages(context.Background(), msg); err != nil {
				log.Println("outbox worker write:", err)
				continue
			}

			if _, err := db.Exec(`UPDATE outbox SET processed = true WHERE id = $1`, id); err != nil {
				log.Println("outbox worker update:", err)
			}
		}
		rows.Close()
	}
}
