# Stage 1: сборка
FROM golang:1.23.7-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o orders-service ./cmd/orders-service/main.go

# Stage 2: минимальный runtime
FROM alpine:3.18
WORKDIR /app
COPY --from=builder /app/orders-service .

EXPOSE 8081
ENTRYPOINT ["./orders-service"]
