package websocket

import (
	"encoding/json"
	"log"

	"gin-websocket-auth/common"
	"gin-websocket-auth/room"
)

// HandleSignaling 处理接收到的信令消息并转发
func HandleSignaling(client *common.WebSocketClient, data []byte) {
	// 解析信令消息
	var msg common.SignalingMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("解析信令消息失败: %v", err)
		sendError(client, "无效的消息格式")
		return
	}

	// 验证必要字段
	if msg.From == "" {
		sendError(client, "缺少'from'字段")
		return
	}
	if msg.Type == "" {
		sendError(client, "缺少'type'字段")
		return
	}

	// 确保客户端在房间内
	if client.RoomID == "" {
		sendError(client, "未加入房间")
		return
	}

	// 获取房间
	roomInstance, exists := room.GlobalRoomManager.GetRoom(client.RoomID)
	if !exists {
		sendError(client, "房间不存在")
		return
	}

	// 根据目标用户转发消息
	if msg.To != "" {
		// 点对点转发
		forwardToUser(roomInstance, msg, client)
	} else {
		// 广播给房间内所有其他用户
		broadcastToRoom(roomInstance, msg, client.UserID)
	}
}

// 点对点转发给指定用户
func forwardToUser(room *room.Room, msg common.SignalingMessage, sender *common.WebSocketClient) {
	room.Mu.RLock()
	defer room.Mu.RUnlock()

	// 检查目标用户是否在房间内
	targetClient, exists := room.Clients[msg.To]
	if !exists {
		log.Printf("用户 %s 不在房间 %s 中", msg.To, room.ID)
		sendError(sender, "目标用户不在房间内")
		return
	}

	// 转发消息
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("序列化信令消息失败: %v", err)
		sendError(sender, "处理消息失败")
		return
	}

	targetClient.SendMessage(data) // 使用正确的SendMessage方法
	log.Printf("已转发 %s 从 %s 到 %s (房间 %s)", msg.Type, msg.From, msg.To, room.ID)
}

// 广播给房间内所有其他用户
func broadcastToRoom(room *room.Room, msg common.SignalingMessage, senderID string) {
	room.Mu.RLock()
	defer room.Mu.RUnlock()

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("序列化广播消息失败: %v", err)
		return
	}

	// 发送给房间内所有非发送者用户
	for userID, client := range room.Clients {
		if userID != senderID {
			client.SendMessage(data) // 使用正确的SendMessage方法
		}
	}

	log.Printf("已广播 %s 从 %s 到房间 %s (接收者数量: %d)", 
		msg.Type, msg.From, room.ID, len(room.Clients)-1)
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
	client.SendMessage(data) // 使用正确的SendMessage方法
}
