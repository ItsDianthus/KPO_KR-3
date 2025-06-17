package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/ItsDianthus/shop/payments-service/internal/model"
	"github.com/ItsDianthus/shop/payments-service/internal/repo"

	"github.com/segmentio/kafka-go"
)

func RunInboxOutboxWorker(db *sql.DB, reader *kafka.Reader, writer *kafka.Writer) {
	go func() {
		for {
			msg, err := reader.ReadMessage(context.Background())
			if err != nil {
				log.Println("kafka read:", err)
				time.Sleep(time.Second)
				continue
			}
			messageID := msg.Topic + "-" + strconv.FormatInt(msg.Offset, 10)

			tx, err := db.Begin()
			if err != nil {
				log.Println("tx begin:", err)
				continue
			}
			if err := repo.InsertInbox(tx, messageID, msg.Value); err != nil {
				log.Println("insert inbox:", err)
				tx.Rollback()
				continue
			}
			var processed bool
			tx.QueryRow(`SELECT processed FROM inbox WHERE message_id=$1`, messageID).Scan(&processed)
			if processed {
				tx.Rollback()
				continue
			}
			var evt model.OrderCreatedEvent
			json.Unmarshal(msg.Value, &evt)
			status := "success"
			if _, err := tx.Exec(
				`UPDATE accounts SET balance = balance - $1 WHERE user_id = $2`,
				evt.Amount, evt.UserID,
			); err != nil {
				status = "failed"
			}
			payEvt := model.PaymentProcessedEvent{
				OrderID:     evt.OrderID,
				UserID:      evt.UserID,
				Amount:      evt.Amount,
				Status:      status,
				ProcessedAt: time.Now(),
			}
			data, _ := json.Marshal(payEvt)
			repo.InsertOutbox(tx, "payments.processed", data)

			repo.MarkInboxProcessed(tx, messageID)
			tx.Commit()
		}
	}()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		rows, err := repo.ListUnprocessedOutbox(db)
		if err != nil {
			log.Println("outbox list:", err)
			continue
		}
		for rows.Next() {
			var id int
			var topic string
			var payload []byte
			rows.Scan(&id, &topic, &payload)
			if err := writer.WriteMessages(context.Background(),
				kafka.Message{Value: payload},
			); err != nil {
				log.Println("kafka write:", err)
				continue
			}
			repo.MarkOutboxProcessed(db, id)
		}
		rows.Close()
	}
}
