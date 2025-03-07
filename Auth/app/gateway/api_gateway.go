// gateway/api_gateway.go
package gateway

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ProxyRequest(c *gin.Context) {
	proxyURL := fmt.Sprintf("http://user-service:8081%s", c.Request.URL.Path)
	c.Redirect(http.StatusTemporaryRedirect, proxyURL)
}

// Дополнительная логика API Gateway, например аутентификация или обработка ошибок
func SetupRoutes(r *gin.Engine) {
	r.Any("/*proxyPath", ProxyRequest)
}
