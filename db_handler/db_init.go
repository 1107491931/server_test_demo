package db_handler

import (
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitDB 初始化数据库连接
func InitDB(database *gorm.DB) {
	DB = database
}
