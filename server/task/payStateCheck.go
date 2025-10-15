package task

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	exampleModel "github.com/flipped-aurora/gin-vue-admin/server/model/example"
	"github.com/flipped-aurora/gin-vue-admin/server/service"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/flipped-aurora/gin-vue-admin/server/utils/pay/xingyi"

	"go.uber.org/zap"
)

// OrderMonitorTask 订单监控任务结构体
type OrderMonitorTask struct {
	UserID    uint          // 用户ID
	Amount    string        // 金额
	RedisKey  string        // Redis键
	TTL       time.Duration // 过期时间
	StartTime time.Time     // 开始时间
	EndTime   time.Time     // 结束时间
	TaskID    string        // 任务ID
	once      sync.Once     // 确保任务只停止一次
	stopChan  chan struct{} // 停止信号
	MerType   string        // 商户类型
	OrderId   string        // 订单ID
	MerUserId int32         // 商户用户ID

}

// NewOrderMonitorTask 创建新的订单监控任务
func NewOrderMonitorTask(userID uint,
	amount, redisKey string,
	ttl time.Duration,
	merType string,
	orderId string,
	merUserId int32,
	startTime, endTime time.Time,
) *OrderMonitorTask {
	return &OrderMonitorTask{
		UserID:    userID,
		Amount:    amount,
		RedisKey:  redisKey,
		TTL:       ttl,
		TaskID:    fmt.Sprintf("order_monitor_%d_%s_%d", userID, amount, time.Now().Unix()),
		stopChan:  make(chan struct{}),
		MerType:   merType,
		OrderId:   orderId,
		MerUserId: merUserId,
		StartTime: startTime,
		EndTime:   endTime,
	}
}

// Start 启动订单监控任务
func (task *OrderMonitorTask) Start() {
	// 添加到全局定时器
	global.GVA_Timer.AddTaskByFunc(task.TaskID, "@every 10s", task.execute, task.TaskID)

	// 启动过期检查协程
	go task.checkExpiration()

	global.GVA_LOG.Info("订单监控任务已启动",
		zap.String("taskID", task.TaskID),
		zap.Uint("userID", task.UserID),
		zap.String("amount", task.Amount),
		zap.Duration("ttl", task.TTL))
}

