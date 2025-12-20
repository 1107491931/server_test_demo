package initialize

import (
	"log"
	"os"
	"strconv"
)

// Config 服务配置
type Config struct {
	Env        string
	DBDSN      string
	ServerPort string
}

// LoadConfig 加载配置
func LoadConfig() *Config {
	// 1. 服务运行环境：环境变量仅用于main.go中打印了， 比如ENV=staging
	env := os.Getenv("ENV")
	if env == "" {
		log.Fatal("ENV is not set")
	}
     
	// 2. 服务数据库路径：
	// 获取 SQLite 数据库的文件路径，用于建立数据库连接。
	// 在运行本服务时指定环境变量示例：DB_DSN=dbs/staging/user_staging.db
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("DB_DSN is not set")
	}

	// 3. 本服务运行的端口
	// 本服务运行在哪个端口上, 8081登录、8082动态
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		log.Fatal("SERVER_PORT is not set")
	}

	return &Config{
		Env:        env,
		DBDSN:      dsn,
		ServerPort: port,
	}
}

// GetEnv 获取环境变量，如果不存在则返回默认值
func GetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// GetEnvAsInt 获取整数类型的环境变量
func GetEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}
