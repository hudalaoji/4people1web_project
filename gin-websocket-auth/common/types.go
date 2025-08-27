package common

import (
	"encoding/json"
	"github.com/gorilla/websocket"
)

// ClientInterface 客户端接口，用于房间管理
type ClientInterface interface {
	GetUserID() string
	SendMessage(message []byte) // 发送消息方法
}

// SignalingMessage 信令消息格式
type SignalingMessage struct {
	Type string          `json:"type"`   // offer/answer/ice/candidate/join/leave/user_joined/user_left/room_info
	From string          `json:"from"`   // 发送方用户ID
	To   string          `json:"to"`     // 接收方用户ID，空表示广播
	Data json.RawMessage `json:"data"`   // 信令内容
}

// WebSocketClient WebSocket客户端基础结构
type WebSocketClient struct {
	Conn   *websocket.Conn  // WebSocket连接
	UserID string           // 用户ID
	SendCh chan []byte      // 消息发送通道（避免与方法名冲突）
	RoomID string           // 房间ID
}

// GetUserID 实现ClientInterface接口
func (c *WebSocketClient) GetUserID() string {
	return c.UserID
}

// SendMessage 实现ClientInterface接口，发送消息到通道
func (c *WebSocketClient) SendMessage(message []byte) {
	select {
	case c.SendCh <- message: // 向通道发送消息
	default:
		close(c.SendCh)       // 通道满时关闭连接
		c.Conn.Close()
	}
}
