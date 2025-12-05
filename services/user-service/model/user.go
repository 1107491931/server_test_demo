package model

import (
	"gorm.io/gorm"
)

// User 用户模型
// 包含用户的基本信息：用户名、邮箱、密码、头像
type User struct {
	gorm.Model
	Username string `gorm:"type:varchar(100);not null" json:"username"`          // 用户名
	Email    string `gorm:"type:varchar(100);uniqueIndex;not null" json:"email"` // 邮箱，唯一索引
	Password string `gorm:"type:varchar(255);not null" json:"-"`                 // 密码，json序列化时忽略
	Avatar   string `gorm:"type:varchar(500);default:''" json:"avatar"`          // 头像URL
}

// AutoMigrate 自动迁移数据库表结构
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&User{})
}
