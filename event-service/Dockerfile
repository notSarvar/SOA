# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Установка зависимостей
RUN apk add --no-cache git

# Копирование файлов
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Сборка приложения
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/server

# Final stage
FROM alpine:3.18

WORKDIR /app

# Установка необходимых пакетов
RUN apk add --no-cache ca-certificates tzdata

# Копирование бинарного файла из builder
COPY --from=builder /app/server .

# Копирование конфигурационных файлов
COPY config/config.yaml ./config/

# Создание директории для миграций
COPY migrations ./migrations

# Установка переменных окружения
ENV TZ=UTC

# Открытие порта
EXPOSE 50053

# Запуск приложения
CMD ["./server"] 