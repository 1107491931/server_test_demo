package database

import (
	"fmt"
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Connect 建立数据库连接 (支持 MySQL 和 SQLite)
func Connect(dsn string) (*gorm.DB, error) {
	var dialector gorm.Dialector

	if strings.Contains(dsn, "@tcp(") {
		dialector = mysql.Open(dsn)
	} else {
		dialector = sqlite.Open(dsn)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	return db, nil
}
