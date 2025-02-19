# User Service

gRPC сервис для управления пользователями с поддержкой JWT аутентификации.

## Функциональность

- Регистрация пользователей
- Аутентификация (JWT)
- Управление профилем пользователя
- Защита endpoints с помощью middleware
- Хеширование паролей

## Технологии

- Go 1.21
- gRPC
- PostgreSQL
- JWT для аутентификации
- Docker
- Makefile для автоматизации

## Структура проекта

```
.
├── cmd/                    # Точка входа в приложение
├── internal/              # Внутренняя логика сервиса
│   ├── config/           # Конфигурация
│   ├── models/           # Модели данных
│   ├── repository/       # Работа с базой данных
│   │   └── migrations/   # SQL миграции
│   ├── server/          # gRPC сервер и middleware
│   └── service/         # Бизнес-логика
├── Dockerfile            # Сборка контейнера
└── go.mod               # Go модули
```

## Зависимости

- Go 1.21 или выше
- PostgreSQL 15
- Docker и Docker Compose
- Make

## Конфигурация

Сервис настраивается через переменные окружения:

```env
DATABASE_URL=postgres://postgres:postgres@db:5432/promoservice?sslmode=disable
JWT_SECRET=your-secret-key
GRPC_PORT=50051
TOKEN_TTL=24h
BCRYPT_COST=10
```

## Запуск

### Локальный запуск

1. Установите зависимости:
```bash
go mod download
```

2. Запустите PostgreSQL:
```bash
docker run -d \
  --name postgres \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=promoservice \
  -p 5432:5432 \
  postgres:15
```

3. Примените миграции:
```bash
psql -U postgres -d promoservice -f internal/repository/migrations/000001_init_schema.up.sql
```

4. Запустите сервис:
```bash
go run cmd/main.go
```

### Запуск в Docker

1. Соберите образ:
```bash
docker build -t user-service .
```

2. Запустите контейнер:
```bash
docker run -d \
  --name user-service \
  -p 50051:50051 \
  -e DATABASE_URL="postgres://postgres:postgres@db:5432/promoservice?sslmode=disable" \
  -e JWT_SECRET="your-secret-key" \
  -e GRPC_PORT="50051" \
  user-service
```

### Запуск через Docker Compose

```bash
docker-compose up -d
```

## API

### Register

Регистрация нового пользователя.

```protobuf
rpc Register(RegisterRequest) returns (RegisterResponse)
```

Пример запроса:
```json
{
    "login": "testuser",
    "password": "password123",
    "email": "test@example.com"
}
```

### Login

Аутентификация пользователя.

```protobuf
rpc Login(LoginRequest) returns (LoginResponse)
```

Пример запроса:
```json
{
    "login": "testuser",
    "password": "password123"
}
```

### GetProfile

Получение профиля пользователя (требует JWT токен).

```protobuf
rpc GetProfile(GetProfileRequest) returns (GetProfileResponse)
```

### UpdateProfile

Обновление профиля пользователя (требует JWT токен).

```protobuf
rpc UpdateProfile(UpdateProfileRequest) returns (UpdateProfileResponse)
```

Пример запроса:
```json
{
    "email": "new@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "phone": "+1234567890",
    "birth_date": "1990-01-01T00:00:00Z"
}
```

## Тестирование

### Unit тесты

```bash
go test ./...
```

### Тестирование API

Используя grpcurl:

```bash
# Регистрация
grpcurl -plaintext -d '{"login": "testuser", "password": "password123", "email": "test@example.com"}' \
    localhost:50051 user.UserService/Register

# Логин
grpcurl -plaintext -d '{"login": "testuser", "password": "password123"}' \
    localhost:50051 user.UserService/Login

# Получение профиля (с токеном)
grpcurl -plaintext -H 'authorization: Bearer <your-token>' -d '{}' \
    localhost:50051 user.UserService/GetProfile

# Обновление профиля (с токеном)
grpcurl -plaintext -H 'authorization: Bearer <your-token>' \
    -d '{"first_name": "John", "last_name": "Doe"}' \
    localhost:50051 user.UserService/UpdateProfile
```

## Мониторинг

### Логи

```bash
# Docker логи
docker logs user-service

# Docker Compose логи
docker-compose logs user-service
```

### Метрики

Сервис предоставляет следующие метрики:
- Количество активных пользователей
- Время ответа методов
- Количество ошибок

## Разработка

### Добавление нового метода

1. Добавьте определение метода в proto файл
2. Сгенерируйте код:
```bash
make proto
```

3. Реализуйте метод в `internal/server/server.go`
4. Добавьте бизнес-логику в `internal/service`
5. Добавьте тесты

### Создание миграции

```bash
# Создайте файлы миграции
touch internal/repository/migrations/XXXXXX_name.up.sql
touch internal/repository/migrations/XXXXXX_name.down.sql

# Примените миграцию
make migrate-up
```

## Безопасность

- Все пароли хешируются с использованием bcrypt
- JWT токены используются для аутентификации
- Все защищенные endpoints требуют валидный токен
- Реализована защита от основных атак

## Производительность

- Поддержка connection pooling для базы данных
- Эффективная обработка gRPC запросов
- Оптимизированные SQL запросы

## Решение проблем

### Частые ошибки

1. `connection refused`:
   - Проверьте доступность базы данных
   - Проверьте правильность DATABASE_URL

2. `invalid token`:
   - Проверьте срок действия токена
   - Убедитесь, что используется правильный JWT_SECRET

3. `user already exists`:
   - Логин или email уже используются
   - Попробуйте другие данные

## Лицензия

MIT 