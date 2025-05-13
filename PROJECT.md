# PromoService - Система управления промокодами

## Общее описание
PromoService - это микросервисная система для управления промокодами, состоящая из нескольких сервисов, взаимодействующих через gRPC и HTTP API. Система предоставляет функционал для создания, управления и отслеживания использования промокодов, а также аналитики их эффективности.

## Архитектура системы

### 1. API Gateway (api-gateway)
- Основной входной пункт для всех клиентских запросов
- Обрабатывает HTTP запросы и маршрутизирует их к соответствующим микросервисам
- Реализует аутентификацию и авторизацию через JWT
- Использует Gin для обработки HTTP запросов
- Основные эндпоинты:
  - `/api/v1/users/*` - управление пользователями
  - `/api/v1/promos/*` - управление промокодами
  - `/api/v1/promos/*/events/*` - отслеживание событий

#### Необходимые изменения:
1. Добавить новые эндпоинты:
   ```go
   // Комментарии
   POST /api/v1/promos/{id}/comments
   GET /api/v1/promos/{id}/comments
   
   // События
   POST /api/v1/promos/{id}/view
   POST /api/v1/promos/{id}/click
   POST /api/v1/promos/{id}/like
   ```

2. Обновить обработчики для работы с событиями:
   - Добавить валидацию входных данных
   - Добавить логирование событий
   - Реализовать обработку ошибок
   - Добавить rate limiting
   - Добавить кэширование для часто запрашиваемых данных
   - Добавить метрики для мониторинга

### 2. User Service (user-service)
- Управление пользователями и их профилями
- Аутентификация и авторизация
- Хранение данных пользователей в PostgreSQL
- Основные функции:
  - Регистрация и авторизация
  - Управление профилем
  - Ролевая модель (админ, менеджер, пользователь)

#### Необходимые изменения:
1. Добавить отправку событий в Kafka:
   ```go
   // При регистрации пользователя
   type UserRegistrationEvent struct {
       UserID        string    `json:"user_id"`
       Email         string    `json:"email"`
       RegisteredAt  time.Time `json:"registered_at"`
       FirstName     string    `json:"first_name,omitempty"`
       LastName      string    `json:"last_name,omitempty"`
   }
   
   // При обновлении профиля
   type UserProfileUpdateEvent struct {
       UserID        string    `json:"user_id"`
       UpdatedAt     time.Time `json:"updated_at"`
       UpdatedFields []string  `json:"updated_fields"`
   }
   ```

2. Интеграция с Kafka:
   - Добавить конфигурацию для подключения к Kafka
   - Реализовать отправку событий в соответствующие топики
   - Добавить обработку ошибок при отправке событий
   - Добавить retry механизм для отправки событий
   - Добавить метрики для мониторинга отправки событий

3. Улучшения в существующем коде:
   - Добавить валидацию email и других полей
   - Улучшить обработку ошибок при работе с базой данных
   - Добавить метрики и логирование
   - Добавить проверку роли пользователя при обновлении профиля
   - Добавить транзакции для атомарных операций

### 3. Promo Service (promo-service)
- Управление промокодами
- Хранение данных о промокодах в PostgreSQL
- Основные функции:
  - Создание и управление промокодами
  - Валидация промокодов
  - Отслеживание использования
  - Управление комментариями к промокодам

#### Необходимые изменения:
1. Добавить таблицу для комментариев:
   ```sql
   CREATE TABLE promo_comments (
       id UUID PRIMARY KEY,
       promo_id UUID NOT NULL,
       user_id UUID NOT NULL,
       comment_text TEXT NOT NULL,
       created_at TIMESTAMP NOT NULL,
       updated_at TIMESTAMP NOT NULL,
       FOREIGN KEY (promo_id) REFERENCES promos(id),
       FOREIGN KEY (user_id) REFERENCES users(id)
   );
   ```

2. Добавить отправку событий в Kafka:
   ```go
   // При создании промокода
   type PromoCreatedEvent struct {
       PromoID      string    `json:"promo_id"`
       CreatorID    string    `json:"creator_id"`
       CreatedAt    time.Time `json:"created_at"`
       Code         string    `json:"code"`
       DiscountAmount float32 `json:"discount_amount"`
   }
   
   // При обновлении промокода
   type PromoUpdatedEvent struct {
       PromoID      string    `json:"promo_id"`
       UpdatedBy    string    `json:"updated_by"`
       UpdatedAt    time.Time `json:"updated_at"`
       UpdatedFields []string `json:"updated_fields"`
   }
   
   // При удалении промокода
   type PromoDeletedEvent struct {
       PromoID      string    `json:"promo_id"`
       DeletedBy    string    `json:"deleted_by"`
       DeletedAt    time.Time `json:"deleted_at"`
   }
   
   // При просмотре промокода
   type PromoViewEvent struct {
       PromoID    string    `json:"promo_id"`
       UserID     string    `json:"user_id"`
       ViewedAt   time.Time `json:"viewed_at"`
   }
   
   // При клике по промокоду
   type PromoClickEvent struct {
       PromoID    string    `json:"promo_id"`
       UserID     string    `json:"user_id"`
       ClickedAt  time.Time `json:"clicked_at"`
   }
   
   // При добавлении комментария
   type PromoCommentEvent struct {
       PromoID      string    `json:"promo_id"`
       UserID       string    `json:"user_id"`
       CommentText  string    `json:"comment_text"`
       CommentedAt  time.Time `json:"commented_at"`
   }
   ```

