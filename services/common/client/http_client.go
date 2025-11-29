package client

import (
	"context"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
)

// HTTPClient HTTP客户端（基于 resty）
type HTTPClient struct {
	client  *resty.Client
	baseURL string
}

// NewHTTPClient 创建HTTP客户端
func NewHTTPClient(baseURL string) *HTTPClient {
	client := resty.New().
		SetBaseURL(baseURL).
		SetTimeout(10 * time.Second).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json")

	return &HTTPClient{
		client:  client,
		baseURL: baseURL,
	}
}

// Get 发送GET请求，返回响应体字节数组
func (c *HTTPClient) Get(path string) ([]byte, error) {
	resp, err := c.client.R().
		SetError(&ErrorResponse{}).
		Get(path)

	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("HTTP error: %d - %s", resp.StatusCode(), resp.String())
	}

	return resp.Body(), nil
}

// Post 发送POST请求，返回响应体字节数组
func (c *HTTPClient) Post(path string, data interface{}) ([]byte, error) {
	resp, err := c.client.R().
		SetBody(data).
		SetError(&ErrorResponse{}).
		Post(path)

	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("HTTP error: %d - %s", resp.StatusCode(), resp.String())
	}

	return resp.Body(), nil
}

// GetJSON 发送GET请求并自动反序列化JSON响应到指定结构体（同步）
func (c *HTTPClient) GetJSON(path string, result interface{}) error {
	return c.GetJSONWithContext(context.Background(), path, result)
}

// PostJSON 发送POST请求并自动反序列化JSON响应到指定结构体（同步）
func (c *HTTPClient) PostJSON(path string, data interface{}, result interface{}) error {
	return c.PostJSONWithContext(context.Background(), path, data, result)
}

// GetJSONWithContext 发送GET请求并自动反序列化JSON响应（支持Context和超时控制）
func (c *HTTPClient) GetJSONWithContext(ctx context.Context, path string, result interface{}) error {
	resp, err := c.client.R().
		SetContext(ctx).
		SetResult(result).
		SetError(&ErrorResponse{}).
		Get(path)

	if err != nil {
		// 检查是否是超时错误
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("request timeout: %w", err)
		}
		if ctx.Err() == context.Canceled {
			return fmt.Errorf("request canceled: %w", err)
		}
		return fmt.Errorf("request failed: %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("HTTP error: %d - %s", resp.StatusCode(), resp.String())
	}

	return nil
}

// PostJSONWithContext 发送POST请求并自动反序列化JSON响应（支持Context和超时控制）
func (c *HTTPClient) PostJSONWithContext(ctx context.Context, path string, data interface{}, result interface{}) error {
	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(data).
		SetResult(result).
		SetError(&ErrorResponse{}).
		Post(path)

	if err != nil {
		// 检查是否是超时错误
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("request timeout: %w", err)
		}
		if ctx.Err() == context.Canceled {
			return fmt.Errorf("request canceled: %w", err)
		}
		return fmt.Errorf("request failed: %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("HTTP error: %d - %s", resp.StatusCode(), resp.String())
	}

	return nil
}

// GetJSONAsync 异步发送GET请求，通过channel返回结果
func (c *HTTPClient) GetJSONAsync(ctx context.Context, path string, result interface{}) <-chan error {
	errChan := make(chan error, 1)
	go func() {
		defer close(errChan)
		errChan <- c.GetJSONWithContext(ctx, path, result)
	}()
	return errChan
}

// PostJSONAsync 异步发送POST请求，通过channel返回结果
func (c *HTTPClient) PostJSONAsync(ctx context.Context, path string, data interface{}, result interface{}) <-chan error {
	errChan := make(chan error, 1)
	go func() {
		defer close(errChan)
		errChan <- c.PostJSONWithContext(ctx, path, data, result)
	}()
	return errChan
}

// ErrorResponse 错误响应结构
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

