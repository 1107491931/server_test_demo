package handler

import (
	"context"
	"fmt"
	"net/http"
	"post-service/client"
	"post-service/dao"
	"post-service/model"
	"time"

	"common/dto"
	"common/errs"
	"common/middleware"
	"common/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreatePostRequest 创建动态请求参数
type CreatePostRequest struct {
	Content string `json:"content" binding:"required"`
	Image   string `json:"image"`
}

// GetPostByIDRequest 根据ID获取动态请求参数
type GetPostByIDRequest struct {
	PostID uint `json:"postId" binding:"required"`
}

// GetPostsByUserIDRequest 根据用户ID获取动态列表请求参数
type GetPostsByUserIDRequest struct {
	// UserID   uint `json:"userId"` // 不再需要入参传递，从Token获取
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}

// GetAllPostsRequest 获取所有动态请求参数
type GetAllPostsRequest struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}

// GetUserByPostIDRequest 根据动态ID获取用户请求参数
type GetUserByPostIDRequest struct {
	PostID uint `json:"postId" binding:"required"`
}

// PostResponse 动态响应数据
type PostResponse struct {
	PostID        uint     `json:"postId"`
	UserID        uint     `json:"userId"`
	Username      string   `json:"username"`
	Avatar        string   `json:"avatar"`
	Content       string   `json:"content"`
	Images        []string `json:"images"`
	LikeCount     int      `json:"likeCount"`
	DislikeCount  int      `json:"dislikeCount"` // 踩数
	FavoriteCount int      `json:"favoriteCount"`
	ShareCount    int      `json:"shareCount"`
	IsLiked       bool     `json:"isLiked"`     // 当前用户是否点赞
	IsDisliked    bool     `json:"isDisliked"`  // 当前用户是否踩
	IsFavorited   bool     `json:"isFavorited"` // 当前用户是否收藏
	CreatedAt     string   `json:"createdAt"`
}

// toPostResponse 将Post模型转换为PostResponse
func toPostResponse(post *model.Post, userID uint) PostResponse {
	// 检查用户互动状态
	isLiked, _ := dao.CheckUserLikedPost(userID, post.ID)
	isDisliked, _ := dao.CheckUserDislikedPost(userID, post.ID)
	isFavorited, _ := dao.CheckUserFavoritedPost(userID, post.ID)

	return PostResponse{
		PostID:        post.ID,
		UserID:        post.UserID,
		Username:      post.Username,
		Avatar:        post.Avatar,
		Content:       post.Content,
		Images:        post.Images,
		LikeCount:     post.LikeCount,
		DislikeCount:  post.DislikeCount,
		FavoriteCount: post.FavoriteCount,
		ShareCount:    post.ShareCount,
		IsLiked:       isLiked,
		IsDisliked:    isDisliked,
		IsFavorited:   isFavorited,
		CreatedAt:     post.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// CreatePost 发布动态
// @Summary      发布动态
// @Description  发布新动态（文字+图片）
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        request body CreatePostRequest true "动态信息"
// @Success      200  {object}  utils.Response
// @Router       /api/v1/posts [post]
func CreatePost(c *gin.Context) {
	var req CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	// 从Context中获取用户ID
	userID, exists := middleware.GetUserID(c)
	if !exists {
		utils.Error(c, http.StatusOK, errs.UNAUTHORIZED, "User ID not found in token")
		return
	}

	// 调用用户服务获取用户信息
	// 创建一个独立的 context，避免请求 context 被取消导致调用失败
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 从请求头中获取Authorization信息
	authHeader := c.GetHeader("Authorization")
	headers := map[string]string{}
	if authHeader != "" {
		headers["Authorization"] = authHeader
	}

	userClient := client.NewUserClient()
	userInfoChan, errChan := userClient.GetUserInfo(ctx, userID, headers)

	var userInfo *dto.UserInfo
	select {
	case userInfo = <-userInfoChan:
		// 成功获取用户信息
		if userInfo == nil {
			utils.NotFound(c, "User not found")
			return
		}
	case err := <-errChan:
		// 如果获取用户信息失败
		utils.InternalServerError(c, fmt.Sprintf("Failed to get user info: %v", err))
		return
	case <-ctx.Done():
		// 超时或取消
		utils.InternalServerError(c, fmt.Sprintf("Request timeout: %v", ctx.Err()))
		return
	}

	// 处理图片：将单张图片转换为StringArray
	images := model.StringArray{}
	if req.Image != "" {
		images = append(images, req.Image)
	}

	// 创建动态
	post := model.Post{
		UserID:        userID,
		Username:      userInfo.Username,
		Avatar:        userInfo.Avatar,
		Content:       req.Content,
		Images:        images,
		LikeCount:     0,
		FavoriteCount: 0,
		ShareCount:    0,
	}

	if err := dao.CreatePost(&post); err != nil {
		utils.InternalServerError(c, "Failed to create post")
		return
	}

	utils.SuccessWithMessage(c, "发布成功", gin.H{})
}

// GetPostByID 获取动态详情
// @Summary      获取动态详情
// @Description  根据动态ID获取动态详情
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        request body GetPostByIDRequest true "动态ID"
// @Success      200  {object}  utils.Response
// @Router       /api/v1/posts/get_by_id [post]
func GetPostByID(c *gin.Context) {
	var req GetPostByIDRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	// 获取动态信息
	post, err := dao.GetPostByID(req.PostID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.NotFound(c, "Post not found")
			return
		}
		utils.InternalServerError(c, "Database error")
		return
	}

	// 从Context中获取用户ID
	userID, _ := middleware.GetUserID(c)

	// 转换为响应格式（不包含用户信息，仅返回动态数据）
	response := toPostResponse(post, userID)

	utils.Success(c, response)
}

// GetPostsByUserID 获取用户的所有动态
// @Summary      获取用户的所有动态
// @Description  根据用户ID获取该用户的所有动态（分页，供服务间调用）
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        request body GetPostsByUserIDRequest true "用户ID和分页参数"
// @Success      200  {object}  utils.Response
// @Router       /api/v1/posts/get_by_user_id [post]
func GetPostsByUserID(c *gin.Context) {
	var req GetPostsByUserIDRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 忽略错误，使用默认值
	}

	// 从Context获取用户ID
	userID, exists := middleware.GetUserID(c)
	if !exists {
		utils.InternalServerError(c, "User ID not found in token")
		return
	}

	// 设置默认分页参数
	page := req.Page
	pageSize := req.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 获取动态列表
	posts, total, err := dao.GetPostsByUserID(userID, page, pageSize)
	if err != nil {
		utils.InternalServerError(c, "Database error")
		return
	}

	// 转换为响应格式
	responses := make([]PostResponse, 0)
	if posts != nil {
		responses = make([]PostResponse, 0, len(posts))
		for _, post := range posts {
			responses = append(responses, toPostResponse(&post, userID))
		}
	}

	// 组装响应数据
	result := gin.H{
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
		"posts":    responses,
	}

	utils.Success(c, result)
}

