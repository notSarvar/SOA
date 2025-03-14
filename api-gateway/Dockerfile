FROM golang:1.21-alpine AS builder

WORKDIR /src

# Устанавливаем необходимые инструменты
RUN apk add --no-cache git protobuf-dev

# Копируем все файлы проекта
COPY . .

# Загружаем зависимости
RUN cd api-gateway && go mod download

# Собираем приложение
RUN cd api-gateway && CGO_ENABLED=0 GOOS=linux go build -o /src/bin/api-gateway ./cmd/main.go

# Создаем финальный образ
FROM alpine:latest

WORKDIR /app

# Копируем бинарный файл из builder
COPY --from=builder /src/bin/api-gateway .

# Запускаем приложение
CMD ["./api-gateway"] 