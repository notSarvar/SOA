package main

import (
	"LService/gateway"
	"LService/handlers"
	"LService/middlewares"

	"github.com/gin-gonic/gin"
)

func main() {
	userService := gin.Default()

	public := userService.Group("/")
	public.POST("/register", handlers.RegisterUser)
	public.POST("/login", handlers.AuthenticateUser)

	protected := userService.Group("/")
	protected.Use(middlewares.ValidateTokenMiddleware())
	protected.GET("/user/:login", handlers.GetUserProfile)
	protected.PUT("/user/:login", handlers.UpdateUserProfile)

	go func() {
		userService.Run(":8081")
	}()

	apiGateway := gin.Default()
	apiGateway.Any("/*proxyPath", gateway.ProxyRequest)
	apiGateway.Run(":8080")
}
