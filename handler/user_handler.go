package handler

import (
	"net/http"
	"server_test_demo/db_handler"
	"server_test_demo/model"

	"github.com/gin-gonic/gin"
)

// RegisterRequest 注册请求参数
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginRequest 登录请求参数
type LoginRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// @Summary      用户注册
// @Description  注册新用户
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request body RegisterRequest true "注册信息"
// @Success      200  {object}  map[string]interface{}
// @Router       /register [post]
func Register(c *gin.Context) {
	// 入参合法性检测
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查手机号是否已存在
	existingUser, err := db_handler.GetUserByPhone(req.Phone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if existingUser != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Phone number already registered"})
		return
	}

	// 创建用户
	user := model.User{
		Username: req.Username,
		Phone:    req.Phone,
		Password: req.Password, // 实际项目中应该加密存储
	}

	// 创建用户
	if err := db_handler.CreateUser(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User registered successfully", "user": user})
}

// @Summary      用户登录
// @Description  用户登录
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request body LoginRequest true "登录信息"
// @Success      200  {object}  map[string]interface{}
// @Router       /login [post]
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := db_handler.GetUserByPhone(req.Phone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if user == nil || user.Password != req.Password { // 实际项目中应该比对加密后的密码
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid phone or password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Login successful", "user": user})
}

// @Summary      获取用户信息
// @Description  根据手机号获取用户信息
// @Tags         users
// @Produce      json
// @Param        phone path string true "手机号"
// @Success      200  {object}  model.User
// @Router       /users/{phone} [get]
func GetUserByPhone(c *gin.Context) {
	phone := c.Param("phone")
	user, err := db_handler.GetUserByPhone(phone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

// @Summary      获取所有用户
// @Description  获取所有用户信息
// @Tags         users
// @Produce      json
// @Success      200  {array}  model.User
// @Router       /users [get]
func GetAllUsers(c *gin.Context) {
	users, err := db_handler.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	c.JSON(http.StatusOK, users)
}
