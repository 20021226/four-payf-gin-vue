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
	"sync"
	"time"

	"go.uber.org/zap"
)

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