// GetAllPosts 获取所有动态
// @Summary      获取所有动态
// @Description  获取所有动态（分页）
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        request body GetAllPostsRequest true "分页参数"
// @Success      200  {object}  utils.Response
// @Router       /api/v1/posts/get_all [post]
func GetAllPosts(c *gin.Context) {
	var req GetAllPostsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 如果没有传参数，使用默认值
		req.Page = 1
		req.PageSize = 10
	}

	// 设置默认分页参数
	page := req.Page
	pageSize := req.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 获取动态列表
	posts, total, err := dao.GetAllPosts(page, pageSize)
	if err != nil {
		utils.InternalServerError(c, "Database error")
		return
	}

	// 从Context中获取用户ID
	userID, _ := middleware.GetUserID(c)

	// 转换为响应格式（不包含用户信息，仅返回动态数据）
	responses := make([]PostResponse, 0)
	if posts != nil {
		responses = make([]PostResponse, 0, len(posts))
		for _, post := range posts {
			responses = append(responses, toPostResponse(&post, userID))
		}
	}

	// 组装响应数据
	result := gin.H{
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
		"posts":    responses,
	}

	utils.Success(c, result)
}

// LikePostRequest 点赞请求参数
type LikePostRequest struct {
	PostID  uint  `json:"postId" binding:"required"`
	IsLiked *bool `json:"isLiked" binding:"required"`
}

// FavoritePostRequest 收藏请求参数
type FavoritePostRequest struct {
	PostID   uint  `json:"postId" binding:"required"`
	Favorite *bool `json:"favorite" binding:"required"`
}

// SharePostRequest 分享请求参数
type SharePostRequest struct {
	PostID uint `json:"postId" binding:"required"`
}

// DislikePostRequest 踩请求参数
type DislikePostRequest struct {
	PostID     uint  `json:"postId" binding:"required"`
	IsDisliked *bool `json:"isDisliked" binding:"required"`
}

