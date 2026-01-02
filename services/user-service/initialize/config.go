package initialize

import (
	"common/config"
	"os"
)

// Config 扩展基础配置
type Config struct {
	*config.BaseConfig
	// 在这里添加 user-service 特有的配置
	PostServiceURL string
}

// LoadConfig 加载 user-service 配置
func LoadConfig() *Config {
	base := config.LoadBaseConfig()

	return &Config{
		BaseConfig:     base,
		PostServiceURL: config.GetEnv("POST_SERVICE_URL", "http://localhost:8082"),
	}
}

// 获取私钥（示例：特定业务逻辑可以留在这里，或者也移入 common）
func (c *Config) GetPrivateKey() string {
	return os.Getenv("JWT_PRIVATE_KEY")
}
