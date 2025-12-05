package dto

// UserInfo 用户信息DTO（用于服务间传输）
type UserInfo struct {
	UserID    uint   `json:"userId"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Avatar    string `json:"avatar"`
	CreatedAt string `json:"createdAt"`
}
