package models

// 쪽지 목록용 정보
type ChatItem struct {
	Sender    UserBasicInfo `json:"sender"`
	Uid       uint          `json:"uid"`
	Message   string        `json:"message"`
	Timestamp uint64        `json:"timestamp"`
}

// 쪽지 내용 보기용 정보
type ChatHistory struct {
	Uid       uint   `json:"uid"`
	UserUid   uint   `json:"userUid"`
	Message   string `json:"message"`
	Timestamp uint64 `json:"timestamp"`
}
