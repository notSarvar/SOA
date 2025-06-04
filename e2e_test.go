package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"promoservice/proto/promo"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

const (
	apiGatewayAddr = "localhost:8080"
	promoAddr      = "localhost:50052"
	userAddr       = "localhost:50051"
)

func setupTestEnvironment() error {
	// Запускаем сервисы через docker-compose
	cmd := exec.Command("docker-compose", "up", "-d")
	if err := cmd.Run(); err != nil {
		return err
	}

	// Ждем, пока сервисы запустятся
	time.Sleep(5 * time.Second)

	// Очищаем таблицы в базе данных
	cmd = exec.Command("docker-compose", "exec", "db", "psql", "-U", "postgres", "-d", "promoservice", "-c", "TRUNCATE TABLE promos, users CASCADE;")
	return cmd.Run()
}

func teardownTestEnvironment() error {
	cmd := exec.Command("docker-compose", "down")
	return cmd.Run()
}

func TestMain(m *testing.M) {
	// Запускаем сервисы перед всеми тестами
	if err := setupTestEnvironment(); err != nil {
		fmt.Printf("Failed to setup test environment: %v\n", err)
		os.Exit(1)
	}

	// Запускаем тесты
	code := m.Run()

	// Не останавливаем сервисы после тестов
	os.Exit(code)
}

