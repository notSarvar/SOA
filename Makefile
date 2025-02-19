.PHONY: build proto migrate-up migrate-down test clean docker-up docker-down logs help

# Переменные
PROTO_DIR=proto
USER_SERVICE_DIR=user-service
API_GATEWAY_DIR=api-gateway
MIGRATIONS_DIR=$(USER_SERVICE_DIR)/internal/repository/migrations
DOCKER_COMPOSE=docker-compose

# Сборка
build:
	@echo "Building user-service..."
	cd $(USER_SERVICE_DIR) && go build -o bin/user-service ./cmd
	@echo "Building api-gateway..."
	cd $(API_GATEWAY_DIR) && go build -o bin/api-gateway ./cmd

# Генерация proto файлов
proto:
	@echo "Generating proto files..."
	cd $(PROTO_DIR) && protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		user/user.proto

# Миграции
migrate-up:
	@echo "Applying migrations..."
	$(DOCKER_COMPOSE) exec -T db psql -U postgres -d promoservice -f /migrations/000001_init_schema.up.sql

migrate-down:
	@echo "Rolling back migrations..."
	$(DOCKER_COMPOSE) exec -T db psql -U postgres -d promoservice -f /migrations/000001_init_schema.down.sql

# Тесты
test:
	@echo "Running tests..."
	cd $(USER_SERVICE_DIR) && go test -v ./...
	cd $(API_GATEWAY_DIR) && go test -v ./...

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
