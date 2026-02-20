package config

import (
	"log"
	"os"
	"strconv"
)

// BaseConfig 基础服务配置
type BaseConfig struct {
	Env           string
	DBDSN         string
	ServerPort    string
	LokiURL       string
	LokiUserID    string
	LokiToken     string
	PromRemoteURL string
	PromUserID    string
	PromToken     string

	// 日志配置
	LogLevel string

	// Redis 配置
	RedisHost     string
	RedisPort     int
	RedisPassword string
	RedisDB       int

	// JWT 配置
	JWTPrivateKey string
	JWTPublicKey  string
	JWTIssuer     string
}

// GlobalConfig 全局配置实例 (可选使用单例模式)
var GlobalConfig *BaseConfig

// LoadBaseConfig 加载通用基础配置
func LoadBaseConfig() *BaseConfig {
	// API环境， staging、pre、prod
	env := MustGetEnv("ENV")
	// 数据库连接字符串， 用于指定数据库连接信息, 如 `dbs/staging/user_staging.db`
	dsn := MustGetEnv("DB_DSN")
	// 服务端口， 用于指定服务运行的端口, 如 `8081`
	port := MustGetEnv("SERVER_PORT")

	// Loki/Prometheus 配置
	// 统合 Token：优先从 GRAFANA_TOKEN 获取，实现一键配置
	grafanaToken := os.Getenv("GRAFANA_TOKEN")

	// Loki 相关
	lokiURL := GetEnv("LOKI_URL", "https://logs-prod-021.grafana.net/loki/api/v1/push")
	lokiUserID := os.Getenv("LOKI_USER_ID")
	lokiToken := GetEnv("LOKI_TOKEN", grafanaToken)

	// Prometheus 相关
	promRemoteURL := os.Getenv("PROM_REMOTE_URL")
	promUserID := os.Getenv("PROM_USER_ID")
	promToken := GetEnv("PROM_TOKEN", grafanaToken)

	// 日志级别逻辑：开发环境默认 debug，其他默认 info。支持环境变量覆盖。
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		if env == "development" {
			logLevel = "debug"
		} else {
			logLevel = "info"
		}
	}

	GlobalConfig = &BaseConfig{
		Env:           env,
		DBDSN:         dsn,
		ServerPort:    port,
		LokiURL:       lokiURL,
		LokiUserID:    lokiUserID,
		LokiToken:     lokiToken,
		PromRemoteURL: promRemoteURL,
		PromUserID:    promUserID,
		PromToken:     promToken,
		LogLevel:      logLevel,

		// Redis
		RedisHost:     GetEnv("REDIS_HOST", "localhost"),
		RedisPort:     GetEnvAsInt("REDIS_PORT", 6379),
		RedisPassword: GetEnv("REDIS_PASSWORD", ""),
		RedisDB:       GetEnvAsInt("REDIS_DB", 0),

		// JWT
		JWTPrivateKey: GetEnv("JWT_PRIVATE_KEY", ""),
		JWTPublicKey:  GetEnv("JWT_PUBLIC_KEY", ""),
		JWTIssuer:     GetEnv("JWT_ISSUER", "we-circle-prod"),
	}

	return GlobalConfig
}

// GetEnv 获取环境变量，如果不存在则返回默认值
func GetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// MustGetEnv 获取环境变量，如果不存在则 Fatal
func MustGetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("Environment variable %s is not set", key)
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
