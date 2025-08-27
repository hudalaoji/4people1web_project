package config

import "time"

// Config 系统配置
type Config struct {
	ServerPort     string        // 服务器端口
	MaxMessageSize int64         // 最大消息大小（字节）
	ReadTimeout    time.Duration // 读取超时
	WriteTimeout   time.Duration // 写入超时
}

// DefaultConfig 创建默认配置
func DefaultConfig() *Config {
	return &Config{
		ServerPort:     ":8080",
		MaxMessageSize: 512,
		ReadTimeout:    60 * time.Second,  // 直接使用time.Duration类型
		WriteTimeout:   10 * time.Second,  // 直接使用time.Duration类型
	}
}
