package task

import (
	"context"
	"fmt"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/example"
	"github.com/flipped-aurora/gin-vue-admin/server/service"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/flipped-aurora/gin-vue-admin/server/utils/pay/xianxiang"
	"github.com/flipped-aurora/gin-vue-admin/server/utils/pay/xingyi"
	"github.com/flipped-aurora/gin-vue-admin/server/utils/request"
	"github.com/google/uuid"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// MerUserTaskManager 全局任务管理器，确保每个 meruser 只有一个定时任务
type MerUserTaskManager struct {
	tasks map[int64]*MerUserMonitorTask // key: merUserId, value: 任务实例
	mutex sync.RWMutex                  // 读写锁保护 tasks map
}

// 全局任务管理器实例
var GlobalTaskManager = &MerUserTaskManager{
	tasks: make(map[int64]*MerUserMonitorTask),
}

// MerUserMonitorTask 基于 meruser ID 的监控任务结构体
type MerUserMonitorTask struct {
	MerUserId   int64         // 商户用户ID
	TaskID      string        // 任务ID
	once        sync.Once     // 确保任务只停止一次
	stopChan    chan struct{} // 停止信号
	CallBackUrl string        // 回调URL
	MerType     string        // 商户类型
	StartTime   time.Time     // 任务开始时间
}

// StartMerUserTask 启动或获取 meruser 的监控任务
func (manager *MerUserTaskManager) StartMerUserTask(merUserId int64, callbackUrl, merType string) *MerUserMonitorTask {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	// 检查是否已存在任务
	if existingTask, exists := manager.tasks[merUserId]; exists {
		global.GVA_LOG.Info("MerUser 任务已存在，返回现有任务",
			zap.Int64("merUserId", merUserId),
			zap.String("taskID", existingTask.TaskID))
		return existingTask
	}

	// 创建新任务
	task := &MerUserMonitorTask{
		MerUserId:   merUserId,
		TaskID:      fmt.Sprintf("meruser_monitor_%d_%s", merUserId, uuid.New().String()),
		stopChan:    make(chan struct{}),
		CallBackUrl: callbackUrl,
		MerType:     merType,
		StartTime:   time.Now(),
	}

	// 添加到管理器
	manager.tasks[merUserId] = task

	// 启动任务
	task.start()

	global.GVA_LOG.Info("MerUser 监控任务已启动",
		zap.Int64("merUserId", merUserId),
		zap.String("taskID", task.TaskID),
		zap.String("merType", merType))

	return task
}

// StopMerUserTask 停止指定 meruser 的监控任务
func (manager *MerUserTaskManager) StopMerUserTask(merUserId int64) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	if task, exists := manager.tasks[merUserId]; exists {
		task.stop()
		delete(manager.tasks, merUserId)
		global.GVA_LOG.Info("MerUser 监控任务已停止",
			zap.Int64("merUserId", merUserId),
			zap.String("taskID", task.TaskID))
	}
}

// GetMerUserTask 获取指定 meruser 的监控任务
func (manager *MerUserTaskManager) GetMerUserTask(merUserId int64) (*MerUserMonitorTask, bool) {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()

	task, exists := manager.tasks[merUserId]
	return task, exists
}

// start 启动 meruser 监控任务
func (task *MerUserMonitorTask) start() {
	// 添加到全局定时器，每30秒检查一次
	global.GVA_Timer.AddTaskByFunc(task.TaskID, "@every 30s", task.execute, task.TaskID)

	global.GVA_LOG.Info("MerUser 监控任务定时器已启动",
		zap.String("taskID", task.TaskID),
		zap.Int64("merUserId", task.MerUserId))
}

