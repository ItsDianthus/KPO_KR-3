# Stage 1: сборка
FROM golang:1.23.7-alpine AS builder
WORKDIR /app

# Копируем только go.mod и загружаем зависимости
COPY go.mod ./
RUN go mod download

# Копируем весь код и собираем бинарник
COPY . .
RUN go build -o api-gateway ./cmd/main.go

# Stage 2: минимальный runtime-образ
FROM alpine:3.18
WORKDIR /app

COPY --from=builder /app/api-gateway .
EXPOSE 8080
ENTRYPOINT ["./api-gateway"]