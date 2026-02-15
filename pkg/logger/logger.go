package logger

import (
	"io"
	"os"
	"sync"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/sirupsen/logrus"
)

var (
	log  *logrus.Logger
	once sync.Once
)

// Config 日志配置
type Config struct {
	Level     string // debug, info, warn, error
	Format    string // json, text
	Output    io.Writer
	DebugMode bool
}

// Init 初始化日志系统
func Init(cfg *Config) {
	once.Do(func() {
		log = logrus.New()

		// 设置日志级别
		level, err := logrus.ParseLevel(cfg.Level)
		if err != nil {
			level = logrus.InfoLevel
		}
		log.SetLevel(level)

		// 设置日志格式
		if cfg.Format == "json" {
			log.SetFormatter(&logrus.JSONFormatter{
				TimestampFormat: "2006-01-02 15:04:05",
			})
		} else {
			log.SetFormatter(&logrus.TextFormatter{
				TimestampFormat: "2006-01-02 15:04:05",
				FullTimestamp:   true,
			})
		}

		// 设置输出
		if cfg.Output != nil {
			log.SetOutput(cfg.Output)
		} else {
			log.SetOutput(os.Stdout)
		}

		// Debug 模式
		if cfg.DebugMode {
			log.SetLevel(logrus.DebugLevel)
			log.SetReportCaller(true)
		}
	})
}

// InitDefault 使用默认配置初始化
func InitDefault() {
	Init(&Config{
		Level:  "info",
		Format: "text",
	})
}

// GetLogger 获取日志实例
func GetLogger() *logrus.Logger {
	if log == nil {
		InitDefault()
	}
	return log
}

// WithFields 带字段的日志
func WithFields(fields logrus.Fields) *logrus.Entry {
	return GetLogger().WithFields(fields)
}

// WithContext 从 Hertz Context 创建日志器
func WithContext(c *app.RequestContext) *logrus.Entry {
	fields := logrus.Fields{}

	// 尝试获取请求 ID
	if reqID := c.GetString("request_id"); reqID != "" {
		fields["request_id"] = reqID
	}

	// 获取请求信息
	fields["method"] = string(c.Method())
	fields["path"] = string(c.Path())

	return GetLogger().WithFields(fields)
}

// WithRepo 带仓库信息的日志
func WithRepo(repoKey, repoPath string) *logrus.Entry {
	return WithFields(logrus.Fields{
		"repo_key":  repoKey,
		"repo_path": repoPath,
	})
}

// WithOperation 带操作信息的日志
func WithOperation(op string) *logrus.Entry {
	return WithFields(logrus.Fields{
		"operation": op,
	})
}

// Debug 调试日志
func Debug(msg string, fields ...logrus.Fields) {
	if len(fields) > 0 {
		WithFields(fields[0]).Debug(msg)
	} else {
		GetLogger().Debug(msg)
	}
}

// Info 信息日志
func Info(msg string, fields ...logrus.Fields) {
	if len(fields) > 0 {
		WithFields(fields[0]).Info(msg)
	} else {
		GetLogger().Info(msg)
	}
}

// Warn 警告日志
func Warn(msg string, fields ...logrus.Fields) {
	if len(fields) > 0 {
		WithFields(fields[0]).Warn(msg)
	} else {
		GetLogger().Warn(msg)
	}
}

// Error 错误日志
func Error(msg string, fields ...logrus.Fields) {
	if len(fields) > 0 {
		WithFields(fields[0]).Error(msg)
	} else {
		GetLogger().Error(msg)
	}
}

// ErrorWithErr 带错误对象的错误日志
func ErrorWithErr(msg string, err error, fields ...logrus.Fields) {
	f := logrus.Fields{"error": err.Error()}
	if len(fields) > 0 {
		for k, v := range fields[0] {
			f[k] = v
		}
	}
	WithFields(f).Error(msg)
}
