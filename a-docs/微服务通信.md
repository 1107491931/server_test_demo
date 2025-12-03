# 用户模块与发布动态模块通信流程

## 概述
本文档介绍用户模块（user-service）与发布动态模块（post-service）之间的通信流程，包括两个模块之间的相互调用方式、接口定义和数据交互。

## 通信方式
两个模块之间通过RESTful API进行通信，使用HTTP协议作为通信协议。每个模块都实现了对应的客户端来调用另一个模块的API。

## 环境配置
- 用户服务默认运行在 `http://localhost:8081`
- 发布动态服务默认运行在 `http://localhost:8082`
- 服务地址可通过环境变量 `USER_SERVICE_URL` 和 `POST_SERVICE_URL` 进行配置

## 通信流程

### 1. 用户模块调用发布动态模块

**场景：根据用户ID获取用户发布的所有动态**

#### 调用流程：
1. 用户模块中的 `GetUserWithPosts` 处理器函数接收到请求
2. 创建 `PostClient` 实例
3. 调用 `PostClient.GetUserPosts()` 方法
4. 发送HTTP GET请求到发布动态模块的 `/api/v1/posts/user/{user_id}` 接口
5. 接收响应并解析返回的动态列表数据

#### 相关代码：
```go
// 用户模块中的GetUserWithPosts函数
type UserWithPostsResponse struct {
    UserResponse
    Posts []client.PostInfo `json:"posts,omitempty"`
    Total int               `json:"total,omitempty"`
}

// GetUserWithPosts 获取用户信息及其所有动态
func GetUserWithPosts(c *gin.Context) {
    // ...获取用户ID和分页参数...
    
    // 调用动态服务获取用户的所有动态
    postClient := client.NewPostClient()
    posts, total, err := postClient.GetUserPosts(userID, page, pageSize)
    
    // ...处理响应...
}
```

```go
// PostClient中的GetUserPosts方法
func (c *PostClient) GetUserPosts(userID uint, page, pageSize int) ([]PostInfo, int, error) {
    path := fmt.Sprintf("/api/v1/posts/user/%d?page=%d&page_size=%d", userID, page, pageSize)
    
    data, err := c.httpClient.Get(path)
    // ...解析响应...
    
    return response.Data.Posts, response.Data.Total, nil
}
```

### 2. 发布动态模块调用用户模块

**场景：根据动态的userId获取完整的用户信息**

#### 调用流程：
1. 发布动态模块中的 `GetPostByID` 或 `GetPostsByUserID` 或 `GetAllPosts` 处理器函数接收到请求
2. 创建 `UserClient` 实例
3. 根据需要调用：
   - `UserClient.GetUserInfo()` 方法获取单个用户信息
   - `UserClient.BatchGetUserInfo()` 方法批量获取多个用户信息
4. 发送HTTP请求到用户模块的对应接口
5. 接收响应并解析返回的用户信息

#### 相关代码：

**单个用户信息获取：**
```go
// GetPostByID 获取动态详情
func GetPostByID(c *gin.Context) {
    // ...获取动态信息...
    
    // 调用用户服务获取用户信息
    userClient := client.NewUserClient()
    userInfo, err := userClient.GetUserInfo(post.UserID)
    
    // ...处理响应...
}
```

**批量用户信息获取：**
```go
// GetAllPosts 获取所有动态
func GetAllPosts(c *gin.Context) {
    // ...获取动态列表...
    
    // 批量获取用户信息
    var userInfoMap map[uint]client.UserInfo
    if len(userIDs) > 0 {
        userClient := client.NewUserClient()
        userInfoMap, err = userClient.BatchGetUserInfo(userIDs)
    }
    
    // ...处理响应...
}
```

```go
// UserClient中的方法
func (c *UserClient) GetUserInfo(userID uint) (*UserInfo, error) {
    path := fmt.Sprintf("/api/v1/users/%d", userID)
    data, err := c.httpClient.Get(path)
    // ...解析响应...
}

func (c *UserClient) BatchGetUserInfo(userIDs []uint) (map[uint]UserInfo, error) {
    path := "/api/v1/users/batch"
    requestData := map[string]interface{}{
        "user_ids": userIDs,
    }
    data, err := c.httpClient.Post(path, requestData)
    // ...解析响应...
}
```

## 错误处理
两个模块在调用对方服务时都实现了错误处理机制：
- 当调用失败时，会记录错误日志
- 即使获取关联数据失败，仍会返回主要数据，只是关联数据为空或不包含

## 数据模型

### 动态信息结构
```go
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
```

### 用户信息结构
```go
type UserInfo struct {
    UserID    uint   `json:"user_id"`
    Username  string `json:"username"`
    Email     string `json:"email"`
    Phone     string `json:"phone"`
    CreatedAt string `json:"created_at"`
}
```

## 总结
用户模块和发布动态模块通过RESTful API实现了双向通信，使得：
1. 用户模块可以获取用户的所有动态信息
2. 发布动态模块可以获取动态对应的用户信息

这种通信方式实现了两个模块之间的数据解耦，同时又保证了数据的完整性和关联性。