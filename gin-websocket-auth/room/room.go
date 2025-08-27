package room

import (
	"sync"
	"gin-websocket-auth/common"
)

// Room 视频通话房间
type Room struct {
	ID      string                          // 房间ID
	Clients map[string]common.ClientInterface // 房间内的客户端（userID -> Client）
	Mu      sync.RWMutex                    // 读写锁（改为大写导出）
}

// RoomManager 房间管理器
type RoomManager struct {
	Rooms map[string]*Room                  // 所有房间（roomID -> Room）
	Mu    sync.RWMutex                      // 读写锁（改为大写导出）
}

// 全局房间管理器实例
var GlobalRoomManager = &RoomManager{
	Rooms: make(map[string]*Room),
}

// CreateRoom 创建新房间
func (rm *RoomManager) CreateRoom(roomID string) *Room {
	rm.Mu.Lock()
	defer rm.Mu.Unlock()
	
	room := &Room{
		ID:      roomID,
		Clients: make(map[string]common.ClientInterface),
	}
	rm.Rooms[roomID] = room
	return room
}

// GetRoom 获取房间
func (rm *RoomManager) GetRoom(roomID string) (*Room, bool) {
	rm.Mu.RLock()
	defer rm.Mu.RUnlock()
	
	room, exists := rm.Rooms[roomID]
	return room, exists
}

// AddClient 添加用户到房间（最多4人）
func (r *Room) AddClient(client common.ClientInterface) bool {
	r.Mu.Lock()
	defer r.Mu.Unlock()
	
	if len(r.Clients) >= 4 {
		return false // 房间已满
	}
	r.Clients[client.GetUserID()] = client
	return true
}

// RemoveClient 从房间移除用户
func (r *Room) RemoveClient(userID string) {
	r.Mu.Lock()
	defer r.Mu.Unlock()
	
	delete(r.Clients, userID)
}

// Broadcast 向房间内所有其他用户广播消息
func (r *Room) Broadcast(senderID string, message []byte) {
	r.Mu.RLock()
	defer r.Mu.RUnlock()
	
	for userID, client := range r.Clients {
		if userID != senderID {
			client.SendMessage(message)
		}
	}
}

// GetOtherClients 获取房间内其他用户
func (r *Room) GetOtherClients(currentUserID string) []common.ClientInterface {
	r.Mu.RLock()
	defer r.Mu.RUnlock()
	
	var clients []common.ClientInterface
	for userID, client := range r.Clients {
		if userID != currentUserID {
			clients = append(clients, client)
		}
	}
	return clients
}
