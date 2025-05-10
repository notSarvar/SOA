FROM golang:1.21-alpine AS builder

WORKDIR /src

# Устанавливаем необходимые инструменты
RUN apk add --no-cache git protobuf-dev

# Копируем все файлы проекта
COPY . .

# Загружаем зависимости
RUN cd promo-service && go mod download

# Собираем приложение
RUN cd promo-service && CGO_ENABLED=0 GOOS=linux go build -o /src/bin/promo-service ./cmd/main.go

# Создаем финальный образ
FROM alpine:latest

WORKDIR /app

# Копируем бинарный файл из builder
COPY --from=builder /src/bin/promo-service .

# Запускаем приложение
CMD ["./promo-service"] 