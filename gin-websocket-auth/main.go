package main

import (
	"log"
	"net/http" // 添加http包导入
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gin-websocket-auth/config"
	"gin-websocket-auth/middleware"
	"gin-websocket-auth/service"
	"gin-websocket-auth/websocket"
)

func main() {
	// 加载配置
	cfg := config.DefaultConfig()

	// 创建Gin引擎
	r := gin.Default()

	// 配置CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // 生产环境需限制为具体域名
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 健康检查接口
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{ // 现在可以正确使用http.StatusOK
			"status":  "ok",
			"service": "webrtc-signaling-server",
		})
	})

	// 首页接口
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{ // 现在可以正确使用http.StatusOK
			"message": "四人视频通话信令服务器",
			"使用说明": "通过WebSocket连接/ws?user_id=xxx&token=valid-token",
		})
	})

	// WebSocket路由（带认证）
	wsRoute := r.Group("/ws")
	wsRoute.Use(middleware.AuthMiddleware(cfg))
	{
		// 注入service方法，避免循环依赖
		wsRoute.GET("", websocket.HandleWebSocket(cfg, service.JoinRoom, service.CreateRoom))
	}

	// 启动服务器
	log.Printf("服务器启动，监听端口: %s", cfg.ServerPort)
	log.Printf("房间最大容量: 4人")
    if err := r.Run("0.0.0.0:8080"); err != nil {
    log.Fatalf("服务器启动失败: %v", err)
}
}
