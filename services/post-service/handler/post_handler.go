package handler

import (
	"fmt"
	"post-service/client"
	"post-service/dao"
	"post-service/model"
	"strconv"

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

// LikeRequest 点赞请求参数
type LikeRequest struct {
	UserID uint `json:"user_id" binding:"required"`
}

// ForwardRequest 转发请求参数
type ForwardRequest struct {
	UserID uint `json:"user_id" binding:"required"`
}

// FavoriteRequest 收藏请求参数
type FavoriteRequest struct {
	UserID uint `json:"user_id" binding:"required"`
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

// PostDetailResponse 动态详情响应数据（包含用户信息）
type PostDetailResponse struct {
	PostID        uint          `json:"post_id"`
	UserID        uint          `json:"user_id"`
	UserInfo      *dto.UserInfo `json:"user_info,omitempty"`
	Content       string        `json:"content"`
	Images        []string      `json:"images"`
	LikeCount     int           `json:"like_count"`
	ForwardCount  int           `json:"forward_count"`
	FavoriteCount int           `json:"favorite_count"`
	CreatedAt     string        `json:"created_at"`
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
// @Description  根据动态ID获取动态详情（包含用户信息）
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

	// 异步调用用户服务获取用户信息
	userClient := client.NewUserClient()
	userInfoChan, errChan := userClient.GetUserInfo(c.Request.Context(), post.UserID)

	var userInfo *dto.UserInfo
	select {
	case userInfo = <-userInfoChan:
		// 成功获取用户信息
	case err := <-errChan:
		// 如果获取用户信息失败，仍然返回动态信息，但不包含用户信息
		fmt.Printf("Failed to get user info: %v\n", err)
	case <-c.Request.Context().Done():
		// 请求被取消
		fmt.Printf("Request canceled while getting user info\n")
	}

	// 组装响应数据
	response := PostDetailResponse{
		PostID:        post.ID,
		UserID:        post.UserID,
		UserInfo:      userInfo,
		Content:       post.Content,
		Images:        post.Images,
		LikeCount:     post.LikeCount,
		ForwardCount:  post.ForwardCount,
		FavoriteCount: post.FavoriteCount,
		CreatedAt:     post.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	utils.Success(c, response)
}

// GetPostsByUserID 获取用户的所有动态
// @Summary      获取用户的所有动态
// @Description  根据用户ID获取该用户的所有动态（分页，包含用户信息）
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

	// 异步调用用户服务获取用户信息
	userClient := client.NewUserClient()
	userInfoChan, errChan := userClient.GetUserInfo(c.Request.Context(), userID)

	var userInfo *dto.UserInfo
	select {
	case userInfo = <-userInfoChan:
		// 成功获取用户信息
	case err := <-errChan:
		// 如果获取用户信息失败，记录日志但不影响返回动态列表
		fmt.Printf("Failed to get user info: %v\n", err)
	case <-c.Request.Context().Done():
		// 请求被取消
		fmt.Printf("Request canceled while getting user info\n")
	}

	// 转换为响应格式（包含用户信息）
	var responses []PostDetailResponse
	for _, post := range posts {
		response := PostDetailResponse{
			PostID:        post.ID,
			UserID:        post.UserID,
			UserInfo:      userInfo,
			Content:       post.Content,
			Images:        post.Images,
			LikeCount:     post.LikeCount,
			ForwardCount:  post.ForwardCount,
			FavoriteCount: post.FavoriteCount,
			CreatedAt:     post.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		responses = append(responses, response)
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

	// 获取所有用户ID
	userIDs := make([]uint, 0)
	userIDMap := make(map[uint]bool)
	for _, post := range posts {
		if !userIDMap[post.UserID] {
			userIDs = append(userIDs, post.UserID)
			userIDMap[post.UserID] = true
		}
	}

	// 异步批量获取用户信息
	var userInfoMap map[uint]dto.UserInfo
	if len(userIDs) > 0 {
		userClient := client.NewUserClient()
		userInfoMapChan, errChan := userClient.BatchGetUserInfo(c.Request.Context(), userIDs)

		select {
		case userInfoMap = <-userInfoMapChan:
			// 成功获取用户信息
		case err := <-errChan:
			fmt.Printf("Failed to batch get user info: %v\n", err)
		case <-c.Request.Context().Done():
			fmt.Printf("Request canceled while batch getting user info\n")
		}
	}

	// 转换为响应格式（包含用户信息）
	var responses []PostDetailResponse
	for _, post := range posts {
		response := PostDetailResponse{
			PostID:        post.ID,
			UserID:        post.UserID,
			Content:       post.Content,
			Images:        post.Images,
			LikeCount:     post.LikeCount,
			ForwardCount:  post.ForwardCount,
			FavoriteCount: post.FavoriteCount,
			CreatedAt:     post.CreatedAt.Format("2006-01-02 15:04:05"),
		}

		// 添加用户信息
		if userInfo, ok := userInfoMap[post.UserID]; ok {
			response.UserInfo = &userInfo
		}

		responses = append(responses, response)
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
// @Description  点赞动态
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        post_id path int true "动态ID"
// @Param        request body LikeRequest true "用户ID"
// @Success      200  {object}  utils.Response
// @Router       /api/v1/posts/{post_id}/like [post]
func LikePost(c *gin.Context) {
	postIDParam := c.Param("post_id")
	var postID uint
	if _, err := fmt.Sscanf(postIDParam, "%d", &postID); err != nil {
		utils.BadRequest(c, "Invalid post ID")
		return
	}

	var req LikeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
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
// @Description  转发动态
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        post_id path int true "动态ID"
// @Param        request body ForwardRequest true "用户ID"
// @Success      200  {object}  utils.Response
// @Router       /api/v1/posts/{post_id}/forward [post]
func ForwardPost(c *gin.Context) {
	postIDParam := c.Param("post_id")
	var postID uint
	if _, err := fmt.Sscanf(postIDParam, "%d", &postID); err != nil {
		utils.BadRequest(c, "Invalid post ID")
		return
	}

	var req ForwardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
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
// @Description  收藏动态
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        post_id path int true "动态ID"
// @Param        request body FavoriteRequest true "用户ID"
// @Success      200  {object}  utils.Response
// @Router       /api/v1/posts/{post_id}/favorite [post]
func FavoritePost(c *gin.Context) {
	postIDParam := c.Param("post_id")
	var postID uint
	if _, err := fmt.Sscanf(postIDParam, "%d", &postID); err != nil {
		utils.BadRequest(c, "Invalid post ID")
		return
	}

	var req FavoriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
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
