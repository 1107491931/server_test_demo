package handler

import (
	"fmt"
	"strings"
	"user-service/client"
	"user-service/dao"
	"user-service/model"

	"common/auth"
	"common/dto"
	"common/errs"
	"common/middleware"
	"common/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// tokenManager 全局TokenManager实例
var tokenManager *auth.TokenManager

// SetTokenManager 设置TokenManager
func SetTokenManager(tm *auth.TokenManager) {
	tokenManager = tm
}

// RegisterRequest 注册请求参数
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Avatar   string `json:"avatar"`
}

// LoginRequest 登录请求参数
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RefreshTokenRequest 刷新Token请求参数
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

// LogoutRequest 登出请求参数
type LogoutRequest struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

// BatchGetUsersRequest 批量获取用户请求参数
type BatchGetUsersRequest struct {
	UserIDs []uint `json:"userIds" binding:"required"`
}

// GetUserByIDRequest 根据ID获取用户请求参数
type GetUserByIDRequest struct {
	// UserID uint `json:"userId"` // 不再需要入参传递，从Token获取
}

// GetUserByEmailRequest 根据邮箱获取用户请求参数
type GetUserByEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// GetUserWithPostsRequest 获取用户及动态请求参数
type GetUserWithPostsRequest struct {
	// UserID   uint `json:"userId"` // 不再需要入参传递，从Token获取
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}

