package dao

import (
	"user-service/model"
)

// CreateUser 创建用户
func CreateUser(user *model.User) error {
	return DB.Create(user).Error
}

// GetUserByID 根据用户ID获取用户信息
func GetUserByID(userID uint) (*model.User, error) {
	var user model.User
	err := DB.Where("id = ?", userID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail 根据邮箱获取用户信息
func GetUserByEmail(email string) (*model.User, error) {
	var user model.User
	err := DB.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUsersByIDs 批量获取用户信息
func GetUsersByIDs(userIDs []uint) ([]model.User, error) {
	var users []model.User
	err := DB.Where("id IN ?", userIDs).Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

// GetAllUsers 获取所有用户
func GetAllUsers() ([]model.User, error) {
	var users []model.User
	err := DB.Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

// UpdateUser 更新用户信息
func UpdateUser(user *model.User) error {
	return DB.Save(user).Error
}

// DeleteUser 删除用户（软删除）
func DeleteUser(userID uint) error {
	return DB.Delete(&model.User{}, userID).Error
}
