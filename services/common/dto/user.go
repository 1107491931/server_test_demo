package dto

// UserInfo 用户信息DTO（用于服务间传输）
type UserInfo struct {
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	CreatedAt string `json:"created_at"`
}

