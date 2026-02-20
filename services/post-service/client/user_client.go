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

// UserClient 用户服务客户端（异步）
type UserClient struct {
	httpClient *client.HTTPClient
}

// NewUserClient 创建用户服务客户端
func NewUserClient() *UserClient {
	baseURL := config.GetEnv("USER_SERVICE_URL", "http://localhost:8081")

	return &UserClient{
		httpClient: client.NewHTTPClient(baseURL),
	}
}

// GetUserInfo 异步获取用户信息，通过channel返回结果（默认5秒超时）
// headers参数用于传递请求头信息，如Authorization
func (c *UserClient) GetUserInfo(ctx context.Context, userID uint, headers ...map[string]string) (<-chan *dto.UserInfo, <-chan error) {
	// 如果没有设置超时，默认5秒超时
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
	}

	resultChan := make(chan *dto.UserInfo, 1)
	errChan := make(chan error, 1)

	go func() {
		defer close(resultChan)
		defer close(errChan)

		path := "/api/v1/users/get_by_id"
		requestData := map[string]interface{}{
			"userId": userID,
		}

		var response dto.UserInfoResponse

		if err := c.httpClient.PostJSONWithContext(ctx, path, requestData, &response, headers...); err != nil {
			errChan <- err
			return
		}

		if response.Code != errs.SUCCESS {
			errChan <- fmt.Errorf("failed to get user info: %s", response.Message)
			return
		}

		resultChan <- &response.Data
	}()

	return resultChan, errChan
}

// BatchGetUserInfo 异步批量获取用户信息，通过channel返回结果（默认10秒超时）
// headers参数用于传递请求头信息，如Authorization
func (c *UserClient) BatchGetUserInfo(ctx context.Context, userIDs []uint, headers ...map[string]string) (<-chan map[uint]dto.UserInfo, <-chan error) {
	// 如果没有设置超时，默认10秒超时（批量请求可能需要更长时间）
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	resultChan := make(chan map[uint]dto.UserInfo, 1)
	errChan := make(chan error, 1)

	go func() {
		defer close(resultChan)
		defer close(errChan)

		path := "/api/v1/users/batch"
		requestData := map[string]interface{}{
			"userIds": userIDs,
		}

		var response dto.UserListResponse

		if err := c.httpClient.PostJSONWithContext(ctx, path, requestData, &response, headers...); err != nil {
			errChan <- err
			return
		}

		if response.Code != errs.SUCCESS {
			errChan <- fmt.Errorf("failed to batch get user info: %s", response.Message)
			return
		}

		// 转换为map，方便查找
		userMap := make(map[uint]dto.UserInfo)
		for _, user := range response.Data {
			userMap[user.UserID] = user
		}

		resultChan <- userMap
	}()

	return resultChan, errChan
}
