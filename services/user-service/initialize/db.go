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
	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		log.Printf("failed to connect database, dsn: %s, error: %v", dsn, err)
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
