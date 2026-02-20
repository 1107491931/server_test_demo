package client

import (
	"context"
	"fmt"
	"time"

	"common/client"
	"common/config"
	"common/dto"
	"common/errs"
)

// PostClient 动态服务客户端
type PostClient struct {
	httpClient *client.HTTPClient
}

// NewPostClient 创建动态服务客户端
func NewPostClient() *PostClient {
	baseURL := config.GetEnv("POST_SERVICE_URL", "http://localhost:8082")

	return &PostClient{
		httpClient: client.NewHTTPClient(baseURL),
	}
}

// GetUserPosts 获取用户的所有动态（同步调用，默认10秒超时）
func (c *PostClient) GetUserPosts(userID uint, page, pageSize int, headers ...map[string]string) ([]dto.PostInfo, int, error) {
	return c.GetUserPostsWithContext(context.Background(), userID, page, pageSize, headers...)
}

// GetUserPostsWithContext 获取用户的所有动态（支持Context和自定义超时）
func (c *PostClient) GetUserPostsWithContext(ctx context.Context, userID uint, page, pageSize int, headers ...map[string]string) ([]dto.PostInfo, int, error) {
	path := "/api/v1/posts/get_by_user_id"

	// 构建请求体
	requestBody := map[string]interface{}{
		"userId":   userID,
		"page":     page,
		"pageSize": pageSize,
	}

	var response dto.PostListResponse

	// 如果没有设置超时，默认10秒超时
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	if err := c.httpClient.PostJSONWithContext(ctx, path, requestBody, &response, headers...); err != nil {
		return nil, 0, err
	}

	if response.Code != errs.SUCCESS {
		return nil, 0, fmt.Errorf("failed to get user posts: %s", response.Message)
	}

	return response.Data.Posts, response.Data.Total, nil
}
