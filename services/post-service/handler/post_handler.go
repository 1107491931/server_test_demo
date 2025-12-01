package handler

import (
	"context"
	"fmt"
	"post-service/client"
	"post-service/dao"
	"post-service/model"
	"strconv"
	"time"

	"common/dto"
	"common/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreatePostRequest 创建动态请求参数
type CreatePostRequest struct {
	UserID  uint     `json:"user_id" binding:"required"`
	Content string   `json:"content" binding:"required"`
	Images  []string `json:"images"`
}


// PostResponse 动态响应数据
type PostResponse struct {
	PostID        uint     `json:"post_id"`
	UserID        uint     `json:"user_id"`
	Content       string   `json:"content"`
	Images        []string `json:"images"`
	LikeCount     int      `json:"like_count"`
	ForwardCount  int      `json:"forward_count"`
	FavoriteCount int      `json:"favorite_count"`
	CreatedAt     string   `json:"created_at"`
}

// toPostResponse 将Post模型转换为PostResponse
func toPostResponse(post *model.Post) PostResponse {
	return PostResponse{
		PostID:        post.ID,
		UserID:        post.UserID,
		Content:       post.Content,
		Images:        post.Images,
		LikeCount:     post.LikeCount,
		ForwardCount:  post.ForwardCount,
		FavoriteCount: post.FavoriteCount,
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

	// 创建动态
	post := model.Post{
		UserID:        req.UserID,
		Content:       req.Content,
		Images:        req.Images,
		LikeCount:     0,
		ForwardCount:  0,
		FavoriteCount: 0,
	}

	if err := dao.CreatePost(&post); err != nil {
		utils.InternalServerError(c, "Failed to create post")
		return
	}

	utils.SuccessWithMessage(c, "发布成功", toPostResponse(&post))
}

// GetPostByID 获取动态详情
// @Summary      获取动态详情
// @Description  根据动态ID获取动态详情
// @Tags         posts
// @Produce      json
// @Param        post_id path int true "动态ID"
// @Success      200  {object}  utils.Response
// @Router       /api/v1/posts/{post_id} [get]
func GetPostByID(c *gin.Context) {
	postIDParam := c.Param("post_id")
	var postID uint
	if _, err := fmt.Sscanf(postIDParam, "%d", &postID); err != nil {
		utils.BadRequest(c, "Invalid post ID")
		return
	}

	// 获取动态信息
	post, err := dao.GetPostByID(postID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.NotFound(c, "Post not found")
			return
		}
		utils.InternalServerError(c, "Database error")
		return
	}

	// 转换为响应格式（不包含用户信息，仅返回动态数据）
	response := toPostResponse(post)

	utils.Success(c, response)
}

// GetPostsByUserID 获取用户的所有动态
// @Summary      获取用户的所有动态
// @Description  根据用户ID获取该用户的所有动态（分页，供服务间调用）
// @Tags         posts
// @Produce      json
// @Param        user_id path int true "用户ID"
// @Param        page query int false "页码" default(1)
// @Param        page_size query int false "每页数量" default(10)
// @Success      200  {object}  utils.Response
// @Router       /api/v1/posts/user/{user_id} [get]
func GetPostsByUserID(c *gin.Context) {
	userIDParam := c.Param("user_id")
	var userID uint
	if _, err := fmt.Sscanf(userIDParam, "%d", &userID); err != nil {
		utils.BadRequest(c, "Invalid user ID")
		return
	}

	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

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
	var responses []PostResponse
	for _, post := range posts {
		responses = append(responses, toPostResponse(&post))
	}

	// 组装响应数据
	result := gin.H{
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"posts":     responses,
	}

	utils.Success(c, result)
}

// GetAllPosts 获取所有动态
// @Summary      获取所有动态
// @Description  获取所有动态（分页）
// @Tags         posts
// @Produce      json
// @Param        page query int false "页码" default(1)
// @Param        page_size query int false "每页数量" default(10)
// @Success      200  {object}  utils.Response
// @Router       /api/v1/posts [get]
func GetAllPosts(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

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

	// 转换为响应格式（不包含用户信息，仅返回动态数据）
	var responses []PostResponse
	for _, post := range posts {
		responses = append(responses, toPostResponse(&post))
	}

	// 组装响应数据
	result := gin.H{
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"posts":     responses,
	}

	utils.Success(c, result)
}

// LikePost 点赞动态
// @Summary      点赞动态
// @Description  对指定动态进行点赞操作
// @Tags         posts
// @Produce      json
// @Param        post_id path int true "动态ID"
// @Success      200  {object}  utils.Response
// @Router       /api/v1/posts/{post_id}/like [post]
func LikePost(c *gin.Context) {
	postIDParam := c.Param("post_id")
	var postID uint
	if _, err := fmt.Sscanf(postIDParam, "%d", &postID); err != nil {
		utils.BadRequest(c, "Invalid post ID")
		return
	}

	// 检查动态是否存在
	post, err := dao.GetPostByID(postID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.NotFound(c, "Post not found")
			return
		}
		utils.InternalServerError(c, "Database error")
		return
	}

	// 增加点赞数
	if err := dao.IncrementLikeCount(postID); err != nil {
		utils.InternalServerError(c, "Failed to like post")
		return
	}

	// 重新获取动态信息
	post, _ = dao.GetPostByID(postID)

	result := gin.H{
		"post_id":    post.ID,
		"like_count": post.LikeCount,
	}

	utils.SuccessWithMessage(c, "点赞成功", result)
}