3. Реализовать методы для работы с комментариями:
   - Добавление комментария
   - Получение списка комментариев
   - Удаление комментария
   - Обновление комментария

4. Улучшения в существующем коде:
   - Добавить валидацию входных данных
   - Улучшить обработку ошибок при работе с базой данных
   - Добавить метрики и логирование
   - Добавить проверку срока действия промокода
   - Добавить проверку лимита использования промокода
   - Добавить транзакции для атомарных операций
   - Добавить кэширование для часто запрашиваемых данных

### 4. Event Service (event-service)
- Сбор и обработка событий
- Отправка событий в Kafka для аналитики
- Основные события:
  - Просмотр промокода
  - Клик по промокоду
  - Добавление комментария
  - Регистрация пользователя
- Использует Kafka для асинхронной обработки событий

#### Необходимые изменения:
1. Добавить новые топики в Kafka:
   ```
   - user-registrations
   - user-profile-updates
   - promo-created
   - promo-updated
   - promo-deleted
   - promo-views
   - promo-clicks
   - promo-comments
   - promo-likes
   ```

2. Реализовать обработчики для новых событий:
   ```go
   // Обработчик регистрации пользователя
   func (s *Service) HandleUserRegistration(event *UserRegistrationEvent) error
   
   // Обработчик обновления профиля
   func (s *Service) HandleUserProfileUpdate(event *UserProfileUpdateEvent) error
   
   // Обработчик создания промокода
   func (s *Service) HandlePromoCreated(event *PromoCreatedEvent) error
   
   // Обработчик обновления промокода
   func (s *Service) HandlePromoUpdated(event *PromoUpdatedEvent) error
   
   // Обработчик удаления промокода
   func (s *Service) HandlePromoDeleted(event *PromoDeletedEvent) error
   
   // Обработчик просмотра промокода
   func (s *Service) HandlePromoView(event *PromoViewEvent) error
   
   // Обработчик клика по промокоду
   func (s *Service) HandlePromoClick(event *PromoClickEvent) error
   
   // Обработчик комментария к промокоду
   func (s *Service) HandlePromoComment(event *PromoCommentEvent) error
   ```

3. Добавить конфигурацию для Kafka:
   ```yaml
   kafka:
     brokers:
       - localhost:9092
     topics:
       user-registrations: user-registrations
       user-profile-updates: user-profile-updates
       promo-created: promo-created
       promo-updated: promo-updated
       promo-deleted: promo-deleted
       promo-views: promo-views
       promo-clicks: promo-clicks
       promo-comments: promo-comments
       promo-likes: promo-likes
     consumer:
       group_id: event-service
       auto_offset_reset: earliest
     producer:
       acks: all
       retries: 3
   ```

4. Улучшения в существующем коде:
   - Добавить валидацию входных данных
   - Улучшить обработку ошибок при работе с Kafka
   - Добавить метрики и логирование
   - Добавить retry механизм для обработки событий
   - Добавить dead letter queue для неудачных событий
   - Добавить мониторинг очередей Kafka

## Технологический стек
- Go (основной язык программирования)
- gRPC (межсервисное взаимодействие)
- PostgreSQL (основное хранилище данных)
- Kafka (обработка событий)
- Gin (HTTP фреймворк для API Gateway)
- JWT (аутентификация)
- Docker (контейнеризация)
- Docker Compose (оркестрация)
- Kafka UI (мониторинг Kafka)

## Структура проекта
```
.
├── api-gateway/          # API Gateway сервис
├── user-service/         # Сервис управления пользователями
├── promo-service/        # Сервис управления промокодами
├── event-service/        # Сервис обработки событий
├── proto/               # Protobuf определения
├── docker-compose.yml   # Конфигурация Docker Compose
└── Makefile            # Скрипты сборки и запуска
```

## Взаимодействие сервисов
1. Клиент делает запрос к API Gateway
2. API Gateway аутентифицирует запрос и перенаправляет к соответствующему сервису
3. Сервисы взаимодействуют между собой через gRPC
4. Event Service собирает события и отправляет их в Kafka
5. Аналитические системы могут подписываться на события через Kafka

## Особенности реализации
- Микросервисная архитектура обеспечивает масштабируемость и изоляцию
- Асинхронная обработка событий через Kafka
- Единый протокол взаимодействия через gRPC
- Централизованная аутентификация через API Gateway
- Контейнеризация для упрощения развертывания
- Мониторинг событий через Kafka UI 

1. **Makefile**:
   - Makefile находится в корневой директории проекта и предназначен для сборки всего проекта. Однако, каждый микросервис (например, event-service, user-service и promo-service) должен иметь возможность собираться независимо, без использования общего Makefile. Это позволяет упростить процесс разработки и тестирования отдельных сервисов.

2. **Расположение proto-файлов**:
   - Все proto-файлы должны находиться в корневой директории проекта в папке `proto/`. Это упрощает доступ к ним и делает структуру проекта более понятной. Важно, чтобы все сервисы ссылались на эти файлы, используя относительные пути.

3. **Наименование mod файлов**:
   - В каждом модуле (например, в event-service и proto) необходимо использовать правильные наименования для go.mod файлов. Они должны следовать формату `module promoservice/<имя_сервиса>` для микросервисов и `module promoservice/proto` для директории с proto-файлами. Это обеспечит корректную работу с зависимостями и упростит импорт.