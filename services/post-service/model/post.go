package model

import (
	"database/sql/driver"
	"encoding/json"

	"gorm.io/gorm"
)

// StringArray 字符串数组类型（用于存储图片URL列表）
type StringArray []string

// Scan 实现 sql.Scanner 接口
func (s *StringArray) Scan(value interface{}) error {
	if value == nil {
		*s = []string{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, s)
}

// Value 实现 driver.Valuer 接口
func (s StringArray) Value() (driver.Value, error) {
	if len(s) == 0 {
		return "[]", nil
	}
	return json.Marshal(s)
}

// Post 动态模型
type Post struct {
	gorm.Model
	UserID        uint        `gorm:"not null;index" json:"user_id"`     // 发布用户ID
	Username      string      `gorm:"type:varchar(100);not null" json:"username"`      // 用户名
	Avatar        string      `gorm:"type:varchar(500);default:''" json:"avatar"`     // 用户头像URL
	Content       string      `gorm:"type:text;not null" json:"content"` // 动态文本内容
	Images        StringArray `gorm:"type:text" json:"images"`           // 图片URL列表
	LikeCount     int         `gorm:"default:0" json:"like_count"`       // 点赞数
	DislikeCount  int         `gorm:"default:0" json:"dislike_count"`    // 踩数
	FavoriteCount int         `gorm:"default:0" json:"favorite_count"`   // 收藏数
	ShareCount    int         `gorm:"default:0" json:"share_count"`      // 分享数
}

// Like 点赞模型
type Like struct {
	gorm.Model
	UserID uint `gorm:"not null;index" json:"user_id"` // 用户ID
	PostID uint `gorm:"not null;index" json:"post_id"` // 动态ID
}

// Dislike 踩模型
type Dislike struct {
	gorm.Model
	UserID uint `gorm:"not null;index" json:"user_id"` // 用户ID
	PostID uint `gorm:"not null;index" json:"post_id"` // 动态ID
}

// Favorite 收藏模型
type Favorite struct {
	gorm.Model
	UserID uint `gorm:"not null;index" json:"user_id"` // 用户ID
	PostID uint `gorm:"not null;index" json:"post_id"` // 动态ID
}

// AutoMigrate 自动迁移数据库表结构
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&Post{}, &Like{}, &Dislike{}, &Favorite{})
}