// execute 执行监控操作（每30秒执行一次）
func (task *MerUserMonitorTask) execute() {
	select {
	case <-task.stopChan:
		return
	default:
		// 监控该 meruser 的金额占用情况
		ctx := context.Background()

		// 获取商户用户信息
		merUser, err := service.ServiceGroupApp.ExampleServiceGroup.MerUserService.GetMerUser(ctx, fmt.Sprintf("%d", task.MerUserId))
		if err != nil {
			global.GVA_LOG.Error("从数据库获取商户用户失败",
				zap.String("taskID", task.TaskID),
				zap.Int64("merUserId", task.MerUserId),
				zap.Error(err))
			return
		}

		// 构建 Redis 键的模式，查找该用户的所有金额占用
		keyPattern := fmt.Sprintf("%s:%d:*", global.PAY_AMOUNT_USED_KEY, task.MerUserId)

		// 获取所有已占用的金额键
		var occupiedKeys []string
		iter := global.GVA_REDIS.Scan(ctx, 0, keyPattern, 100).Iterator()

		for iter.Next(ctx) {
			key := iter.Val()
			occupiedKeys = append(occupiedKeys, key)
		}

		if err := iter.Err(); err != nil {
			global.GVA_LOG.Error("Redis SCAN 操作失败",
				zap.String("taskID", task.TaskID),
				zap.Int64("merUserId", task.MerUserId),
				zap.Error(err))
			return
		}

		if len(occupiedKeys) == 0 {
			global.GVA_LOG.Debug("MerUser 当前无金额占用",
				zap.String("taskID", task.TaskID),
				zap.Int64("merUserId", task.MerUserId))
			return
		}

		global.GVA_LOG.Info("MerUser 金额占用检查",
			zap.String("taskID", task.TaskID),
			zap.Int64("merUserId", task.MerUserId),
			zap.Int("occupiedCount", len(occupiedKeys)))

		// 检查每个占用的金额是否有对应的支付
		for _, redisKey := range occupiedKeys {
			// 从键中提取金额部分
			parts := strings.Split(redisKey, ":")
			if len(parts) < 3 {
				continue
			}

			amountStr := parts[2]

			// 获取订单ID
			orderIdStr, err := global.GVA_REDIS.Get(ctx, redisKey).Result()
			if err != nil {
				global.GVA_LOG.Error("获取订单ID失败",
					zap.String("redisKey", redisKey),
					zap.Error(err))
				continue
			}

			orderId, err := strconv.ParseInt(orderIdStr, 10, 64)
			if err != nil {
				global.GVA_LOG.Error("解析订单ID失败",
					zap.String("orderIdStr", orderIdStr),
					zap.Error(err))
				continue
			}

			// 检查订单状态
			payOrder, err := service.ServiceGroupApp.ExampleServiceGroup.GetMerPayOrder(ctx, orderId)
			if err != nil {
				global.GVA_LOG.Error("获取支付订单失败",
					zap.Int64("orderId", orderId),
					zap.Error(err))
				continue
			}

			// 如果订单已支付或已失败，跳过
			if payOrder.State != nil && (*payOrder.State == *global.MER_PAY_ORDER_PAID || *payOrder.State == *global.MER_PAY_ORDER_FAILED) {
				continue
			}

			// 检查订单是否超时
			if payOrder.CreateTime != nil {
				createTime := *payOrder.CreateTime
				if payOrder.Expires != nil && time.Since(createTime) > time.Duration(*payOrder.Expires)*time.Second {
					// 订单超时，标记为失败
					err := service.ServiceGroupApp.ExampleServiceGroup.UpdateMerPayOrder(ctx, example.MerPayOrder{
						Id:    &orderId,
						State: global.MER_PAY_ORDER_FAILED,
					})
					if err != nil {
						global.GVA_LOG.Error("更新超时订单状态失败",
							zap.Int64("orderId", orderId),
							zap.Error(err))
					} else {
						global.GVA_LOG.Info("订单已超时，标记为失败",
							zap.Int64("orderId", orderId),
							zap.String("amount", amountStr))
					}
					continue
				}
			}

			// 检查是否有对应的支付
			found := task.checkPaymentForAmount(ctx, merUser, amountStr, payOrder)
			if found {
				// 找到支付，处理支付成功
				now := time.Now()
				task.handlePaymentSuccess(ctx, now, task.MerType, orderId, amountStr, payOrder)
			}
		}
	}
}

// checkPaymentForAmount 检查指定金额是否有对应的支付
func (task *MerUserMonitorTask) checkPaymentForAmount(ctx context.Context, merUser *example.MerUser, amountStr string, payOrder *example.MerPayOrder) bool {
	if task.MerType == global.MER_TYPE_XINGYI {
		return task.checkXingyiPayment(ctx, merUser, amountStr, payOrder)
	} else if task.MerType == global.MER_TYPE_XIANG_XIAN {
		return task.checkXianxiangPayment(ctx, merUser, amountStr, payOrder)
	} else if task.MerType == global.MER_TYPE_RICH {
		// 富掌柜支付检查逻辑
		return false
	}
	return false
}