// execute 执行监控操作（每10秒执行一次）
func (task *OrderMonitorTask) execute() {
	select {
	case <-task.stopChan:
		return
	default:
		// 这里可以添加具体的监控操作
		// 例如：检查订单状态、发送通知等
		global.GVA_LOG.Info("执行订单监控检查",
			zap.String("taskID", task.TaskID),
			zap.Uint("userID", task.UserID),
			zap.String("amount", task.Amount),
			zap.String("redisKey", task.RedisKey))

		// TODO: 在这里添加具体的监控逻辑
		// 例如：
		// 1. 检查Redis中的订单状态
		// 2. 检查支付状态
		// 3. 发送状态更新通知
		// 4. 记录监控日志等
		// 通过 Redis key 获取订单数据
		ctx := context.Background()
		var merUser exampleModel.MerUser
		redisData, err := global.GVA_REDIS.Get(ctx, fmt.Sprintf("%s:%s", task.MerType, task.OrderId)).Result()
		if err != nil {
			merUser, err = service.ServiceGroupApp.ExampleServiceGroup.MerUserService.GetMerUser(ctx, fmt.Sprintf("%d", task.MerUserId))
			if err != nil {
				global.GVA_LOG.Error("从数据库获取商户用户失败",
					zap.String("taskID", task.TaskID),
					zap.Int32("merUserId", task.MerUserId),
					zap.Error(err))
				return
			}
			global.GVA_LOG.Error("从 Redis 获取订单数据失败",
				zap.String("taskID", task.TaskID),
				zap.String("redisKey", task.RedisKey),
				zap.Error(err))
			return
		}

		global.GVA_LOG.Info("成功获取 Redis 订单数据",
			zap.String("taskID", task.TaskID),
			zap.String("redisKey", task.RedisKey),
			zap.String("redisData", redisData))
		//xingyi
		if task.MerType == global.MER_TYPE_XINGYI {
			// 调用星驿付相关服务
			xingyiService := xingyi.NewService(nil, nil)

			token_key := fmt.Sprintf("%s:%s:%s", global.REDIS_PAY_REQUEST_TOKEN, task.MerType, task.MerUserId)
			tokenData, err := global.GVA_REDIS.Get(ctx, token_key).Result()
			if err != nil {
				global.GVA_LOG.Error("从 Redis 获取星驿付 token 失败",
					zap.String("taskID", task.TaskID),
					zap.String("redisKey", token_key),
					zap.Error(err))
				return
			}

			pwd, err := xingyiService.GetLoginResult(*merUser.UserName, *merUser.Password, fmt.Sprintf("%d", merUser.Id))
			if err != nil {
				global.GVA_LOG.Error("星驿付登录失败", zap.Error(err))
				return
			}
			fmt.Println("PWD_RESPDATA:", pwd)

			token, err := xingyiService.GetAccessToken(pwd)
			if err != nil {
				log.Fatalf("get token error: %v", err)
			}
			fmt.Println("ACCESS_TOKEN:", token)

			// 计算格式化的开始时间和结束时间字符串
			startTimeStr, endTimeStr := utils.GetTimeRangeStr("20060102150405", task.TTL)
			
			// 计算实际的时间范围用于检查
			var ttl time.Duration
			if task.TTL == 0 {
				ttl = time.Duration(5) * time.Minute
			} else {
				ttl = task.TTL
			}
			now := time.Now()
			checkStartTime := now.Add(-ttl)
			checkEndTime := now
			
			global.GVA_LOG.Info("查询支付列表时间范围",
				zap.String("startTimeStr", startTimeStr),
				zap.String("endTimeStr", endTimeStr),
				zap.Time("checkStartTime", checkStartTime),
				zap.Time("checkEndTime", checkEndTime),
				zap.Duration("ttl", task.TTL))

			body, err := xingyiService.GetPayList(token, startTimeStr, endTimeStr, fmt.Sprintf("%d", merUser.Id))
			if err != nil {
				global.GVA_LOG.Error("获取支付列表失败", zap.Error(err))
				return
			}

			global.GVA_LOG.Info("获取支付列表成功", zap.String("response", string(body)))

			// 检查是否有符合条件的订单
			found, err := xingyiService.GetCheckResult(string(body), checkStartTime, checkEndTime, task.Amount)
			if err != nil {
				global.GVA_LOG.Error("检查支付结果失败", zap.Error(err))
				return
			}

			if found {
				global.GVA_LOG.Info("找到匹配的支付订单，停止监控任务",
					zap.String("taskID", task.TaskID),
					zap.String("amount", task.Amount))
				task.Stop()
				return
			} else {
				global.GVA_LOG.Debug("未找到匹配的支付订单，继续监控",
					zap.String("taskID", task.TaskID),
					zap.String("amount", task.Amount))
			}

		} else if task.MerType == "1" {
			// 富掌柜
		} else if task.MerType == "2" {
			//xiangxian

		}
	}
}

// checkExpiration 检查任务是否过期
func (task *OrderMonitorTask) checkExpiration() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-task.stopChan:
			return
		case <-ticker.C:
			if time.Since(task.StartTime) >= task.TTL {
				task.Stop()
				return
			}
		}
	}
}

// Stop 停止订单监控任务
func (task *OrderMonitorTask) Stop() {
	task.once.Do(func() {
		close(task.stopChan)
		global.GVA_Timer.RemoveTaskByName(task.TaskID, task.TaskID)
		global.GVA_LOG.Info("订单监控任务已停止",
			zap.String("taskID", task.TaskID),
			zap.Uint("userID", task.UserID),
			zap.String("amount", task.Amount),
			zap.Duration("运行时长", time.Since(task.StartTime)))
	})
}
