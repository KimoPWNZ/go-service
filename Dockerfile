# Многоступенчатая сборка для уменьшения размера образа
FROM golang:1.23-alpine AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем файлы зависимостей
COPY go.mod go.sum ./

# Скачиваем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN go build -o main ./cmd/server

# Финальный этап: используем минимальный образ
FROM alpine:latest

# Устанавливаем ca-certificates для HTTPS-запросов
RUN apk --no-cache add ca-certificates

# Копируем бинарный файл из builder
COPY --from=builder /app/main /app/main

# Копируем .env.example если нужно
COPY --from=builder /app/.env.example /app/.env.example

# Устанавливаем рабочую директорию
WORKDIR /app

# Открываем порт
EXPOSE 8080

# Запускаем приложение
CMD ["./main"]