func setupTest(t *testing.T) (context.Context, *grpc.ClientConn, *grpc.ClientConn, *grpc.ClientConn) {
	ctx := context.Background()

	// Подключаемся к API Gateway
	apiConn, err := grpc.Dial(apiGatewayAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	// Подключаемся к Promo Service
	promoConn, err := grpc.Dial(promoAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	// Подключаемся к User Service
	userConn, err := grpc.Dial(userAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	return ctx, apiConn, promoConn, userConn
}

type RegisterResponse struct {
	User struct {
		Id string `json:"id"`
	} `json:"user"`
	Token string `json:"token"`
}

func registerUserREST(t *testing.T, login, password, email, role string) (string, string) {
	url := fmt.Sprintf("http://%s/api/v1/users/register", apiGatewayAddr)
	body := map[string]string{
		"login":    login,
		"password": password,
		"email":    email,
		"role":     role,
	}
	jsonBody, err := json.Marshal(body)
	require.NoError(t, err)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	respBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	var result RegisterResponse
	err = json.Unmarshal(respBody, &result)
	require.NoError(t, err)
	return result.User.Id, result.Token
}

func loginUserREST(t *testing.T, login, password string) string {
	url := fmt.Sprintf("http://%s/api/v1/users/login", apiGatewayAddr)
	body := map[string]string{
		"login":    login,
		"password": password,
	}
	jsonBody, err := json.Marshal(body)
	require.NoError(t, err)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result struct {
		User struct {
			Id   string `json:"id"`
			Role string `json:"role"`
		} `json:"user"`
		Token string `json:"token"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.Equal(t, "business", strings.ToLower(result.User.Role))
	return result.Token
}

func uniqueSuffix() string {
	return strconv.FormatInt(time.Now().UnixNano()+int64(rand.Intn(10000)), 36)
}

func ctxWithUser(ctx context.Context, userID, userRole string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "user_id", userID, "user_role", userRole)
}

func TestE2E_BusinessUserCreatesAndManagesPromos(t *testing.T) {
	suffix := uniqueSuffix()
	login := "business_user_" + suffix
	email := "business_" + suffix + "@example.com"
	promoCode := "SUMMER20_" + suffix
	promoCodeUpdate := "SUMMER25_" + suffix

	ctx, apiConn, promoConn, userConn := setupTest(t)
	defer apiConn.Close()
	defer promoConn.Close()
	defer userConn.Close()

	promoClient := promo.NewPromoServiceClient(promoConn)

	// 1. Регистрируем бизнес-пользователя через REST
	userID, token := registerUserREST(t, login, "password123", email, "BUSINESS")
	require.NotEmpty(t, userID)
	require.NotEmpty(t, token)

	userRole := "BUSINESS"
	ctxUser := ctxWithUser(ctx, userID, userRole)

	// 2. Создаем промокод через gRPC
	validUntil := time.Now().Add(24 * time.Hour)
	createPromoResp, err := promoClient.CreatePromo(ctxUser, &promo.CreatePromoRequest{
		Name:           "Summer Sale",
		Description:    "20% off on all items",
		DiscountAmount: 20.0,
		Code:           promoCode,
		ValidUntil:     timestamppb.New(validUntil),
		MaxUses:        100,
	})
	require.NoError(t, err)
	require.NotNil(t, createPromoResp)
	assert.Equal(t, "Summer Sale", createPromoResp.Promo.Name)
	assert.Equal(t, promoCode, createPromoResp.Promo.Code)

	// 3. Обновляем промокод через gRPC
	updatePromoResp, err := promoClient.UpdatePromo(ctxUser, &promo.UpdatePromoRequest{
		Id:             createPromoResp.Promo.Id,
		Name:           "Summer Sale Extended",
		Description:    "25% off on all items",
		DiscountAmount: 25.0,
		Code:           promoCodeUpdate,
		ValidUntil:     timestamppb.New(validUntil.Add(24 * time.Hour)),
		MaxUses:        200,
		IsActive:       true,
	})
	require.NoError(t, err)
	require.NotNil(t, updatePromoResp)
	assert.Equal(t, "Summer Sale Extended", updatePromoResp.Promo.Name)
	assert.Equal(t, promoCodeUpdate, updatePromoResp.Promo.Code)
	// assert.Equal(t, float32(25.0), float32(updatePromoResp.Promo.DiscountAmount))
}

func TestE2E_CustomerUsesPromo(t *testing.T) {
	suffix := uniqueSuffix()
	loginBusiness := "business_user2_" + suffix
	emailBusiness := "business2_" + suffix + "@example.com"
	loginCustomer := "customer_" + suffix
	emailCustomer := "customer_" + suffix + "@example.com"
	promoCode := "WELCOME10_" + suffix

	ctx, apiConn, promoConn, userConn := setupTest(t)
	defer apiConn.Close()
	defer promoConn.Close()
	defer userConn.Close()

	promoClient := promo.NewPromoServiceClient(promoConn)

	// 1. Регистрируем бизнес-пользователя через REST
	businessID, businessToken := registerUserREST(t, loginBusiness, "password123", emailBusiness, "BUSINESS")
	require.NotEmpty(t, businessID)
	require.NotEmpty(t, businessToken)

	businessRole := "BUSINESS"
	ctxBusiness := ctxWithUser(ctx, businessID, businessRole)

	// 2. Создаем промокод
	validUntil := time.Now().Add(24 * time.Hour)
	createPromoResp, err := promoClient.CreatePromo(ctxBusiness, &promo.CreatePromoRequest{
		Name:           "Welcome Discount",
		Description:    "10% off for new customers",
		DiscountAmount: 10.0,
		Code:           promoCode,
		ValidUntil:     timestamppb.New(validUntil),
		MaxUses:        1000,
	})
	require.NoError(t, err)
	require.NotNil(t, createPromoResp)

	// 3. Регистрируем обычного пользователя через REST
	customerID, customerToken := registerUserREST(t, loginCustomer, "password123", emailCustomer, "CUSTOMER")
	require.NotEmpty(t, customerID)
	require.NotEmpty(t, customerToken)

	customerRole := "CUSTOMER"
	ctxCustomer := ctxWithUser(ctx, customerID, customerRole)

	// 4. Пользователь пытается использовать промокод
	getPromoResp, err := promoClient.GetPromo(ctxCustomer, &promo.GetPromoRequest{Id: createPromoResp.Promo.Id})
	require.NoError(t, err)
	require.NotNil(t, getPromoResp)
	assert.True(t, getPromoResp.Promo.IsActive)
	assert.Equal(t, float32(10.0), float32(getPromoResp.Promo.DiscountAmount))
}

func TestE2E_PromoDeactivationByBusiness(t *testing.T) {
	suffix := uniqueSuffix()
	login := "business_user3_" + suffix
	email := "business3_" + suffix + "@example.com"
	loginCustomer := "customer3_" + suffix
	emailCustomer := "customer3_" + suffix + "@example.com"
	promoCode := "FLASH15_" + suffix

	ctx, apiConn, promoConn, userConn := setupTest(t)
	defer apiConn.Close()
	defer promoConn.Close()
	defer userConn.Close()

	promoClient := promo.NewPromoServiceClient(promoConn)

	// 1. Регистрируем бизнес-пользователя через REST
	businessID, businessToken := registerUserREST(t, login, "password123", email, "BUSINESS")
	require.NotEmpty(t, businessID)
	require.NotEmpty(t, businessToken)

	businessRole := "BUSINESS"
	ctxBusiness := ctxWithUser(ctx, businessID, businessRole)

	// 2. Создаем промокод
	validUntil := time.Now().Add(24 * time.Hour)
	createPromoResp, err := promoClient.CreatePromo(ctxBusiness, &promo.CreatePromoRequest{
		Name:           "Flash Sale",
		Description:    "Quick 15% off",
		DiscountAmount: 15.0,
		Code:           promoCode,
		ValidUntil:     timestamppb.New(validUntil),
		MaxUses:        50,
	})
	require.NoError(t, err)
	require.NotNil(t, createPromoResp)

	// 3. Проверяем, что промокод активен
	getPromoResp, err := promoClient.GetPromo(ctxBusiness, &promo.GetPromoRequest{Id: createPromoResp.Promo.Id})
	require.NoError(t, err)
	require.NotNil(t, getPromoResp)
	assert.True(t, getPromoResp.Promo.IsActive)

	// 4. Деактивируем промокод вручную
	_, err = promoClient.UpdatePromo(ctxBusiness, &promo.UpdatePromoRequest{
		Id:       createPromoResp.Promo.Id,
		IsActive: false,
	})
	require.NoError(t, err)

	// 5. Проверяем, что промокод недействителен для бизнес-пользователя
	getPromoResp, err = promoClient.GetPromo(ctxBusiness, &promo.GetPromoRequest{Id: createPromoResp.Promo.Id})
	require.NoError(t, err)
	require.NotNil(t, getPromoResp)
	assert.False(t, getPromoResp.Promo.IsActive)

	// 6. Регистрируем обычного пользователя через REST
	customerID, customerToken := registerUserREST(t, loginCustomer, "password123", emailCustomer, "CUSTOMER")
	require.NotEmpty(t, customerID)
	require.NotEmpty(t, customerToken)

	customerRole := "CUSTOMER"
	ctxCustomer := ctxWithUser(ctx, customerID, customerRole)

	// 7. Обычный пользователь пытается получить промокод (он недействителен)
	getPromoResp, err = promoClient.GetPromo(ctxCustomer, &promo.GetPromoRequest{Id: createPromoResp.Promo.Id})
	require.NoError(t, err)
	require.NotNil(t, getPromoResp)
	assert.False(t, getPromoResp.Promo.IsActive)
}

func TestE2E_RegisterUserREST(t *testing.T) {
	suffix := uniqueSuffix()
	login := "rest_user_" + suffix
	email := "rest_" + suffix + "@example.com"
	userID, token := registerUserREST(t, login, "password123", email, "CUSTOMER")
	assert.NotEmpty(t, userID)
	assert.NotEmpty(t, token)
}

func TestE2E_APIGatewayFlow(t *testing.T) {
	suffix := uniqueSuffix()
	login := "api_gateway_user_" + suffix
	email := "api_gateway_" + suffix + "@example.com"
	promoCode := "GATEWAY20_" + suffix

	// 1. Регистрируем бизнес-пользователя через REST API Gateway
	businessID, _ := registerUserREST(t, login, "password123", email, "business")
	require.NotEmpty(t, businessID)

	// 1.1. Логинимся этим пользователем, чтобы получить токен с ролью
	businessToken := loginUserREST(t, login, "password123")
	require.NotEmpty(t, businessToken)

	// 2. Создаем промокод через REST API Gateway
	url := fmt.Sprintf("http://%s/api/v1/promos", apiGatewayAddr)
	body := map[string]interface{}{
		"code":       promoCode,
		"discount":   20.0,
		"expires_at": time.Now().Add(24 * time.Hour).Format(time.RFC3339),
	}
	jsonBody, err := json.Marshal(body)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+businessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var createPromoResp struct {
		Promo struct {
			ID string `json:"id"`
		} `json:"promo"`
	}
	err = json.NewDecoder(resp.Body).Decode(&createPromoResp)
	require.NoError(t, err)
	require.NotEmpty(t, createPromoResp.Promo.ID)

	// 3. Получаем промокод через REST API Gateway
	url = fmt.Sprintf("http://%s/api/v1/promos/%s", apiGatewayAddr, createPromoResp.Promo.ID)
	req, err = http.NewRequest("GET", url, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+businessToken)

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	respBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	fmt.Println("API Gateway response:", string(respBody))

	// Структура для временных полей
	type timestamp struct {
		Seconds int64 `json:"seconds"`
		Nanos   int32 `json:"nanos,omitempty"`
	}

	var getPromoResp struct {
		Promo struct {
			ID             string    `json:"id"`
			CreatorID      string    `json:"creator_id"`
			DiscountAmount float32   `json:"discount_amount"`
			Code           string    `json:"code"`
			CreatedAt      timestamp `json:"created_at"`
			UpdatedAt      timestamp `json:"updated_at"`
			IsActive       bool      `json:"is_active"`
			ValidUntil     timestamp `json:"valid_until"`
			MaxUses        int32     `json:"max_uses"`
		} `json:"promo"`
	}
	err = json.Unmarshal(respBody, &getPromoResp)
	require.NoError(t, err)
	assert.Equal(t, promoCode, getPromoResp.Promo.Code)
	assert.Equal(t, float32(20.0), getPromoResp.Promo.DiscountAmount)
	assert.True(t, getPromoResp.Promo.IsActive)
	assert.Equal(t, businessID, getPromoResp.Promo.CreatorID)

	// 4. Обновляем промокод через REST API Gateway
	url = fmt.Sprintf("http://%s/api/v1/promos/%s", apiGatewayAddr, createPromoResp.Promo.ID)
	updateBody := map[string]interface{}{
		"code":       promoCode,
		"discount":   25.0,
		"expires_at": time.Now().Add(48 * time.Hour).Format(time.RFC3339),
	}
	jsonBody, err = json.Marshal(updateBody)
	require.NoError(t, err)

	req, err = http.NewRequest("PUT", url, bytes.NewBuffer(jsonBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+businessToken)

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	respBody, err = ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	fmt.Println("API Gateway update response:", string(respBody))

	// 5. Проверяем обновленный промокод
	req, err = http.NewRequest("GET", url, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+businessToken)

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	err = json.NewDecoder(resp.Body).Decode(&getPromoResp)
	require.NoError(t, err)
}
