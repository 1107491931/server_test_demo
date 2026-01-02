package logger

import (
	"os"
	"path/filepath"

	// 日志上传

	// 日志打印
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	// 日志存储
	"gopkg.in/natefinch/lumberjack.v2"
)

/**
文件的作用：控制日志的输出，包括文件输出，控制台输出
*/

var Log *zap.Logger = zap.NewNop()

// Config 日志配置
type Config struct {
	Level      string // 日志级别: debug, info, warn, error
	Filename   string // 日志文件路径
	MaxSize    int    // 每个日志文件最大尺寸 (MB)
	MaxBackups int    // 保留旧日志文件的最大个数
	MaxAge     int    // 保留旧日志文件的最大天数
	Compress   bool   // 是否压缩旧日志文件
	Console    bool   // 是否在终端打印
	Loki       LokiConfig
}

// LokiConfig Loki 配置
type LokiConfig struct {
	Enabled bool
	URL     string
	UserID  string
	Token   string
	Labels  map[string]string
}

// InitLogger 初始化日志
func InitLogger(cfg Config) error {
	// 1. 设置日志级别
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		level = zapcore.InfoLevel
	}

	// 2. 创建核心(Core)
	core := zapcore.NewTee(
		getFileCore(cfg, level),
		getConsoleCore(cfg, level),
		NewLokiCore(cfg.Loki, level),
	)

	// 3. 创建 Logger
	// 添加 AddCallerSkip(1) 是为了让日志显示的行号跳过封装层，指向真正的业务代码
	Log = zap.New(core,
		zap.AddCaller(),                       // 文件名 + 行号
		zap.AddCallerSkip(1),                  // 跳过封装层
		zap.AddStacktrace(zapcore.ErrorLevel), // 仅在 error 级别时添加堆栈跟踪
	)

	// 替换全局 Logger
	zap.ReplaceGlobals(Log)

	return nil
}

// InitStandardLogger 以标准默认值初始化日志
func InitStandardLogger(serviceName string, env string, logLevel string, lokiURL, lokiUserID, lokiToken string) error {
	return InitLogger(Config{
		Level:      logLevel,                         // 日志级别: debug, info, warn, error
		Filename:   "./logs/" + serviceName + ".log", // 日志文件路径
		MaxSize:    100,                              // 每个日志文件最大尺寸 (MB)
		MaxBackups: 5,                                // 保留旧日志文件的最大个数
		MaxAge:     30,                               // 保留旧日志文件的最大天数
		Compress:   true,                             // 是否压缩旧日志文件
		Console:    true,                             // 是否在终端打印
		Loki: LokiConfig{ // Loki 配置
			Enabled: true,                                                  // 是否启用 Loki
			URL:     lokiURL,                                               // Loki URL
			UserID:  lokiUserID,                                            // Loki 用户ID
			Token:   lokiToken,                                             // Loki Token
			Labels:  map[string]string{"service": serviceName, "env": env}, // Loki 标签
		},
	})
}

// getFileCore 获取日志文件的 zapcore.Core
func getFileCore(cfg Config, level zapcore.Level) zapcore.Core {
	if cfg.Filename == "" {
		return zapcore.NewNopCore()
	}

	// 确保目录存在
	dir := filepath.Dir(cfg.Filename)
	// 创建目录， 0755 代表的是 文件或目录的访问权限控制模式，自己可以随意进入目录，查看文件，创建新日志，删除旧日志， 其他人只能查看日志
	if err := os.MkdirAll(dir, 0755); err != nil {
		panic(err)
	}

	hook := &lumberjack.Logger{
		Filename:   cfg.Filename,
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		Compress:   cfg.Compress,
	}

	encoderConfig := zap.NewProductionEncoderConfig() // 使用生产环境的编码器配置
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder // 使用 ISO8601 格式编码时间
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder // 使用大写编码级别, 如 INFO, ERROR 等

	return zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig), // 使用 JSON 编码器
		zapcore.AddSync(hook),                 // 将日志写入文件
		level,                                 // 日志级别
	)
}

// getConsoleCore 获取控制台打印的 zapcore.Core
func getConsoleCore(cfg Config, level zapcore.Level) zapcore.Core {
	if !cfg.Console {
		return zapcore.NewNopCore()
	}

	encoderConfig := zap.NewDevelopmentEncoderConfig() // 使用开发环境的编码器配置
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05") // 使用时间格式化布局
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // 使用大写颜色编码级别, 如 INFO, ERROR 等
	encoderConfig.ConsoleSeparator = " | " // 控制台打印的分隔符

	return zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig), // 使用控制台编码器
		zapcore.AddSync(os.Stdout),               // 将日志写入控制台
		level,                                    // 日志级别
	)
}

// GetLogger 返回全局 Logger
func GetLogger() *zap.Logger {
	return Log
}

// FromContext 从 gin.Context 获取 Logger
func FromContext(c interface{}) *zap.Logger {
	if gc, ok := c.(interface {
		Get(string) (interface{}, bool)
	}); ok {
		if val, exists := gc.Get("logger"); exists {
			if l, ok := val.(*zap.Logger); ok {
				return l
			}
		}
	}
	return Log
}

// 包装常用的 zap 字段，这样业务服务就不需要直接依赖 zap 库

// Field 是 zapcore.Field 的别名
type Field = zapcore.Field

// String 包装 zap.String
func String(key string, value string) Field {
	return zap.String(key, value)
}

// Int 包装 zap.Int
func Int(key string, value int) Field {
	return zap.Int(key, value)
}

// Err 包装 zap.Error
func Err(err error) Field {
	return zap.Error(err)
}

// Any 包装 zap.Any
func Any(key string, value interface{}) Field {
	return zap.Any(key, value)
}

// 快捷调用函数

func Debug(msg string, fields ...Field) {
	Log.Debug(msg, fields...)
}

func Info(msg string, fields ...Field) {
	Log.Info(msg, fields...)
}

func Warn(msg string, fields ...Field) {
	Log.Warn(msg, fields...)
}

func Error(msg string, fields ...Field) {
	Log.Error(msg, fields...)
}

func Fatal(msg string, fields ...Field) {
	Log.Fatal(msg, fields...)
}

func Panic(msg string, fields ...Field) {
	Log.Panic(msg, fields...)
}
