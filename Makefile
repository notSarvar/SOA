.PHONY: proto build run test clean migrate-up migrate-down

# Переменные
PROTO_DIR=proto
USER_SERVICE_DIR=user-service
API_GATEWAY_DIR=api-gateway
PROMO_SERVICE_DIR=promo-service
USER_MIGRATIONS_DIR=$(USER_SERVICE_DIR)/internal/repository/migrations
PROMO_MIGRATIONS_DIR=$(PROMO_SERVICE_DIR)/internal/repository/migrations
DOCKER_COMPOSE=docker-compose

# Генерация proto файлов
proto:
	cd proto && \
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		user/user.proto promo/promo.proto

# Сборка всех сервисов
build: proto
	cd api-gateway && go build -o bin/api-gateway cmd/main.go
	cd user-service && go build -o bin/user-service cmd/main.go
	cd promo-service && go build -o bin/promo-service cmd/main.go

# Запуск через Docker Compose
run:
	docker-compose up --build

# Запуск тестов
test:
	cd api-gateway && go test ./...
	cd user-service && go test ./...
	cd promo-service && go test ./...

# Очистка бинарных файлов
clean:
	rm -f api-gateway/bin/*
	rm -f user-service/bin/*
	rm -f promo-service/bin/*

# Установка зависимостей
deps:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Обновление зависимостей
tidy:
	cd api-gateway && go mod tidy
	cd user-service && go mod tidy
	cd promo-service && go mod tidy
	cd proto && go mod tidy

# Миграции
migrate-up:
	@echo "Applying user-service migrations..."
	$(DOCKER_COMPOSE) exec -T db psql -U postgres -d promoservice -f /migrations/user/000001_init_schema.up.sql
	$(DOCKER_COMPOSE) exec -T db psql -U postgres -d promoservice -f /migrations/user/000002_add_role.up.sql
	@echo "Applying promo-service migrations..."
	$(DOCKER_COMPOSE) exec -T db psql -U postgres -d promoservice -f /migrations/promo/000001_init_schema.up.sql

migrate-down:
	@echo "Rolling back promo-service migrations..."
	$(DOCKER_COMPOSE) exec -T db psql -U postgres -d promoservice -f /migrations/promo/000001_init_schema.down.sql
	@echo "Rolling back user-service migrations..."
	$(DOCKER_COMPOSE) exec -T db psql -U postgres -d promoservice -f /migrations/user/000002_add_role.down.sql
	$(DOCKER_COMPOSE) exec -T db psql -U postgres -d promoservice -f /migrations/user/000001_init_schema.down.sql

docker-up:
	@echo "Starting services..."
	$(DOCKER_COMPOSE) up -d --build

docker-down:
	@echo "Stopping services..."
	$(DOCKER_COMPOSE) down

docker-restart:
	@echo "Restarting services..."
	$(DOCKER_COMPOSE) restart

logs:
	$(DOCKER_COMPOSE) logs -f

logs-user:
	$(DOCKER_COMPOSE) logs -f user-service

logs-db:
	$(DOCKER_COMPOSE) logs -f db

# Помощь
help:
	@echo "Available commands:"
	@echo "  build         - Сборка сервисов"
	@echo "  proto         - Генерация proto файлов"
	@echo "  migrate-up    - Применение миграций"
	@echo "  migrate-down  - Откат миграций"
	@echo "  test         - Запуск тестов"
	@echo "  docker-up    - Запуск сервисов в Docker"
	@echo "  docker-down  - Остановка сервисов"
	@echo "  docker-restart - Перезапуск сервисов"
	@echo "  logs         - Просмотр логов всех сервисов"
	@echo "  logs-user    - Просмотр логов user-service"
	@echo "  logs-db      - Просмотр логов базы данных"
	@echo "  help         - Показать эту справку"
