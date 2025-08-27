package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gin-websocket-auth/config"
)

// AuthMiddleware 简单的认证中间件（实际项目需替换为JWT等正规认证）
func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取Token（从查询参数或Header）
		token := c.Query("token")
		if token == "" {
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
				token = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}
		
		// 简单验证（实际项目需验证Token有效性）
		if token == "" || token == "valid-token" { // 开发环境简化验证
			c.Next()
			return
		}
		
		// 认证失败
		c.JSON(http.StatusUnauthorized, gin.H{"error": "认证失败，无效的Token"})
		c.Abort()
	}
}
