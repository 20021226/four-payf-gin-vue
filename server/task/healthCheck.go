package task

import (
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"go.uber.org/zap"
)

// 健康检查配置常量
const (
	//// 是否启用健康检查，默认关闭（代码级开关）
	//HEALTH_CHECK_ENABLED = false
	//// 要检查的服务器IP地址，默认使用Google DNS
	//HEALTH_CHECK_SERVER_IP = "8.8.8.8"
	//// 最大连续失败次数，达到后退出程序
	//HEALTH_CHECK_MAX_FAILURES = 10
	//// 检查间隔（秒），默认10秒
	//HEALTH_CHECK_INTERVAL = 10

	HEALTH_CHECK_ENABLED      = false
	HEALTH_CHECK_SERVER_IP    = "192.168.0.208"
	HEALTH_CHECK_MAX_FAILURES = 2
	HEALTH_CHECK_INTERVAL     = 2
)

// HealthChecker 健康检查器结构体
type HealthChecker struct {
	serverIP     string        // 要检查的服务器IP
	timeout      time.Duration // ping超时时间
	failureCount int           // 连续失败次数
	maxFailures  int           // 最大失败次数
	enabled      bool          // 是否启用健康检查
	mu           sync.RWMutex  // 读写锁
}

// NewHealthChecker 创建新的健康检查器
func NewHealthChecker(serverIP string, maxFailures int, enabled bool) *HealthChecker {
	return &HealthChecker{
		serverIP:     serverIP,
		timeout:      3 * time.Second, // 3秒超时
		failureCount: 0,
		maxFailures:  maxFailures,
		enabled:      enabled,
	}
}

// IsEnabled 检查是否启用
func (hc *HealthChecker) IsEnabled() bool {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	return hc.enabled
}

// SetEnabled 设置启用状态
func (hc *HealthChecker) SetEnabled(enabled bool) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	hc.enabled = enabled
}

// GetFailureCount 获取失败次数
func (hc *HealthChecker) GetFailureCount() int {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	return hc.failureCount
}

// ping 执行ping检查
func (hc *HealthChecker) ping() bool {
	if hc.serverIP == "" {
		global.GVA_LOG.Warn("服务器IP为空，跳过健康检查")
		return true
	}

	// 使用TCP连接测试，比ICMP ping更可靠
	conn, err := net.DialTimeout("tcp", hc.serverIP+":80", hc.timeout)
	if err != nil {
		// 如果80端口不通，尝试443端口
		conn, err = net.DialTimeout("tcp", hc.serverIP+":443", hc.timeout)
		if err != nil {
			// 如果都不通，尝试简单的网络连接测试
			conn, err = net.DialTimeout("tcp", hc.serverIP+":22", hc.timeout)
			if err != nil {
				return false
			}
		}
	}

	if conn != nil {
		conn.Close()
	}
	return true
}

// Check 执行健康检查
func (hc *HealthChecker) Check() {
	if !hc.IsEnabled() {
		return
	}

	success := hc.ping()

	hc.mu.Lock()
	defer hc.mu.Unlock()

	if success {
		if hc.failureCount > 0 {
			global.GVA_LOG.Info("服务器连接恢复正常",
				zap.String("serverIP", hc.serverIP),
				zap.Int("之前失败次数", hc.failureCount))
		}
		hc.failureCount = 0
	} else {
		hc.failureCount++
		global.GVA_LOG.Warn("服务器连接失败",
			zap.String("serverIP", hc.serverIP),
			zap.Int("连续失败次数", hc.failureCount),
			zap.Int("最大失败次数", hc.maxFailures))

		if hc.failureCount >= hc.maxFailures {
			global.GVA_LOG.Error("服务器连续失败次数达到上限，程序即将退出",
				zap.String("serverIP", hc.serverIP),
				zap.Int("失败次数", hc.failureCount),
				zap.Int("最大失败次数", hc.maxFailures))

			// 优雅退出程序
			go func() {
				time.Sleep(1 * time.Second) // 给日志一点时间写入
				os.Exit(1)
			}()
		}
	}
}

// Run 实现定时任务接口
func (hc *HealthChecker) Run() {
	hc.Check()
}

// 全局健康检查器实例
var (
	globalHealthChecker *HealthChecker
	healthCheckerOnce   sync.Once
)

// GetHealthChecker 获取全局健康检查器实例
func GetHealthChecker() *HealthChecker {
	return globalHealthChecker
}

// InitHealthChecker 初始化健康检查器
func InitHealthChecker() {
	healthCheckerOnce.Do(func() {
		globalHealthChecker = NewHealthChecker(
			HEALTH_CHECK_SERVER_IP,
			HEALTH_CHECK_MAX_FAILURES,
			HEALTH_CHECK_ENABLED,
		)
		global.GVA_LOG.Info("健康检查器初始化完成",
			zap.String("serverIP", HEALTH_CHECK_SERVER_IP),
			zap.Int("maxFailures", HEALTH_CHECK_MAX_FAILURES),
			zap.Bool("enabled", HEALTH_CHECK_ENABLED),
			zap.Int("interval", HEALTH_CHECK_INTERVAL))
	})
}

// StartHealthCheckTask 启动健康检查定时任务
func StartHealthCheckTask() {
	if globalHealthChecker == nil {
		global.GVA_LOG.Error("健康检查器未初始化")
		return
	}

	if !globalHealthChecker.IsEnabled() {
		global.GVA_LOG.Info("健康检查功能已禁用")
		return
	}

	// 使用常量配置构建cron表达式
	cronExpr := fmt.Sprintf("*/%d * * * * *", HEALTH_CHECK_INTERVAL)

	// 使用全局定时器添加健康检查任务
	_, err := global.GVA_Timer.AddTaskByJobWithSeconds(
		"health_check",        // cron名称
		cronExpr,              // 动态间隔时间
		globalHealthChecker,   // 任务对象
		"server_health_check", // 任务名称
	)

	if err != nil {
		global.GVA_LOG.Error("启动健康检查定时任务失败", zap.Error(err))
		return
	}

	global.GVA_LOG.Info("健康检查定时任务启动成功",
		zap.String("serverIP", globalHealthChecker.serverIP),
		zap.Int("interval", HEALTH_CHECK_INTERVAL),
		zap.String("schedule", fmt.Sprintf("每%d秒执行一次", HEALTH_CHECK_INTERVAL)))
}

// StopHealthCheckTask 停止健康检查定时任务
func StopHealthCheckTask() {
	if globalHealthChecker != nil {
		global.GVA_Timer.RemoveTaskByName("health_check", "server_health_check")
		global.GVA_LOG.Info("健康检查定时任务已停止")
	}
}
