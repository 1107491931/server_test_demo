package model

import (
	"gorm.io/gorm"
)

// User 用户模型
// 包含用户的基本信息：用户名、手机号、密码
type User struct {
	gorm.Model
	Username string `gorm:"type:varchar(100);not null" json:"username"`         // 用户名
	Phone    string `gorm:"type:varchar(20);uniqueIndex;not null" json:"phone"` // 手机号，唯一索引
	Password string `gorm:"type:varchar(255);not null" json:"-"`                // 密码，json序列化时忽略
}

// AutoMigrate 自动迁移数据库表结构
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&User{})
}
