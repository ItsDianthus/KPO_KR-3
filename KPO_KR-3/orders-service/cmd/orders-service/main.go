package main

import (
	"database/sql"
	"fmt"
	"github.com/ItsDianthus/shop/orders-service/internal/handler"
	_ "github.com/lib/pq"
	"github.com/segmentio/kafka-go"
	"log"
	"net/http"
	"os"

	"github.com/ItsDianthus/shop/orders-service/internal/worker"
)

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
	log.Println("Connected to orders DB")

	if _, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS orders (
            id SERIAL PRIMARY KEY,
            user_id TEXT NOT NULL,
            amount NUMERIC NOT NULL,
            status TEXT NOT NULL DEFAULT 'pending',
            created_at TIMESTAMPTZ DEFAULT now()
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
	topic := os.Getenv("ORDERS_TOPIC")
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{broker},
		Topic:   topic,
	})
	defer writer.Close()

	go worker.RunOutboxWorker(db, writer)

	mux := http.NewServeMux()
	mux.HandleFunc("/orders", handler.CreateOrderHandler(db, writer))
	mux.HandleFunc("/orders/", handler.GetOrderOrListHandler(db))

	addr := ":8081"
	log.Println("Orders service listening on", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