// checkXingyiPayment 检查星驿付支付
func (task *MerUserMonitorTask) checkXingyiPayment(ctx context.Context, merUser *example.MerUser, amountStr string, payOrder *example.MerPayOrder) bool {
	xingyiService := xingyi.NewService(nil, nil, xingyi.Cookies{})
	merUserId := strconv.FormatInt(task.MerUserId, 10)

	tokenKey := fmt.Sprintf("%s:%s:%s", global.REDIS_PAY_REQUEST_TOKEN, task.MerType, merUserId)
	tokenData, err := global.GVA_REDIS.Get(ctx, tokenKey).Result()
	if err != nil {
		global.GVA_LOG.Error("从 Redis 获取星驿付 token 失败",
			zap.String("taskID", task.TaskID),
			zap.String("redisKey", tokenKey),
			zap.Error(err))
	}

	// 检查 token 是否为空或校验失败，需要重新获取
	if tokenData == "" || !xingyiService.CheckToken(tokenData, merUserId) {
		newToken, success := task.refreshXingyiToken(ctx, merUser, xingyiService, merUserId, tokenKey)
		if !success {
			return false
		}
		tokenData = newToken
	}

	// 计算时间范围
	var startTime, endTime time.Time
	if payOrder.CreateTime != nil {
		startTime = *payOrder.CreateTime
		if payOrder.Expires != nil {
			endTime = startTime.Add(time.Duration(*payOrder.Expires) * time.Second)
		} else {
			endTime = startTime.Add(5 * time.Minute) // 默认5分钟
		}
	} else {
		startTime = time.Now().Add(-10 * time.Minute)
		endTime = time.Now()
	}

	global.GVA_LOG.Info("查询星驿付支付列表",
		zap.String("startTime", utils.FormatTime(startTime)),
		zap.String("endTime", utils.FormatTime(endTime)),
		zap.String("amount", amountStr))

	body, err := xingyiService.GetPayList(tokenData, utils.FormatTime(startTime), utils.FormatTime(endTime), merUserId)
	if err != nil {
		global.GVA_LOG.Error("获取星驿付支付列表失败", zap.Error(err))
		return false
	}

	// 检查是否有符合条件的订单
	found, err := xingyiService.GetCheckResult(string(body), startTime, endTime, amountStr)
	if err != nil {
		global.GVA_LOG.Error("检查星驿付支付结果失败", zap.Error(err))
		return false
	}

	return found
}

// checkXianxiangPayment 检查先享后付支付
func (task *MerUserMonitorTask) checkXianxiangPayment(ctx context.Context, merUser *example.MerUser, amountStr string, payOrder *example.MerPayOrder) bool {
	xianxiangService := xianxiang.NewService(nil)
	merUserId := fmt.Sprintf("%d", task.MerUserId)

	tokenKey := fmt.Sprintf("%s:%s:%s", global.REDIS_PAY_REQUEST_TOKEN, task.MerType, merUserId)
	tokenData, err := global.GVA_REDIS.Get(ctx, tokenKey).Result()
	if err != nil {
		global.GVA_LOG.Error("从 Redis 获取先享后付 token 失败",
			zap.String("taskID", task.TaskID),
			zap.String("redisKey", tokenKey),
			zap.Error(err))
	}

	// 检查 token 是否为空或校验失败，需要重新获取
	if tokenData == "" || !xianxiangService.CheckToken(tokenData) {
		newToken, success := task.refreshXianxiangToken(ctx, merUser, xianxiangService, merUserId, tokenKey)
		if !success {
			return false
		}
		tokenData = newToken
	}

	// 计算时间范围
	var startTime time.Time
	if payOrder.CreateTime != nil {
		startTime = *payOrder.CreateTime
	} else {
		startTime = time.Now().Add(-10 * time.Minute)
	}

	// 先享后付使用日期格式：2025-10-13
	orderTime := startTime.Format("2006-01-02")

	global.GVA_LOG.Info("查询先享后付订单列表",
		zap.String("orderTime", orderTime),
		zap.String("amount", amountStr))

	body, err := xianxiangService.GetOrderList(tokenData, orderTime)
	if err != nil {
		global.GVA_LOG.Error("获取先享后付订单列表失败", zap.Error(err))
		return false
	}

	// 计算结束时间
	var endTime time.Time
	if payOrder.CreateTime != nil && payOrder.Expires != nil {
		endTime = payOrder.CreateTime.Add(time.Duration(*payOrder.Expires) * time.Second)
	} else {
		endTime = startTime.Add(5 * time.Minute) // 默认5分钟
	}

	// 检查是否有符合条件的订单
	found, _, err := xianxiangService.CheckPayment(string(body), startTime, endTime, amountStr)
	if err != nil {
		global.GVA_LOG.Error("检查先享后付支付结果失败", zap.Error(err))
		return false
	}

	return found
}

