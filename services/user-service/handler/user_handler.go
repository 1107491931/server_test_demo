package handler

import (
	"fmt"
	"strconv"
	"user-service/client"
	"user-service/dao"
	"user-service/model"

	"common/dto"
	"common/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRequest 注册请求参数
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginRequest 登录请求参数
type LoginRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// BatchGetUsersRequest 批量获取用户请求参数
type BatchGetUsersRequest struct {
	UserIDs []uint `json:"user_ids" binding:"required"`
}

// UserResponse 用户响应数据
type UserResponse struct {
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	CreatedAt string `json:"created_at"`
}

// toUserResponse 将User模型转换为UserResponse
func toUserResponse(user *model.User) UserResponse {
	return UserResponse{
		UserID:    user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Phone:     user.Phone,
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
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	// 检查手机号是否已存在
	existingUser, err := dao.GetUserByPhone(req.Phone)
	if err != nil && err != gorm.ErrRecordNotFound {
		utils.InternalServerError(c, "Database error")
		return
	}
	if existingUser != nil {
		utils.BadRequest(c, "Phone number already registered")
		return
	}

	// 检查邮箱是否已存在
	existingUser, err = dao.GetUserByEmail(req.Email)
	if err != nil && err != gorm.ErrRecordNotFound {
		utils.InternalServerError(c, "Database error")
		return
	}
	if existingUser != nil {
		utils.BadRequest(c, "Email already registered")
		return
	}

	// 创建用户
	user := model.User{
		Username: req.Username,
		Email:    req.Email,
		Phone:    req.Phone,
		Password: req.Password, // TODO: 实际项目中应该使用bcrypt加密
	}

	if err := dao.CreateUser(&user); err != nil {
		utils.InternalServerError(c, "Failed to create user")
		return
	}

	utils.SuccessWithMessage(c, "注册成功", toUserResponse(&user))
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

	user, err := dao.GetUserByPhone(req.Phone)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.Unauthorized(c, "Invalid phone or password")
			return
		}
		utils.InternalServerError(c, "Database error")
		return
	}

	// TODO: 实际项目中应该使用bcrypt比对加密后的密码
	if user.Password != req.Password {
		utils.Unauthorized(c, "Invalid phone or password")
		return
	}

	// TODO: 生成JWT Token
	response := gin.H{
		"user":  toUserResponse(user),
		"token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...", // 示例token
	}

	utils.SuccessWithMessage(c, "登录成功", response)
}

// GetUserByID 根据用户ID获取用户信息
// @Summary      获取用户信息
// @Description  根据用户ID获取用户信息
// @Tags         users
// @Produce      json
// @Param        user_id path int true "用户ID"
// @Success      200  {object}  utils.Response
// @Router       /api/v1/users/{user_id} [get]
func GetUserByID(c *gin.Context) {
	var userID uint
	if err := c.ShouldBindUri(&struct {
		UserID uint `uri:"user_id" binding:"required"`
	}{UserID: userID}); err != nil {
		utils.BadRequest(c, "Invalid user ID")
		return
	}

	// 从URI中获取user_id
	userIDParam := c.Param("user_id")
	var id uint
	if _, err := fmt.Sscanf(userIDParam, "%d", &id); err != nil {
		utils.BadRequest(c, "Invalid user ID")
		return
	}

	user, err := dao.GetUserByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.NotFound(c, "User not found")
			return
		}
		utils.InternalServerError(c, "Database error")
		return
	}

	utils.Success(c, toUserResponse(user))
}

// GetUserByPhone 根据手机号获取用户信息
// @Summary      根据手机号获取用户信息
// @Description  根据手机号获取用户信息
// @Tags         users
// @Produce      json
// @Param        phone path string true "手机号"
// @Success      200  {object}  utils.Response
// @Router       /api/v1/users/phone/{phone} [get]
func GetUserByPhone(c *gin.Context) {
	phone := c.Param("phone")
	if phone == "" {
		utils.BadRequest(c, "Phone is required")
		return
	}

	user, err := dao.GetUserByPhone(phone)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.NotFound(c, "User not found")
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
// @Produce      json
// @Success      200  {object}  utils.Response
// @Router       /api/v1/users [get]
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
// @Produce      json
// @Param        user_id path int true "用户ID"
// @Param        page query int false "页码" default(1)
// @Param        page_size query int false "每页数量" default(10)
// @Success      200  {object}  utils.Response
// @Router       /api/v1/users/{user_id}/posts [get]
func GetUserWithPosts(c *gin.Context) {
	// 获取用户ID
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

	// 获取用户信息， 接口中也需要返回用户信息
	user, err := dao.GetUserByID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.NotFound(c, "User not found")
			return
		}
		utils.InternalServerError(c, "Database error")
		return
	}

	// 调用动态服务获取用户的所有动态
	postClient := client.NewPostClient()
	posts, total, err := postClient.GetUserPosts(userID, page, pageSize)
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
