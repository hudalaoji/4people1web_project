package service

import (
	"encoding/json"
	"log"
	"math/rand"
	"time"

	"github.com/gorilla/websocket"
	"gin-websocket-auth/common"
	"gin-websocket-auth/config"
	"gin-websocket-auth/room"
)

// CreateRoom 生成房间ID并初始化房间
func CreateRoom(roomID string) string {
	if roomID == "" {
		roomID = generateRandomRoomID()
	}
	room.GlobalRoomManager.CreateRoom(roomID)
	log.Printf("创建房间: %s", roomID)
	return roomID
}

// JoinRoom 将用户加入房间并广播事件
func JoinRoom(roomID, userID string, conn *websocket.Conn, cfg *config.Config) (bool, string) {
	// 参数验证
	if roomID == "" {
		return false, "房间ID不能为空"
	}
	if userID == "" {
		return false, "用户ID不能为空"
	}
	if conn == nil {
		return false, "WebSocket连接不能为空"
	}
	
	// 获取或创建房间
	r, exists := room.GlobalRoomManager.GetRoom(roomID)
	if !exists {
		r = room.GlobalRoomManager.CreateRoom(roomID)
		log.Printf("房间不存在，自动创建: %s", roomID)
	}
	
	// 创建客户端实例
	client := &common.WebSocketClient{
		Conn:   conn,
		UserID: userID,
		RoomID: roomID,
		SendCh: make(chan []byte, 256), // 使用正确的通道字段名
	}
	
	// 检查房间容量
	if !r.AddClient(client) {
		return false, "房间已满（最多4人）"
	}
	
	// 启动消息读写协程
	go readPump(client, cfg)
	go writePump(client, cfg)
	
	// 广播新用户加入事件
	joinMsg := common.SignalingMessage{
		Type: "user_joined",
		From: userID,
		To:   "",
		Data: json.RawMessage(`{"room_id": "` + roomID + `"}`),
	}
	data, _ := json.Marshal(joinMsg)
	r.Broadcast(userID, data)
	
	// 向新用户发送现有成员信息
	existingClients := r.GetOtherClients(userID)
	userList := make([]string, len(existingClients))
	for i, cli := range existingClients {
		userList[i] = cli.GetUserID()
	}
	
	roomInfoMsg := common.SignalingMessage{
		Type: "room_info",
		From: "server",
		To:   userID,
		Data: json.RawMessage(func() []byte {
			data, _ := json.Marshal(userList)
			return data
		}()),
	}
	data, _ = json.Marshal(roomInfoMsg)
	client.SendMessage(data)
	
	log.Printf("用户 %s 加入房间 %s，当前成员: %v", userID, roomID, userList)
	return true, "加入成功"
}

// LeaveRoom 处理用户离开房间
func LeaveRoom(roomID, userID string) {
	if roomID == "" || userID == "" {
		log.Printf("离开房间参数无效: roomID=%s, userID=%s", roomID, userID)
		return
	}
	
	r, exists := room.GlobalRoomManager.GetRoom(roomID)
	if !exists {
		log.Printf("房间不存在，无法离开: %s", roomID)
		return
	}
	
	// 从房间移除用户
	r.RemoveClient(userID)
	
	// 广播用户离开事件
	leaveMsg := common.SignalingMessage{
		Type: "user_left",
		From: userID,
		To:   "",
		Data: json.RawMessage(`{"room_id": "` + roomID + `"}`),
	}
	data, _ := json.Marshal(leaveMsg)
	r.Broadcast(userID, data)
	
	log.Printf("用户 %s 离开房间 %s", userID, roomID)
}

// 生成随机房间ID
func generateRandomRoomID() string {
	rand.Seed(time.Now().UnixNano())
	chars := "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, 6)
	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

// 读取客户端消息的协程
func readPump(client *common.WebSocketClient, cfg *config.Config) {
	defer func() {
		client.Conn.Close()
		LeaveRoom(client.RoomID, client.UserID)
	}()
	
	// 配置读取参数（已修复时间类型）
	client.Conn.SetReadLimit(cfg.MaxMessageSize)
	client.Conn.SetReadDeadline(time.Now().Add(cfg.ReadTimeout))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(cfg.ReadTimeout))
		return nil
	})
	
	// 循环读取消息
	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("连接错误: %v", err)
			}
			break
		}
		
		// 处理消息
		handleMessage(client, message)
	}
}

// 向客户端发送消息的协程
func writePump(client *common.WebSocketClient, cfg *config.Config) {
	ticker := time.NewTicker(30 * time.Second) // 心跳定时器
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()
	
	for {
		select {
		case message, ok := <-client.SendCh:
			client.Conn.SetWriteDeadline(time.Now().Add(cfg.WriteTimeout)) // 已修复时间类型
			if !ok {
				// 通道关闭，发送关闭消息
				client.Conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				return
			}
			
			// 写入消息
			w, err := client.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)
			
			// 发送队列中所有消息
			n := len(client.SendCh)
			for i := 0; i < n; i++ {
				w.Write(<-client.SendCh)
			}
			
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			// 发送心跳消息
			client.Conn.SetWriteDeadline(time.Now().Add(cfg.WriteTimeout)) // 已修复时间类型
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// 处理客户端消息
func handleMessage(client *common.WebSocketClient, data []byte) {
	var msg common.SignalingMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("解析消息失败: %v", err)
		sendError(client, "消息格式无效")
		return
	}
	
	// 填充发送者ID
	msg.From = client.UserID
	
	// 验证房间状态
	if client.RoomID == "" {
		sendError(client, "未加入任何房间")
		return
	}
	
	// 获取房间
	roomInstance, exists := room.GlobalRoomManager.GetRoom(client.RoomID)
	if !exists {
		sendError(client, "房间不存在")
		return
	}
	
	// 转发消息
	if msg.To != "" {
		forwardToUser(roomInstance, msg)
	} else {
		broadcastToRoom(roomInstance, msg, client.UserID)
	}
}

// 转发消息给指定用户
func forwardToUser(room *room.Room, msg common.SignalingMessage) {
	room.Mu.RLock() // 使用正确的导出锁字段
	defer room.Mu.RUnlock()
	
	for userID, client := range room.Clients {
		if userID == msg.To {
			data, _ := json.Marshal(msg)
			client.SendMessage(data)
			return
		}
	}
}

// 广播消息到房间
func broadcastToRoom(room *room.Room, msg common.SignalingMessage, senderID string) {
	data, _ := json.Marshal(msg)
	room.Broadcast(senderID, data)
}

// 发送错误消息给客户端
func sendError(client *common.WebSocketClient, message string) {
	errMsg := common.SignalingMessage{
		Type: "error",
		From: "server",
		To:   client.UserID,
		Data: json.RawMessage(`{"message": "` + message + `"}`),
	}
	data, _ := json.Marshal(errMsg)
	client.SendMessage(data)
}
