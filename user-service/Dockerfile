FROM golang:1.21-alpine AS builder

WORKDIR /build

# Устанавливаем необходимые зависимости
RUN apk add --no-cache git protobuf-dev

# Копируем файлы проекта
COPY . .

# Устанавливаем зависимости
RUN cd user-service && go mod download

# Собираем бинарный файл
RUN cd user-service && CGO_ENABLED=0 GOOS=linux go build -o /build/bin/user-service ./cmd/main.go

FROM alpine:latest

WORKDIR /app
COPY --from=builder /build/bin/user-service .
COPY --from=builder /build/user-service/internal/repository/migrations /migrations

EXPOSE 50051
CMD ["./user-service"] 