// ForwardPost 转发动态
// @Summary      转发动态
// @Description  对指定动态进行转发操作
// @Tags         posts
// @Produce      json
// @Param        post_id path int true "动态ID"
// @Success      200  {object}  utils.Response
// @Router       /api/v1/posts/{post_id}/forward [post]
func ForwardPost(c *gin.Context) {
	postIDParam := c.Param("post_id")
	var postID uint
	if _, err := fmt.Sscanf(postIDParam, "%d", &postID); err != nil {
		utils.BadRequest(c, "Invalid post ID")
		return
	}

	// 检查动态是否存在
	post, err := dao.GetPostByID(postID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.NotFound(c, "Post not found")
			return
		}
		utils.InternalServerError(c, "Database error")
		return
	}

	// 增加转发数
	if err := dao.IncrementForwardCount(postID); err != nil {
		utils.InternalServerError(c, "Failed to forward post")
		return
	}

	// 重新获取动态信息
	post, _ = dao.GetPostByID(postID)

	result := gin.H{
		"post_id":       post.ID,
		"forward_count": post.ForwardCount,
	}

	utils.SuccessWithMessage(c, "转发成功", result)
}

// FavoritePost 收藏动态
// @Summary      收藏动态
// @Description  对指定动态进行收藏操作
// @Tags         posts
// @Produce      json
// @Param        post_id path int true "动态ID"
// @Success      200  {object}  utils.Response
// @Router       /api/v1/posts/{post_id}/favorite [post]
func FavoritePost(c *gin.Context) {
	postIDParam := c.Param("post_id")
	var postID uint
	if _, err := fmt.Sscanf(postIDParam, "%d", &postID); err != nil {
		utils.BadRequest(c, "Invalid post ID")
		return
	}

	// 检查动态是否存在
	post, err := dao.GetPostByID(postID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.NotFound(c, "Post not found")
			return
		}
		utils.InternalServerError(c, "Database error")
		return
	}

	// 增加收藏数
	if err := dao.IncrementFavoriteCount(postID); err != nil {
		utils.InternalServerError(c, "Failed to favorite post")
		return
	}

	// 重新获取动态信息
	post, _ = dao.GetPostByID(postID)

	result := gin.H{
		"post_id":        post.ID,
		"favorite_count": post.FavoriteCount,
	}

	utils.SuccessWithMessage(c, "收藏成功", result)
}

// GetUserByPostID 根据动态ID获取用户信息
// @Summary      根据动态ID获取用户信息
// @Description  根据动态ID获取该动态发布者的用户信息
// @Tags         posts
// @Produce      json
// @Param        post_id path int true "动态ID"
// @Success      200  {object}  utils.Response
// @Router       /api/v1/posts/{post_id}/user [get]
func GetUserByPostID(c *gin.Context) {
	postIDParam := c.Param("post_id")
	var postID uint
	if _, err := fmt.Sscanf(postIDParam, "%d", &postID); err != nil {
		utils.BadRequest(c, "Invalid post ID")
		return
	}

	// 获取动态信息
	post, err := dao.GetPostByID(postID)
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

	userClient := client.NewUserClient()
	userInfoChan, errChan := userClient.GetUserInfo(ctx, post.UserID)

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
