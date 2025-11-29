package client

// 示例：如何使用异步和并发调用
//
// 1. 同步调用（默认，简单直接）
//    userInfo, err := userClient.GetUserInfo(userID)
//
// 2. 带超时的同步调用
//    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
//    defer cancel()
//    userInfo, err := userClient.GetUserInfoWithContext(ctx, userID)
//
// 3. 异步调用（使用channel）
//    resultChan, errChan := userClient.GetUserInfoAsync(ctx, userID)
//    select {
//    case userInfo := <-resultChan:
//        // 处理结果
//    case err := <-errChan:
//        // 处理错误
//    case <-ctx.Done():
//        // 超时或取消
//    }
//
// 4. 并发调用多个服务
//    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//    defer cancel()
//
//    var wg sync.WaitGroup
//    var userInfo *dto.UserInfo
//    var posts []dto.PostInfo
//    var err1, err2 error
//
//    wg.Add(2)
//    go func() {
//        defer wg.Done()
//        userInfo, err1 = userClient.GetUserInfoWithContext(ctx, userID)
//    }()
//    go func() {
//        defer wg.Done()
//        posts, _, err2 = postClient.GetUserPostsWithContext(ctx, userID, 1, 10)
//    }()
//    wg.Wait()

