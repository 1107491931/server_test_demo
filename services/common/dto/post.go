package dto

// PostInfo 动态信息DTO（用于服务间传输）
type PostInfo struct {
	PostID        uint     `json:"postId"`
	UserID        uint     `json:"userId"`
	Content       string   `json:"content"`
	Images        []string `json:"images"`
	LikeCount     int      `json:"likeCount"`
	ForwardCount  int      `json:"forwardCount"`
	FavoriteCount int      `json:"favoriteCount"`
	CreatedAt     string   `json:"createdAt"`
}
