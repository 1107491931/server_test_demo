package initialize

import (
	"log"
	"user-service/dao"
	"user-service/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// InitDB 初始化数据库
func InitDB(dsn string) *gorm.DB {
	// 连接数据库
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database")
	}

	// 自动迁移
	if err := model.AutoMigrate(db); err != nil {
		log.Fatal("failed to migrate database")
	}

	// 初始化DAO
	dao.InitDB(db)

	return db
}