// refreshXingyiToken 刷新星驿付 token
func (task *MerUserMonitorTask) refreshXingyiToken(ctx context.Context, merUser *example.MerUser, xingyiService *xingyi.Service, merUserId, tokenKey string) (string, bool) {
	maxRetries := 3
	baseDelay := 1 * time.Second

	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			delay := time.Duration(i) * baseDelay
			time.Sleep(delay)
		}

		pwd, err := xingyiService.GetLoginResult(*merUser.UserName, *merUser.Password, merUserId)
		if err != nil {
			global.GVA_LOG.Error("星驿付登录失败",
				zap.Error(err),
				zap.Int("attempt", i+1))
			continue
		}

		token, err := xingyiService.GetAccessToken(pwd, merUserId)
		if err != nil {
			global.GVA_LOG.Error("获取星驿付 token 失败",
				zap.Error(err),
				zap.Int("attempt", i+1))
			continue
		}

		// 保存新 token
		err = global.GVA_REDIS.Set(ctx, tokenKey, token, 1*time.Hour).Err()
		if err != nil {
			global.GVA_LOG.Error("保存星驿付 token 失败", zap.Error(err))
		}

		global.GVA_LOG.Info("星驿付 token 刷新成功",
			zap.String("taskID", task.TaskID),
			zap.Int("attempt", i+1))

		return token, true
	}

	global.GVA_LOG.Error("星驿付 token 刷新失败",
		zap.String("taskID", task.TaskID),
		zap.Int("maxRetries", maxRetries))
	return "", false
}

// refreshXianxiangToken 刷新先享后付 token
func (task *MerUserMonitorTask) refreshXianxiangToken(ctx context.Context, merUser *example.MerUser, xianxiangService *xianxiang.Service, merUserId, tokenKey string) (string, bool) {
	maxRetries := 3

	for i := 0; i < maxRetries; i++ {
		loginResult, err := xianxiangService.GetToken(*merUser.UserName, *merUser.Password)
		if err != nil {
			global.GVA_LOG.Error("先享后付登录失败",
				zap.Error(err),
				zap.Int("attempt", i+1))
			continue
		}

		// 保存 token
		err = xianxiangService.SaveToken(loginResult.AccessToken, merUserId, loginResult.ExpiresIn)
		if err != nil {
			global.GVA_LOG.Error("保存先享后付 token 失败", zap.Error(err))
		}

		global.GVA_LOG.Info("先享后付 token 刷新成功",
			zap.String("taskID", task.TaskID),
			zap.Int("attempt", i+1))

		return loginResult.AccessToken, true
	}

	global.GVA_LOG.Error("先享后付 token 刷新失败",
		zap.String("taskID", task.TaskID),
		zap.Int("maxRetries", maxRetries))
	return "", false
}

// handlePaymentSuccess 处理支付成功的通用逻辑
func (task *MerUserMonitorTask) handlePaymentSuccess(ctx context.Context, payTime time.Time, merType string, orderId int64, amountStr string, payOrder *example.MerPayOrder) {
	// 更新订单状态为已支付
	err := service.ServiceGroupApp.ExampleServiceGroup.UpdateMerPayOrder(ctx, example.MerPayOrder{
		Id:      &orderId,
		State:   global.MER_PAY_ORDER_PAID,
		PayTime: &payTime,
	})
	if err != nil {
		global.GVA_LOG.Error("更新支付订单失败",
			zap.String("taskID", task.TaskID),
			zap.String("paymentMethod", merType),
			zap.Int64("orderId", orderId),
			zap.Error(err))
		return
	}

	// 发送支付成功回调
	if task.CallBackUrl != "" && payOrder.OrderId != nil {
		amount, _ := strconv.ParseFloat(amountStr, 64)
		callbackData := PaymentCallbackData{
			OrderId:       *payOrder.OrderId,
			Amount:        amount,
			PayTime:       payTime,
			PaymentMethod: merType,
			Status:        "success",
			TransactionId: orderId,
		}

		// 异步发送回调
		go func() {
			err := sendPaymentCallback(task.CallBackUrl, callbackData)
			if err != nil {
				global.GVA_LOG.Error("支付回调发送失败",
					zap.String("merType", merType),
					zap.String("orderId", *payOrder.OrderId),
					zap.String("callbackUrl", task.CallBackUrl),
					zap.Error(err))
			}
		}()
	}

	global.GVA_LOG.Info("找到匹配的支付订单",
		zap.String("taskID", task.TaskID),
		zap.String("merType", merType),
		zap.String("amount", amountStr),
		zap.Int64("orderId", orderId))
}

