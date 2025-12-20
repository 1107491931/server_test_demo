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

	// 自动迁移： 自动创建表、新增字段时老数据自动补充字段
	// 如果新增字段没有默认值，则数字类型默认0、bool类型默认false、对象类型默认nil等
	if err := model.AutoMigrate(db); err != nil {
		log.Fatal("failed to migrate database")
	}

	// 初始化DAO
	dao.InitDB(db)

	return db
}
