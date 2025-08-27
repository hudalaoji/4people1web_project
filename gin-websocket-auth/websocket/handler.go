package websocket

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gin-websocket-auth/config"
)

// 升级器配置
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 开发环境允许所有来源，生产环境需限制
	},
}

// HandleWebSocket 处理WebSocket连接（通过回调解耦依赖）
func HandleWebSocket(
	cfg *config.Config,
	joinRoom func(roomID, userID string, conn *websocket.Conn, cfg *config.Config) (bool, string),
	createRoom func(string) string,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取参数
		userID := c.Query("user_id")
		roomID := c.Query("room_id")
		
		// 验证用户ID
		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user_id参数必填"})
			return
		}

		// 升级HTTP连接为WebSocket
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "连接升级失败"})
			return
		}

		// 处理房间逻辑
		if roomID != "" {
			// 加入指定房间
			success, msg := joinRoom(roomID, userID, conn, cfg)
			if !success {
				conn.WriteMessage(websocket.TextMessage, []byte(msg))
				conn.Close()
				c.JSON(http.StatusBadRequest, gin.H{"error": msg})
				return
			}
		} else {
			// 创建新房间并加入
			newRoomID := createRoom("")
			success, msg := joinRoom(newRoomID, userID, conn, cfg)
			if !success {
				conn.WriteMessage(websocket.TextMessage, []byte(msg))
				conn.Close()
				c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
				return
			}
			
			// 通知客户端房间创建成功
			conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"room_created","room_id":"` + newRoomID + `"}`))
		}
	}
}
