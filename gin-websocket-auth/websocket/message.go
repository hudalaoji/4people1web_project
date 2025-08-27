package websocket

import "encoding/json"

// MessageType 定义WebRTC信令消息类型
type MessageType string

const (
	// 房间操作
	MsgJoinRoom    MessageType = "join_room"    // 加入房间
	MsgLeaveRoom   MessageType = "leave_room"   // 离开房间
	MsgRoomInfo    MessageType = "room_info"    // 房间信息
	MsgRoomFull    MessageType = "room_full"    // 房间已满
	MsgUserJoined  MessageType = "user_joined"  // 新用户加入
	MsgUserLeft    MessageType = "user_left"    // 用户离开

	// WebRTC信令
	MsgOffer       MessageType = "offer"        // 提供SDP
	MsgAnswer      MessageType = "answer"       // 应答SDP
	MsgIceCandidate MessageType = "ice_candidate" // ICE候选者
)

// Message 统一消息结构
type Message struct {
	Type    MessageType          `json:"type"`
	RoomID  string               `json:"room_id,omitempty"`
	From    string               `json:"from,omitempty"`    // 发送者userID
	To      string               `json:"to,omitempty"`      // 接收者userID（可选）
	Payload json.RawMessage      `json:"payload,omitempty"` // 消息内容
}

// OfferPayload SDP Offer消息内容
type OfferPayload struct {
	SDP string `json:"sdp"`
}

// AnswerPayload SDP Answer消息内容
type AnswerPayload struct {
	SDP string `json:"sdp"`
}

// IceCandidatePayload ICE候选者消息内容
type IceCandidatePayload struct {
	Candidate     string `json:"candidate"`
	SDPMLineIndex int    `json:"sdpMLineIndex"`
	SDPMid        string `json:"sdpMid"`
}
