FROM golang:1.23.7-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o payments-service ./cmd/payments-service/main.go

FROM alpine:3.18
WORKDIR /app
COPY --from=builder /app/payments-service .
EXPOSE 8082
ENTRYPOINT ["./payments-service"]