// LikePost 点赞/取消点赞动态
// @Summary      点赞/取消点赞动态
// @Description  对指定动态进行点赞或取消点赞操作
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        request body LikePostRequest true "动态ID和点赞状态"
// @Success      200  {object}  utils.Response
// @Router       /api/v1/posts/like [post]
func LikePost(c *gin.Context) {
	var req LikePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	// 检查动态是否存在
	post, err := dao.GetPostByID(req.PostID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.NotFound(c, "Post not found")
			return
		}
		utils.InternalServerError(c, "Database error")
		return
	}

	// 从Context中获取用户ID
	userID, _ := middleware.GetUserID(c)

	// 检查用户当前的点赞状态
	currentLiked, err := dao.CheckUserLikedPost(userID, req.PostID)
	if err != nil {
		utils.InternalServerError(c, "Failed to check like status")
		return
	}

	// 如果请求的点赞状态与当前状态相同，直接返回成功
	if *req.IsLiked == currentLiked {
		status := "点赞"
		if !*req.IsLiked {
			status = "取消点赞"
		}
		utils.SuccessWithMessage(c, "已经"+status+"过该动态", gin.H{"postId": req.PostID})
		return
	}

	// 根据请求的点赞状态执行相应的操作
	if *req.IsLiked {
		// 创建点赞记录
		like := model.Like{
			UserID: userID,
			PostID: req.PostID,
		}
		if err := dao.CreateLike(&like); err != nil {
			utils.InternalServerError(c, "Failed to create like record")
			return
		}

		// 增加点赞数
		if err := dao.IncrementLikeCount(req.PostID); err != nil {
			utils.InternalServerError(c, "Failed to like post")
			return
		}
	} else {
		// 删除点赞记录
		if err := dao.DeleteLike(userID, req.PostID); err != nil {
			utils.InternalServerError(c, "Failed to delete like record")
			return
		}

		// 减少点赞数
		if err := dao.DecrementLikeCount(req.PostID); err != nil {
			utils.InternalServerError(c, "Failed to unlike post")
			return
		}
	}

	// 重新获取动态信息
	post, _ = dao.GetPostByID(req.PostID)

	result := gin.H{
		"postId":    post.ID,
		"likeCount": post.LikeCount,
	}

	// 返回相应的成功消息
	if *req.IsLiked {
		utils.SuccessWithMessage(c, "点赞成功", result)
	} else {
		utils.SuccessWithMessage(c, "取消点赞成功", result)
	}
}

// FavoritePost 收藏/取消收藏动态
// @Summary      收藏/取消收藏动态
// @Description  对指定动态进行收藏或取消收藏操作
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        request body FavoritePostRequest true "动态ID和收藏状态"
// @Success      200  {object}  utils.Response
// @Router       /api/v1/posts/favorite [post]
func FavoritePost(c *gin.Context) {
	var req FavoritePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	// 检查动态是否存在
	post, err := dao.GetPostByID(req.PostID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.NotFound(c, "Post not found")
			return
		}
		utils.InternalServerError(c, "Database error")
		return
	}

	// 从Context中获取用户ID
	userID, _ := middleware.GetUserID(c)

	// 检查用户当前的收藏状态
	currentFavorited, err := dao.CheckUserFavoritedPost(userID, req.PostID)
	if err != nil {
		utils.InternalServerError(c, "Failed to check favorite status")
		return
	}

	// 如果请求的收藏状态与当前状态相同，直接返回成功
	if *req.Favorite == currentFavorited {
		status := "收藏"
		if !*req.Favorite {
			status = "取消收藏"
		}
		utils.SuccessWithMessage(c, "已经"+status+"过该动态", gin.H{"postId": req.PostID})
		return
	}

	// 根据请求的收藏状态执行相应的操作
	if *req.Favorite {
		// 创建收藏记录
		favorite := model.Favorite{
			UserID: userID,
			PostID: req.PostID,
		}
		if err := dao.CreateFavorite(&favorite); err != nil {
			utils.InternalServerError(c, "Failed to create favorite record")
			return
		}

		// 增加收藏数
		if err := dao.IncrementFavoriteCount(req.PostID); err != nil {
			utils.InternalServerError(c, "Failed to favorite post")
			return
		}
	} else {
		// 删除收藏记录
		if err := dao.DeleteFavorite(userID, req.PostID); err != nil {
			utils.InternalServerError(c, "Failed to delete favorite record")
			return
		}

		// 减少收藏数
		if err := dao.DecrementFavoriteCount(req.PostID); err != nil {
			utils.InternalServerError(c, "Failed to unfavorite post")
			return
		}
	}

	// 重新获取动态信息
	post, _ = dao.GetPostByID(req.PostID)

	result := gin.H{
		"postId":        post.ID,
		"favoriteCount": post.FavoriteCount,
	}

	// 返回相应的成功消息
	if *req.Favorite {
		utils.SuccessWithMessage(c, "收藏成功", result)
	} else {
		utils.SuccessWithMessage(c, "取消收藏成功", result)
	}
}

