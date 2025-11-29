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
	Content       string      `gorm:"type:text;not null" json:"content"` // 动态文本内容
	Images        StringArray `gorm:"type:text" json:"images"`           // 图片URL列表
	LikeCount     int         `gorm:"default:0" json:"like_count"`       // 点赞数
	ForwardCount  int         `gorm:"default:0" json:"forward_count"`    // 转发数
	FavoriteCount int         `gorm:"default:0" json:"favorite_count"`   // 收藏数
}

// AutoMigrate 自动迁移数据库表结构
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&Post{})
}
