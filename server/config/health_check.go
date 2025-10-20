package config

// HealthCheck 健康检查配置
type HealthCheck struct {
	Enabled     bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`           // 是否启用健康检查
	ServerIP    string `mapstructure:"server-ip" json:"server-ip" yaml:"server-ip"`    // 要检查的服务器IP地址
	MaxFailures int    `mapstructure:"max-failures" json:"max-failures" yaml:"max-failures"` // 最大连续失败次数，达到后退出程序
	Interval    int    `mapstructure:"interval" json:"interval" yaml:"interval"`       // 检查间隔（秒），默认10秒
}