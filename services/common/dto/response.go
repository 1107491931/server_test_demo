package dto

// BaseResponse 基础响应结构（服务间通信通用）
type BaseResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// UserInfoResponse 用户信息响应结构
type UserInfoResponse struct {
	BaseResponse
	Data UserInfo `json:"data"`
}

// UserListResponse 用户列表响应结构
type UserListResponse struct {
	BaseResponse
	Data []UserInfo `json:"data"`
}

// PostListData 动态列表数据
type PostListData struct {
	Total    int       `json:"total"`
	Page     int       `json:"page"`
	PageSize int       `json:"page_size"`
	Posts    []PostInfo `json:"posts"`
}

// PostListResponse 动态列表响应结构
type PostListResponse struct {
	BaseResponse
	Data PostListData `json:"data"`
}

