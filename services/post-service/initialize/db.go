package initialize

import (
	"common/database"
	"log"
	"post-service/dao"
	"post-service/model"

	"gorm.io/gorm"
)

// InitDB 初始化数据库
func InitDB(dsn string) *gorm.DB {
	db, err := database.Connect(dsn)
	if err != nil {
		log.Fatalf("failed to connect database, dsn: %s, error: %v", dsn, err)
	}

	// 自动迁移
	if err := model.AutoMigrate(db); err != nil {
		log.Printf("warning: failed to migrate database: %v", err)
	}

	// 初始化DAO
	dao.InitDB(db)

	return db
}
