FROM golang:1.21-alpine AS builder

WORKDIR /build

# Устанавливаем необходимые зависимости для сборки
RUN apk add --no-cache git protobuf-dev

# Копируем файлы проекта
COPY . .

# Скачиваем зависимости
RUN cd api-gateway && go mod download

# Собираем приложение
RUN cd api-gateway && CGO_ENABLED=0 GOOS=linux go build -o /build/bin/api-gateway ./cmd/main.go

FROM alpine:latest

WORKDIR /app
COPY --from=builder /build/bin/api-gateway .

EXPOSE 8080
CMD ["./api-gateway"] 