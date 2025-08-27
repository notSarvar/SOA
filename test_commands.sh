#!/bin/bash

# Цвета для вывода
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}Начинаем тестирование функционала...${NC}\n"

# 1. Регистрация бизнес-пользователя
echo -e "${GREEN}1. Регистрация бизнес-пользователя${NC}"
BUSINESS_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/users/register \
  -H "Content-Type: application/json" \
  -d '{
    "login": "business77@example.com",
    "password": "password123",
    "email": "busines4s4@example.com",
    "role": "business"
  }')
echo "$BUSINESS_RESPONSE"
BUSINESS_TOKEN=$(echo $BUSINESS_RESPONSE | jq -r '.token')
echo -e "\n"

# 2. Регистрация обычного пользователя
echo -e "${GREEN}2. Регистрация обычного пользователя${NC}"
USER_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/users/register \
  -H "Content-Type: application/json" \
  -d '{
    "login": "user77@example.com",
    "password": "password123",
    "email": "user34@example.com",
    "role": "user"
  }')
echo "$USER_RESPONSE"
USER_TOKEN=$(echo $USER_RESPONSE | jq -r '.token')
echo -e "\n"

# 3. Вход бизнес-пользователя
echo -e "${GREEN}3. Вход бизнес-пользователя${NC}"
BUSINESS_LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/users/login \
  -H "Content-Type: application/json" \
  -d '{
    "login": "business4@example.com",
    "password": "password123"
  }')
echo "$BUSINESS_LOGIN_RESPONSE"
BUSINESS_TOKEN=$(echo $BUSINESS_LOGIN_RESPONSE | jq -r '.token')
echo -e "\n"

# 4. Получение профиля бизнес-пользователя
echo -e "${GREEN}4. Получение профиля бизнес-пользователя${NC}"
curl -s -X GET http://localhost:8080/api/v1/users/profile \
  -H "Authorization: Bearer $BUSINESS_TOKEN"
echo -e "\n"

# 5. Обновление профиля бизнес-пользователя
echo -e "${GREEN}5. Обновление профиля бизнес-пользователя${NC}"
curl -s -X PUT http://localhost:8080/api/v1/users/profile \
  -H "Authorization: Bearer $BUSINESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "Business",
    "last_name": "User",
    "phone": "+79001234567",
    "email": "business4@example.com"
  }'
echo -e "\n"

# 6. Создание промокода бизнес-пользователем
echo -e "${GREEN}6. Создание промокода бизнес-пользователем${NC}"
PROMO_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/promos \
  -H "Authorization: Bearer $BUSINESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Summer Sale 2024",
    "description": "Special summer discount",
    "discount_amount": 20,
    "code": "SUMMER2024_TEST56",
    "valid_until": "2024-12-31T23:59:59",
    "max_uses": 100,
    "is_active": true
  }')
echo "$PROMO_RESPONSE"
PROMO_ID=$(echo $PROMO_RESPONSE | jq -r '.promo.id // empty')
if [ -z "$PROMO_ID" ]; then
    echo -e "${RED}Ошибка: Не удалось получить ID промокода${NC}"
    exit 1
fi
echo -e "\nPromo ID: $PROMO_ID\n"

# 7. Попытка создания промокода обычным пользователем
echo -e "${GREEN}7. Попытка создания промокода обычным пользователем${NC}"
curl -s -X POST http://localhost:8080/api/v1/promos \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Winter Sale 2024",
    "description": "Special winter discount",
    "discount_amount": 15,
    "code": "WINTER2024_TEST",
    "valid_until": "2024-12-31T23:59:59Z",
    "max_uses": 100,
    "is_active": true
  }'
echo -e "\n"

# 8. Получение списка промокодов
echo -e "${GREEN}8. Получение списка промокодов${NC}"
curl -s -X GET http://localhost:8080/api/v1/promos \
  -H "Authorization: Bearer $BUSINESS_TOKEN"
echo -e "\n"

# Пауза для обеспечения создания промокода
sleep 2

# 9. Получение конкретного промокода
echo -e "${GREEN}9. Получение конкретного промокода${NC}"
curl -s -X GET "http://localhost:8080/api/v1/promos/$PROMO_ID" \
  -H "Authorization: Bearer $BUSINESS_TOKEN"
echo -e "\n"

# 10. Обновление промокода
echo -e "${GREEN}10. Обновление промокода${NC}"
curl -s -X PUT "http://localhost:8080/api/v1/promos/$PROMO_ID" \
  -H "Authorization: Bearer $BUSINESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Summer Sale 2024 Updated",
    "description": "Special summer discount updated",
    "discount_amount": 25,
    "code": "SUMMER2024_TEST",
    "valid_until": "2024-12-31T23:59:59Z",
    "max_uses": 1000,
    "is_active": true
  }'
echo -e "\n"

# 11. Попытка удаления промокода обычным пользователем
echo -e "${GREEN}11. Попытка удаления промокода обычным пользователем${NC}"
curl -s -X DELETE "http://localhost:8080/api/v1/promos/$PROMO_ID" \
  -H "Authorization: Bearer $USER_TOKEN"
echo -e "\n"

# 12. Удаление промокода бизнес-пользователем
echo -e "${GREEN}12. Удаление промокода бизнес-пользователем${NC}"
curl -s -X DELETE "http://localhost:8080/api/v1/promos/$PROMO_ID" \
  -H "Authorization: Bearer $BUSINESS_TOKEN"
echo -e "\n"

echo -e "${GREEN}Тестирование завершено!${NC}" 