// stop 停止 meruser 监控任务
func (task *MerUserMonitorTask) stop() {
	task.once.Do(func() {
		close(task.stopChan)
		global.GVA_Timer.RemoveTaskByName(task.TaskID, task.TaskID)
		global.GVA_LOG.Info("MerUser 监控任务已停止",
			zap.String("taskID", task.TaskID),
			zap.Int64("merUserId", task.MerUserId),
			zap.Duration("运行时长", time.Since(task.StartTime)))
	})
}

// OrderMonitorTask 订单监控任务结构体
type OrderMonitorTask struct {
	UserID      uint          // 用户ID
	Amount      string        // 金额
	RedisKey    string        // Redis键
	TTL         time.Duration // 过期时间
	StartTime   time.Time     // 开始时间
	EndTime     time.Time     // 结束时间
	TaskID      string        // 任务ID
	once        sync.Once     // 确保任务只停止一次
	stopChan    chan struct{} // 停止信号
	MerType     string        // 商户类型
	OrderId     string        // 订单ID
	MerUserId   int64         // 商户用户ID
	CallBackUrl string        // 回调URL
	OrderUniId  int64         // 订单唯一ID

}

// NewOrderMonitorTask 创建新的订单监控任务
func NewOrderMonitorTask(userID uint,
	amount, redisKey string,
	ttl time.Duration,
	merType string,
	orderId string,
	merUserId int64,
	startTime, endTime time.Time,
	callbackUrl string,
	orderUniId int64,
) *OrderMonitorTask {
	return &OrderMonitorTask{
		UserID:      userID,
		Amount:      amount,
		RedisKey:    redisKey,
		TTL:         ttl,
		TaskID:      fmt.Sprintf("order_monitor_%d_%s_%d", userID, amount, uuid.New().String()),
		stopChan:    make(chan struct{}),
		MerType:     merType,
		OrderId:     orderId,
		MerUserId:   merUserId,
		StartTime:   startTime,
		EndTime:     endTime,
		CallBackUrl: callbackUrl,
		OrderUniId:  orderUniId,
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

		// 通过 Redis key 获取订单数据
		ctx := context.Background()

		merUser, err := service.ServiceGroupApp.ExampleServiceGroup.MerUserService.GetMerUser(ctx, fmt.Sprintf("%d", task.MerUserId))
		if err != nil {
			global.GVA_LOG.Error("从数据库获取商户用户失败",
				zap.String("taskID", task.TaskID),
				zap.Int64("merUserId", task.MerUserId),
				zap.Error(err))
			return
		}

		//xingyi
		if task.MerType == global.MER_TYPE_XINGYI {
			// 调用星驿付相关服务
			xingyiService := xingyi.NewService(nil, nil, xingyi.Cookies{})
			merUserId := strconv.FormatInt(task.MerUserId, 10)

			tokenKey := fmt.Sprintf("%s:%s:%s", global.REDIS_PAY_REQUEST_TOKEN, task.MerType, merUserId)
			tokenData, err := global.GVA_REDIS.Get(ctx, tokenKey).Result()
			if err != nil {
				global.GVA_LOG.Error("从 Redis 获取星驿付 token 失败",
					zap.String("taskID", task.TaskID),
					zap.String("redisKey", tokenKey),
					zap.Error(err))
			}

			// 检查 token 是否为空或校验失败，需要重新获取
			if tokenData == "" || !xingyiService.CheckToken(tokenData, merUserId) {
				if tokenData == "" {
					global.GVA_LOG.Warn("星驿付 token 为空，需要重新获取",
						zap.String("taskID", task.TaskID),
						zap.String("redisKey", tokenKey))
				} else {
					global.GVA_LOG.Error("星驿付 token 校验失败",
						zap.String("taskID", task.TaskID),
						zap.String("redisKey", tokenKey),
						zap.Error(err))
				}

				getTokenSuccess := false
				var newToken string
				maxRetries := 5
				baseDelay := 1 * time.Second // 基础延迟时间

				for i := 0; i < maxRetries; i++ {
					// 如果不是第一次重试，则等待一段时间（指数退避）
					if i > 0 {
						delay := time.Duration(i) * baseDelay // 线性增长延迟：1s, 2s, 3s, 4s
						global.GVA_LOG.Info("星驿付token获取重试",
							zap.Int("attempt", i+1),
							zap.Int("maxRetries", maxRetries),
							zap.Duration("delay", delay))
						time.Sleep(delay)
					}

					pwd, err := xingyiService.GetLoginResult(*merUser.UserName, *merUser.Password, merUserId)
					if err != nil {
						global.GVA_LOG.Error("星驿付登录失败",
							zap.Error(err),
							zap.Int("attempt", i+1),
							zap.Int("maxRetries", maxRetries))
						continue
					}

					token, err := xingyiService.GetAccessToken(pwd, merUserId)
					if err != nil {
						global.GVA_LOG.Error("get token error",
							zap.Error(err),
							zap.Int("attempt", i+1),
							zap.Int("maxRetries", maxRetries))
						continue
					}

					global.GVA_LOG.Info("星驿付token获取成功",
						zap.String("ACCESS_TOKEN", token),
						zap.String("PWD_RESPDATA", pwd),
						zap.Int("attempt", i+1))

					newToken = token
					getTokenSuccess = true
					break
				}

				if !getTokenSuccess {
					global.GVA_LOG.Error("获取token失败", zap.Error(err),
						zap.String("taskID", task.TaskID),
						zap.String("redisKey", tokenKey),
						zap.String("message", "标记商户异常"),
					)
					////标记商户异常
					//err := service.ServiceGroupApp.ExampleServiceGroup.MerUserService.UpdateMerUser(ctx, example.MerUser{
					//	Id:    merUser.Id,
					//	State: &getTokenSuccess,
					//})
					//if err != nil {
					//	return
					//}
					return
				} else {
					// 成功获取新 token，保存到 Redis 并更新 tokenData
					err := global.GVA_REDIS.Set(ctx, tokenKey, newToken, 1*time.Hour).Err()
					if err != nil {
						global.GVA_LOG.Error("保存新 token 到 Redis 失败", zap.Error(err))
					} else {
						global.GVA_LOG.Info("成功获取并保存新 token",
							zap.String("taskID", task.TaskID),
							zap.String("redisKey", tokenKey))
					}
					tokenData = newToken
				}
			}

			// 计算格式化的开始时间和结束时间字符串
			startTime := task.StartTime
			endTime := task.EndTime

			global.GVA_LOG.Info("查询支付列表时间范围",
				zap.String("startTime", utils.FormatTime(startTime)),
				zap.String("endTime", utils.FormatTime(endTime)),
				zap.Duration("ttl", task.TTL))

			body, err := xingyiService.GetPayList(tokenData, utils.FormatTime(startTime), utils.FormatTime(endTime), merUserId)
			if err != nil {
				global.GVA_LOG.Error("获取支付列表失败", zap.Error(err))
				return
			}
			global.GVA_LOG.Info("获取支付列表成功", zap.String("response", string(body)))

			// 检查是否有符合条件的订单
			found, err := xingyiService.GetCheckResult(string(body), startTime, endTime, task.Amount)
			if err != nil {
				global.GVA_LOG.Error("检查支付结果失败", zap.Error(err))
				return
			}

			if found {
				// 使用通用的支付成功处理函数
				now := time.Now()
				task.handlePaymentSuccess(ctx, now, task.MerType, task.OrderUniId)
				return
			} else {
				global.GVA_LOG.Debug("未找到匹配的支付订单，继续监控",
					zap.String("taskID", task.TaskID),
					zap.String("amount", task.Amount))
			}

		} else if task.MerType == global.MER_TYPE_RICH {
			// 富掌柜
		} else if task.MerType == global.MER_TYPE_XIANG_XIAN {
			// 先享后付
			xianxiangService := xianxiang.NewService(nil)
			merUserId := fmt.Sprintf("%d", task.MerUserId)

			tokenKey := fmt.Sprintf("%s:%s:%s", global.REDIS_PAY_REQUEST_TOKEN, task.MerType, merUserId)
			tokenData, err := global.GVA_REDIS.Get(ctx, tokenKey).Result()
			if err != nil {
				global.GVA_LOG.Error("从 Redis 获取先付后享支付 token 失败",
					zap.String("taskID", task.TaskID),
					zap.String("redisKey", tokenKey),
					zap.Error(err))
			}

			// 检查 token 是否为空或校验失败，需要重新获取
			if tokenData == "" || !xianxiangService.CheckToken(tokenData) {
				if tokenData == "" {
					global.GVA_LOG.Warn("先付后享支付 token 为空，需要重新获取",
						zap.String("taskID", task.TaskID),
						zap.String("redisKey", tokenKey))
				} else {
					global.GVA_LOG.Error("先付后享支付 token 校验失败",
						zap.String("taskID", task.TaskID),
						zap.String("redisKey", tokenKey),
						zap.Error(err))
				}

				getTokenSuccess := false
				var newToken string
				for i := 0; i < 5; i++ {
					loginResult, err := xianxiangService.GetToken(*merUser.UserName, *merUser.Password)
					if err != nil {
						global.GVA_LOG.Error("先付后享支付登录失败", zap.Error(err))
						continue
					}

					// 保存 token 到 Redis
					err = xianxiangService.SaveToken(loginResult.AccessToken, merUserId, loginResult.ExpiresIn)
					if err != nil {
						global.GVA_LOG.Error("保存先付后享支付 token 失败", zap.Error(err))
					}

					fmt.Println("ACCESS_TOKEN:", loginResult.AccessToken)
					fmt.Println("TOKEN_TYPE:", loginResult.TokenType)
					fmt.Println("EXPIRES_IN:", loginResult.ExpiresIn)
					newToken = loginResult.AccessToken
					getTokenSuccess = true
					break
				}

				if !getTokenSuccess {
					global.GVA_LOG.Error("获取先付后享支付 token 失败",
						zap.String("taskID", task.TaskID),
						zap.String("redisKey", tokenKey))
					////标记商户异常
					//err := service.ServiceGroupApp.ExampleServiceGroup.MerUserService.UpdateMerUser(ctx, example.MerUser{
					//	Id:    merUser.Id,
					//	State: &getTokenSuccess,
					//})
					//if err != nil {
					//	return
					//}
					return
				} else {
					// 成功获取新 token，更新 tokenData
					global.GVA_LOG.Info("成功获取并保存新的先付后享支付 token",
						zap.String("taskID", task.TaskID),
						zap.String("redisKey", tokenKey))
					tokenData = newToken
				}
			}

			// 计算格式化的开始时间和结束时间字符串
			startTime := task.StartTime
			endTime := task.EndTime

			// 先付后享支付使用日期格式：2025-10-13
			orderTime := startTime.Format("2006-01-02")

			global.GVA_LOG.Info("查询支付订单列表时间范围",
				zap.String("orderTime", orderTime),
				zap.Time("startTime", startTime),
				zap.Time("endTime", endTime),
				zap.Duration("ttl", task.TTL))

			body, err := xianxiangService.GetOrderList(tokenData, orderTime)
			if err != nil {
				global.GVA_LOG.Error("获取支付订单列表失败", zap.Error(err))
				return
			}
			global.GVA_LOG.Info("获取支付订单列表成功", zap.String("response", string(body)))

			// 检查是否有符合条件的订单
			found, orderInfo, err := xianxiangService.CheckPayment(string(body), startTime, endTime, task.Amount)
			if err != nil {
				global.GVA_LOG.Error("检查支付结果失败", zap.Error(err))
				return
			}

			if found {
				// 解析支付时间
				var payTime time.Time
				if orderInfo != nil && orderInfo.CreateAt != "" {
					// 将字符串时间转换为time.Time
					parsedTime, err := time.Parse("2006-01-02 15:04:05", orderInfo.CreateAt)
					if err != nil {
						// 如果解析失败，使用当前时间
						payTime = time.Now()
						global.GVA_LOG.Warn("解析支付时间失败，使用当前时间",
							zap.String("原始时间", orderInfo.CreateAt),
							zap.Error(err))
					} else {
						payTime = parsedTime
					}
				} else {
					// 如果没有订单信息或时间为空，使用当前时间
					payTime = time.Now()
				}

				// 使用通用的支付成功处理函数
				task.handlePaymentSuccess(ctx, payTime, task.MerType, task.OrderUniId)
				return
			} else {
				global.GVA_LOG.Debug("未找到匹配的先付后享支付订单，继续监控",
					zap.String("taskID", task.TaskID),
					zap.String("amount", task.Amount))
			}

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
				// 超时处理：将订单状态设置为失败(state=2)
				ctx := context.Background()
				err := service.ServiceGroupApp.ExampleServiceGroup.UpdateMerPayOrder(ctx, example.MerPayOrder{
					Id:    &task.OrderUniId,
					State: global.MER_PAY_ORDER_FAILED,
				})
				if err != nil {
					global.GVA_LOG.Error("更新订单状态为失败时出错",
						zap.String("taskID", task.TaskID),
						zap.Int64("orderUniId", task.OrderUniId),
						zap.Error(err))
				} else {
					global.GVA_LOG.Info("订单监控超时，已将订单状态设置为失败",
						zap.String("taskID", task.TaskID),
						zap.Int64("orderUniId", task.OrderUniId),
						zap.String("amount", task.Amount),
						zap.Duration("超时时长", task.TTL))
				}

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

// PaymentCallbackData 支付回调数据结构
type PaymentCallbackData struct {
	OrderId       string    `json:"orderId"`       // 订单ID
	TransactionId int64     `json:"transactionId"` // 交易ID
	Amount        float64   `json:"amount"`        // 支付金额
	PayTime       time.Time `json:"payTime"`       // 支付时间
	PaymentMethod string    `json:"paymentMethod"` // 支付方式
	Status        string    `json:"status"`        // 支付状态
}

// handlePaymentSuccess 处理支付成功的通用逻辑，包括更新订单状态和发送回调
func (task *OrderMonitorTask) handlePaymentSuccess(ctx context.Context, payTime time.Time, merType string, transactionId int64) {
	// 更新订单状态为已支付
	err := service.ServiceGroupApp.ExampleServiceGroup.UpdateMerPayOrder(ctx, example.MerPayOrder{
		Id:      &task.OrderUniId,
		State:   global.MER_PAY_ORDER_PAID,
		PayTime: &payTime,
	})
	if err != nil {
		global.GVA_LOG.Error("更新支付订单失败",
			zap.String("taskID", task.TaskID),
			zap.String("paymentMethod", merType),
			zap.Int64("merUserId", task.MerUserId),
			zap.Error(err))
		return
	}

	// 发送支付成功回调
	if task.CallBackUrl != "" {
		amount, _ := strconv.ParseFloat(task.Amount, 64)
		callbackData := PaymentCallbackData{
			OrderId:       task.OrderId,
			Amount:        amount,
			PayTime:       payTime,
			PaymentMethod: merType,
			Status:        "success",
			TransactionId: transactionId,
		}

		// 异步发送回调，避免阻塞主流程
		go func() {
			err := sendPaymentCallback(task.CallBackUrl, callbackData)
			if err != nil {
				global.GVA_LOG.Error("支付回调发送失败",
					zap.String("merType", merType),
					zap.String("orderId", task.OrderId),
					zap.String("callbackUrl", task.CallBackUrl),
					zap.Error(err))
			}
		}()
	}

	global.GVA_LOG.Info("找到匹配的支付订单，停止监控任务",
		zap.String("taskID", task.TaskID),
		zap.String("merType", merType),
		zap.String("amount", task.Amount))
	task.Stop()
}

// sendPaymentCallback 发送支付成功回调
func sendPaymentCallback(callbackUrl string, paymentData PaymentCallbackData) error {
	if callbackUrl == "" {
		global.GVA_LOG.Info("回调URL为空，跳过回调")
		return nil
	}

	// 设置请求头
	headers := map[string]string{
		"Content-Type": "application/json",
		"User-Agent":   "Four-Pay-System/1.0",
	}

	// 发送POST请求
	resp, err := request.HttpRequest(callbackUrl, "POST", headers, nil, paymentData)
	if err != nil {
		global.GVA_LOG.Error("发送支付回调失败", zap.String("url", callbackUrl), zap.Error(err))
		return fmt.Errorf("发送回调请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		global.GVA_LOG.Info("支付回调发送成功",
			zap.String("url", callbackUrl),
			zap.Int("statusCode", resp.StatusCode),
			zap.String("orderId", paymentData.OrderId))
	} else {
		global.GVA_LOG.Warn("支付回调响应状态异常",
			zap.String("url", callbackUrl),
			zap.Int("statusCode", resp.StatusCode),
			zap.String("orderId", paymentData.OrderId))
		return fmt.Errorf("回调响应状态码异常: %d", resp.StatusCode)
	}

	return nil
}