// UserResponse 用户响应数据
type UserResponse struct {
	UserID    uint   `json:"userId"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Avatar    string `json:"avatar"`
	CreatedAt string `json:"createdAt"`
}

// toUserResponse 将User模型转换为UserResponse
func toUserResponse(user *model.User) UserResponse {
	return UserResponse{
		UserID:    user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Avatar:    user.Avatar,
		CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// Register 用户注册
// @Summary      用户注册
// @Description  注册新用户
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request body RegisterRequest true "注册信息"
// @Success      200  {object}  utils.Response
// @Router       /api/v1/users/register [post]
func Register(c *gin.Context) {
	var req RegisterRequest
	// ShouldBindJSON 尝试将请求的 JSON 正文解析到 req 结构体中
	// 错误处理：如果 JSON 格式不正确\如果缺少必填字段\如果字段格式不符合要求
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	// 检查邮箱是否已存在
	existingUser, err := dao.GetUserByEmail(req.Email)
	if err != nil && err != gorm.ErrRecordNotFound {
		utils.InternalServerError(c, "Database error")
		return
	}
	if existingUser != nil {
		utils.BusinessError(c, errs.USER_EXISTS, "Email already registered")
		return
	}

	// 创建用户
	user := model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password, // TODO: 实际项目中应该使用bcrypt加密
		Avatar:   req.Avatar,
	}

	if err := dao.CreateUser(&user); err != nil {
		utils.BusinessError(c, errs.USER_CREATE_FAIL, "Failed to create user")
		return
	}

	// 生成Token
	tokenPair, err := tokenManager.GenerateTokenPair(user.ID, user.Username, user.Email)
	if err != nil {
		utils.InternalServerError(c, "Failed to generate token")
		return
	}

	utils.SuccessWithMessage(c, "注册成功", gin.H{
		"user":         toUserResponse(&user),
		"accessToken":  tokenPair.AccessToken,
		"refreshToken": tokenPair.RefreshToken,
		"expiresIn":    tokenPair.ExpiresIn,
		"refreshIn":    tokenPair.RefreshIn,
	})
}

// Login 用户登录
// @Summary      用户登录
// @Description  用户登录
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request body LoginRequest true "登录信息"
// @Success      200  {object}  utils.Response
// @Router       /api/v1/users/login [post]
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	user, err := dao.GetUserByEmail(req.Email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.BusinessError(c, errs.PASSWORD_INCORRECT, "Invalid email or password")
			return
		}
		utils.InternalServerError(c, "Database error")
		return
	}

	// TODO: 实际项目中应该使用bcrypt比对加密后的密码
	if user.Password != req.Password {
		utils.BusinessError(c, errs.PASSWORD_INCORRECT, "Invalid email or password")
		return
	}

	// 生成Token
	tokenPair, err := tokenManager.GenerateTokenPair(user.ID, user.Username, user.Email)
	if err != nil {
		utils.InternalServerError(c, "Failed to generate token")
		return
	}

	utils.SuccessWithMessage(c, "登录成功", gin.H{
		"user":         toUserResponse(user),
		"accessToken":  tokenPair.AccessToken,
		"refreshToken": tokenPair.RefreshToken,
		"expiresIn":    tokenPair.ExpiresIn,
		"refreshIn":    tokenPair.RefreshIn,
	})
}

// GetUserByID 根据用户ID获取用户信息
// @Summary      获取用户信息
// @Description  根据用户ID获取用户信息
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request body GetUserByIDRequest true "用户ID"
// @Success      200  {object}  utils.Response
// @Router       /api/v1/users/get_by_id [post]
func GetUserByID(c *gin.Context) {
	// 从Context获取用户ID
	userID, exists := middleware.GetUserID(c)
	if !exists {
		utils.InternalServerError(c, "User ID not found in token")
		return
	}

	user, err := dao.GetUserByID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.BusinessError(c, errs.USER_NOT_FOUND, "User not found")
			return
		}
		utils.InternalServerError(c, "Database error")
		return
	}

	utils.Success(c, toUserResponse(user))
}

// GetUserByEmail 根据邮箱获取用户信息
// @Summary      根据邮箱获取用户信息
// @Description  根据邮箱获取用户信息
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request body GetUserByEmailRequest true "邮箱"
// @Success      200  {object}  utils.Response
// @Router       /api/v1/users/get_by_email [post]
func GetUserByEmail(c *gin.Context) {
	var req GetUserByEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	user, err := dao.GetUserByEmail(req.Email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.BusinessError(c, errs.USER_NOT_FOUND, "User not found")
			return
		}
		utils.InternalServerError(c, "Database error")
		return
	}

	utils.Success(c, toUserResponse(user))
}

// BatchGetUsers 批量获取用户信息
// @Summary      批量获取用户信息
// @Description  批量获取用户信息（供服务间调用）
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request body BatchGetUsersRequest true "用户ID列表"
// @Success      200  {object}  utils.Response
// @Router       /api/v1/users/batch [post]
func BatchGetUsers(c *gin.Context) {
	var req BatchGetUsersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	if len(req.UserIDs) == 0 {
		utils.BadRequest(c, "User IDs cannot be empty")
		return
	}

	users, err := dao.GetUsersByIDs(req.UserIDs)
	if err != nil {
		utils.InternalServerError(c, "Database error")
		return
	}

	// 转换为响应格式
	var responses []UserResponse
	for _, user := range users {
		responses = append(responses, toUserResponse(&user))
	}

	utils.Success(c, responses)
}

// GetAllUsers 获取所有用户
// @Summary      获取所有用户
// @Description  获取所有用户信息
// @Tags         users
// @Accept       json
// @Produce      json
// @Success      200  {object}  utils.Response
// @Router       /api/v1/users/get_all [post]
func GetAllUsers(c *gin.Context) {
	users, err := dao.GetAllUsers()
	if err != nil {
		utils.InternalServerError(c, "Database error")
		return
	}

	// 转换为响应格式
	var responses []UserResponse
	for _, user := range users {
		responses = append(responses, toUserResponse(&user))
	}

	utils.Success(c, responses)
}

// UserWithPostsResponse 用户及其动态响应数据
type UserWithPostsResponse struct {
	UserResponse
	Posts []dto.PostInfo `json:"posts"`
	Total int            `json:"total"`
}

// GetUserWithPosts 获取用户信息及其所有动态
// @Summary      获取用户信息及其所有动态
// @Description  根据用户ID获取用户信息和该用户发布的所有动态
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request body GetUserWithPostsRequest true "用户ID和分页参数"
// @Success      200  {object}  utils.Response
// @Router       /api/v1/users/get_with_posts [post]
func GetUserWithPosts(c *gin.Context) {
	var req GetUserWithPostsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 如果 body 为空或字段不存在，忽略错误继续执行（分页使用默认值）
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

	// 获取用户信息， 接口中也需要返回用户信息
	user, err := dao.GetUserByID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.BusinessError(c, errs.USER_NOT_FOUND, "User not found")
			return
		}
		utils.InternalServerError(c, "Database error")
		return
	}

	// 转发 Authorization 头
	headers := map[string]string{}
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		headers["Authorization"] = authHeader
	}

	// 调用动态服务获取用户的所有动态
	postClient := client.NewPostClient()
	posts, total, err := postClient.GetUserPosts(userID, page, pageSize, headers)
	if err != nil {
		// 如果获取动态失败，仍然返回用户信息，但不包含动态
		fmt.Printf("Failed to get user posts: %v\n", err)
		posts = []dto.PostInfo{}
		total = 0
	}

	// 组装响应数据
	response := UserWithPostsResponse{
		UserResponse: toUserResponse(user),
		Posts:        posts,
		Total:        total,
	}

	utils.Success(c, response)
}

// RefreshToken 刷新AccessToken
// @Summary      刷新AccessToken
// @Description  使用RefreshToken获取新的AccessToken和RefreshToken
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request body RefreshTokenRequest true "RefreshToken"
// @Success      200  {object}  utils.Response
// @Router       /api/v1/users/refresh [post]
func RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	// 使用RefreshToken生成新的Token对
	tokenPair, err := tokenManager.RefreshAccessToken(req.RefreshToken)
	if err != nil {
		utils.BusinessError(c, errs.TOKEN_REVOKED, "Invalid or expired refresh token")
		return
	}

	utils.SuccessWithMessage(c, "Token刷新成功", tokenPair)
}

// Logout 用户登出
// @Summary      用户登出
// @Description  撤销用户的AccessToken和RefreshToken
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request body LogoutRequest false "Token信息（可选，如不提供则从Authorization头获取accessToken）"
// @Success      200  {object}  utils.Response
// @Router       /api/v1/users/logout [post]
func Logout(c *gin.Context) {
	var req LogoutRequest
	var accessToken string

	// 尝试解析请求体（如果有的话）
	err := c.ShouldBindJSON(&req)
	if err == nil {
		// 请求体解析成功，使用请求体中的Token
		accessToken = req.AccessToken
	} else if err.Error() != "EOF && err.Error() != \"unexpected end of JSON input\"" {
		// 请求体存在但格式错误
		// 注意：gin的ShouldBindJSON在空body时可能会报EOF或unexpected end of JSON input
		if err.Error() != "EOF" {
			utils.BadRequest(c, err.Error())
			return
		}
	}

	// 如果请求体中没有提供accessToken，尝试从Authorization头获取
	if accessToken == "" {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" {
				req.AccessToken = parts[1]
			}
		}
	}

	// 如果仍然没有accessToken，返回错误
	if req.AccessToken == "" {
		utils.BadRequest(c, "Missing access token")
		return
	}

	// 获取当前用户信息（可选，用于日志记录）
	userID, _ := middleware.GetUserID(c)
	username, _ := middleware.GetUsername(c)

	// 撤销Token
	ctx := c.Request.Context()
	if err := tokenManager.Logout(ctx, req.AccessToken, req.RefreshToken); err != nil {
		utils.InternalServerError(c, "Failed to logout")
		return
	}

	// 记录登出日志
	if userID != 0 {
		fmt.Printf("User %s (ID: %d) logged out successfully\n", username, userID)
	}

	utils.SuccessWithMessage(c, "登出成功", nil)
}
