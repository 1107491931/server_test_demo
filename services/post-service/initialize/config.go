package initialize

import (
	"common/config"
)

// Config 扩展基础配置
type Config struct {
	*config.BaseConfig
	// 在这里添加 post-service 特有的配置
	UserServiceURL string
}

// LoadConfig 加载 post-service 配置
func LoadConfig() *Config {
	base := config.LoadBaseConfig()

	return &Config{
		BaseConfig:     base,
		UserServiceURL: config.GetEnv("USER_SERVICE_URL", "http://localhost:8081"),
	}
}