// SharePost 分享动态
// @Summary      分享动态
// @Description  对指定动态进行分享操作
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        request body SharePostRequest true "动态ID"
// @Success      200  {object}  utils.Response
// @Router       /api/v1/posts/share [post]
func SharePost(c *gin.Context) {
	var req SharePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	// 检查动态是否存在
	post, err := dao.GetPostByID(req.PostID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.NotFound(c, "Post not found")
			return
		}
		utils.InternalServerError(c, "Database error")
		return
	}

	// 增加分享数
	if err := dao.IncrementShareCount(req.PostID); err != nil {
		utils.InternalServerError(c, "Failed to share post")
		return
	}

	// 重新获取动态信息
	post, _ = dao.GetPostByID(req.PostID)

	result := gin.H{
		"postId":     post.ID,
		"shareCount": post.ShareCount,
	}

	utils.SuccessWithMessage(c, "分享成功", result)
}

// DislikePost 踩/取消踩动态
// @Summary      踩/取消踩动态
// @Description  对指定动态进行踩或取消踩操作
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        request body DislikePostRequest true "动态ID和踩状态"
// @Success      200  {object}  utils.Response
// @Router       /api/v1/posts/dislike [post]
func DislikePost(c *gin.Context) {
	var req DislikePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	// 检查动态是否存在
	post, err := dao.GetPostByID(req.PostID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.NotFound(c, "Post not found")
			return
		}
		utils.InternalServerError(c, "Database error")
		return
	}

	// 从Context中获取用户ID
	userID, _ := middleware.GetUserID(c)

	// 检查用户当前的踩状态
	currentDisliked, err := dao.CheckUserDislikedPost(userID, req.PostID)
	if err != nil {
		utils.InternalServerError(c, "Failed to check dislike status")
		return
	}

	// 如果请求的踩状态与当前状态相同，直接返回成功
	if *req.IsDisliked == currentDisliked {
		status := "踩"
		if !*req.IsDisliked {
			status = "取消踩"
		}
		utils.SuccessWithMessage(c, "已经"+status+"过该动态", gin.H{"postId": req.PostID})
		return
	}

	// 根据请求的踩状态执行相应的操作
	if *req.IsDisliked {
		// 创建踩记录
		dislike := model.Dislike{
			UserID: userID,
			PostID: req.PostID,
		}
		if err := dao.CreateDislike(&dislike); err != nil {
			utils.InternalServerError(c, "Failed to create dislike record")
			return
		}

		// 增加踩数
		if err := dao.IncrementDislikeCount(req.PostID); err != nil {
			utils.InternalServerError(c, "Failed to dislike post")
			return
		}
	} else {
		// 删除踩记录
		if err := dao.DeleteDislike(userID, req.PostID); err != nil {
			utils.InternalServerError(c, "Failed to delete dislike record")
			return
		}

		// 减少踩数
		if err := dao.DecrementDislikeCount(req.PostID); err != nil {
			utils.InternalServerError(c, "Failed to undislike post")
			return
		}
	}

	// 重新获取动态信息
	post, _ = dao.GetPostByID(req.PostID)

	result := gin.H{
		"postId":       post.ID,
		"dislikeCount": post.DislikeCount,
	}

	// 返回相应的成功消息
	if *req.IsDisliked {
		utils.SuccessWithMessage(c, "踩成功", result)
	} else {
		utils.SuccessWithMessage(c, "取消踩成功", result)
	}
}

// GetUserByPostID 根据动态ID获取用户信息
// @Summary      根据动态ID获取用户信息
// @Description  根据动态ID获取该动态发布者的用户信息
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        request body GetUserByPostIDRequest true "动态ID"
// @Success      200  {object}  utils.Response
// @Router       /api/v1/posts/get_user_by_post_id [post]
func GetUserByPostID(c *gin.Context) {
	var req GetUserByPostIDRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	// 获取动态信息
	post, err := dao.GetPostByID(req.PostID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.NotFound(c, "Post not found")
			return
		}
		utils.InternalServerError(c, "Database error")
		return
	}

	// 调用用户服务获取用户信息
	// 创建一个独立的 context，避免请求 context 被取消导致调用失败
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 从请求头中获取Authorization信息
	authHeader := c.GetHeader("Authorization")
	headers := map[string]string{}
	if authHeader != "" {
		headers["Authorization"] = authHeader
	}

	userClient := client.NewUserClient()
	userInfoChan, errChan := userClient.GetUserInfo(ctx, post.UserID, headers)

	var userInfo *dto.UserInfo
	select {
	case userInfo = <-userInfoChan:
		// 成功获取用户信息
		if userInfo == nil {
			utils.NotFound(c, "User not found")
			return
		}
	case err := <-errChan:
		// 如果获取用户信息失败
		utils.InternalServerError(c, fmt.Sprintf("Failed to get user info: %v", err))
		return
	case <-ctx.Done():
		// 超时或取消
		utils.InternalServerError(c, fmt.Sprintf("Request timeout: %v", ctx.Err()))
		return
	}

	// 返回用户信息
	utils.Success(c, userInfo)
}
