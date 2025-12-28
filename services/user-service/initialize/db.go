package initialize

import (
	"log"
	"strings"
	"user-service/dao"
	"user-service/model"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// InitDB 初始化数据库
func InitDB(dsn string) *gorm.DB {
	var dialector gorm.Dialector

	if strings.Contains(dsn, "@tcp(") {
		// MySQL DSN 特征
		dialector = mysql.Open(dsn)
	} else {
		// 默认 SQLite
		dialector = sqlite.Open(dsn)
	}

	// 连接数据库
	db, err := gorm.Open(dialector, &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true, // 禁用外键约束
	})
	if err != nil {
		log.Printf("failed to connect database, dsn: %s, error: %v", dsn, err)
		log.Fatal("failed to connect database")
	}

	// 自动迁移
	if err := model.AutoMigrate(db); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	// 初始化DAO
	dao.InitDB(db)

	return db
}
