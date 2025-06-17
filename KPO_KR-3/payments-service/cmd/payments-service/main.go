package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/segmentio/kafka-go"

	"github.com/ItsDianthus/shop/payments-service/internal/handler"
	"github.com/ItsDianthus/shop/payments-service/internal/worker"
)

func waitForKafka(broker, topic string) {
	for {
		conn, err := kafka.DialLeader(context.Background(), "tcp", broker, topic, 0)
		if err != nil {
			log.Printf("waiting for Kafka at %s...: %v", broker, err)
			time.Sleep(2 * time.Second)
			continue
		}
		conn.Close()
		log.Printf("connected to Kafka broker at %s (topic %s)", broker, topic)
		return
	}
}

func main() {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("db open:", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatal("db ping:", err)
	}
	log.Println("Connected to payments DB")

	if _, err := db.Exec(`
CREATE TABLE IF NOT EXISTS accounts (
  user_id TEXT PRIMARY KEY,
  balance NUMERIC NOT NULL DEFAULT 0
);
CREATE TABLE IF NOT EXISTS inbox (
  message_id TEXT PRIMARY KEY,
  payload JSONB NOT NULL,
  processed BOOLEAN NOT NULL DEFAULT FALSE,
  received_at TIMESTAMPTZ DEFAULT now()
);
CREATE TABLE IF NOT EXISTS outbox (
  id SERIAL PRIMARY KEY,
  topic TEXT NOT NULL,
  payload JSONB NOT NULL,
  processed BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ DEFAULT now()
);
`); err != nil {
		log.Fatal("migrations:", err)
	}
	log.Println("Migrations applied")

	broker := os.Getenv("KAFKA_BROKER")
	ordersTopic := os.Getenv("ORDERS_TOPIC")
	paymentsTopic := os.Getenv("PAYMENTS_TOPIC")

	waitForKafka(broker, ordersTopic)

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{broker},
		GroupID:     "payments-service",
		GroupTopics: []string{ordersTopic},
	})
	defer reader.Close()

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{broker},
		Topic:   paymentsTopic,
	})
	defer writer.Close()

	go worker.RunInboxOutboxWorker(db, reader, writer)

	mux := http.NewServeMux()
	mux.HandleFunc("/accounts", handler.CreateAccountHandler(db))
	mux.HandleFunc("/accounts/", handler.AccountHandler(db))

	addr := ":8082"
	log.Println("Payments service listening on", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
