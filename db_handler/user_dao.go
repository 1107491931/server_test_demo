package db_handler

import (
	"errors"
	"log"
	"os"
	"server_test_demo/model"

	"gorm.io/gorm"
)

var dbLogger *log.Logger

func init() {
	dbLogger = log.New(os.Stdout, "[DATABASE]", log.LstdFlags)
}

// GetUserByPhone 根据手机号获取用户信息
func GetUserByPhone(phone string) (*model.User, error) {
	var user model.User
	if err := DB.Where("phone = ?", phone).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			dbLogger.Printf("User not found %s\n", phone)
			return nil, nil // 用户不存在
		}
		dbLogger.Printf("GetUserByPhone error: %v\n", err)
		return nil, err
	}
	dbLogger.Printf("GetUserByPhone success: %v\n", user)
	return &user, nil
}

// GetAllUsers 获取所有用户信息
func GetAllUsers() ([]model.User, error) {
	var users []model.User
	if err := DB.Find(&users).Error; err != nil {
		dbLogger.Printf("GetAllUsers error: %v\n", err)
		return nil, err
	}
	dbLogger.Printf("GetAllUsers success: %v\n", users)
	return users, nil
}

// CreateUser 创建用户
func CreateUser(user *model.User) error {
	if err := DB.Create(user).Error; err != nil {
		dbLogger.Printf("CreateUser error: %v\n", err)
		return err
	}
	dbLogger.Printf("CreateUser success: %v\n", user)
	return nil
}
