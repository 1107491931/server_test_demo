package dto

// PostInfo 动态信息DTO（用于服务间传输）
type PostInfo struct {
	PostID        uint     `json:"post_id"`
	UserID        uint     `json:"user_id"`
	Content       string   `json:"content"`
	Images        []string `json:"images"`
	LikeCount     int      `json:"like_count"`
	ForwardCount  int      `json:"forward_count"`
	FavoriteCount int      `json:"favorite_count"`
	CreatedAt     string   `json:"created_at"`
}

