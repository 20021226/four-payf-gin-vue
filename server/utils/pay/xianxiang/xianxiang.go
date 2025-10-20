package xianxiang

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/shopspring/decimal"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"go.uber.org/zap"
)

// LoginResult 登录结果结构体
type LoginResult struct {
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	AccessToken string `json:"access_token"`
}

// OrderListResponse 订单列表响应结构
type OrderListResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Data []OrderItem `json:"data"`
		Meta struct {
			Total       int `json:"total"`
			PerPage     int `json:"per_page"`
			CurrentPage int `json:"current_page"`
			LastPage    int `json:"last_page"`
		} `json:"meta"`
	} `json:"data"`
}

// OrderItem 订单项结构
type OrderItem struct {
	ID           int             `json:"id"`
	OrderSn      string          `json:"order_sn"`
	Amount       decimal.Decimal `json:"price"`
	RealAmount   decimal.Decimal `json:"real_price"`
	Status       int             `json:"status"`
	PayWay       string          `json:"pay_way"`
	CreateAt     string          `json:"create_at"`
	UpdateAt     string          `json:"update_at"`
	PayTime      string          `json:"pay_time"`
	NotifyStatus int             `json:"notify_status"`
}

// Service 服务结构体
type Service struct {
	c *Client
}

// NewService 创建新的服务实例
func NewService(c *Client) *Service {
	if c == nil {
		c = NewClient()
	}
	return &Service{c: c}
}

// safeLog 安全日志函数，如果 global.GVA_LOG 为 nil 则使用 fmt.Printf
func safeLog(level string, msg string, args ...interface{}) {
	if global.GVA_LOG != nil {
		switch level {
		case "info":
			global.GVA_LOG.Info(msg, convertToZapFields(args...)...)
		case "warn":
			global.GVA_LOG.Warn(msg, convertToZapFields(args...)...)
		case "error":
			global.GVA_LOG.Error(msg, convertToZapFields(args...)...)
		case "debug":
			global.GVA_LOG.Debug(msg, convertToZapFields(args...)...)
		}
	} else {
		// 如果 global.GVA_LOG 未初始化，使用 fmt.Printf
		fmt.Printf("[%s] %s", strings.ToUpper(level), msg)
		for i := 0; i < len(args); i += 2 {
			if i+1 < len(args) {
				fmt.Printf(" %v=%v", args[i], args[i+1])
			}
		}
		fmt.Println()
	}
}

// convertToZapFields 将参数转换为 zap 字段
func convertToZapFields(args ...interface{}) []zap.Field {
	var fields []zap.Field
	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			key := fmt.Sprintf("%v", args[i])
			value := args[i+1]
			fields = append(fields, zap.Any(key, value))
		}
	}
	return fields
}

// logInfo 信息日志
func logInfo(msg string, args ...interface{}) {
	safeLog("info", msg, args...)
}

// logWarn 警告日志
func logWarn(msg string, args ...interface{}) {
	safeLog("warn", msg, args...)
}

// logError 错误日志
func logError(msg string, args ...interface{}) {
	safeLog("error", msg, args...)
}

// logDebug 调试日志
func logDebug(msg string, args ...interface{}) {
	safeLog("debug", msg, args...)
}

// GetToken 获取访问令牌
func (s *Service) GetToken(username, password string) (*LoginResult, error) {
	headers := map[string]string{
		"accept":             "*/*",
		"Authorization":      "Bearer",
		"cache-control":      "no-cache",
		"client":             "mobile",
		"Content-Type":       "application/x-www-form-urlencoded",
		"Origin":             "https://b.isv6.com",
		"Pragma":             "no-cache",
		"Referer":            "https://b.isv6.com/",
		"sec-ch-ua":          `"Microsoft Edge";v="141", "Not?A_Brand";v="8", "Chromium";v="141"`,
		"sec-ch-ua-mobile":   "?0",
		"sec-ch-ua-platform": `"Windows"`,
		"sec-fetch-dest":     "empty",
		"sec-fetch-mode":     "cors",
		"sec-fetch-site":     "same-origin",
		"User-Agent":         "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/141.0.0.0 Safari/537.36 Edg/141.0.0.0",
	}

	url := "https://b.isv6.com/api/api/auth/login"
	data := map[string]string{
		"username": username,
		"password": password,
		"type":     "merchant",
	}

	body, resp, err := s.c.PostForm(url, data, headers, nil)
	if err != nil {
		return nil, fmt.Errorf("HTTP 请求失败: %w", err)
	}

	logInfo("GetToken HTTP 响应",
		"statusCode", resp.StatusCode,
		"contentType", resp.Header.Get("Content-Type"),
		"bodyLength", len(body))

	if resp.StatusCode != 200 {
		logError("HTTP 响应状态码异常",
			"statusCode", resp.StatusCode,
			"body", string(body))
		return nil, fmt.Errorf("HTTP 状态码异常: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var response struct {
		Code int         `json:"code"`
		Msg  string      `json:"msg"`
		Data LoginResult `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		logError("JSON 解析失败",
			"error", err,
			"body", string(body))
		return nil, fmt.Errorf("JSON 解析失败: %w", err)
	}

	if response.Code != 200 {
		logError("登录失败",
			"code", response.Code,
			"msg", response.Msg)
		return nil, fmt.Errorf("登录失败: %s", response.Msg)
	}

	if response.Data.AccessToken == "" {
		logError("响应中未找到 access_token")
		return nil, fmt.Errorf("响应中未找到 access_token")
	}

	logInfo("成功获取访问令牌",
		"username", username,
		"tokenType", response.Data.TokenType,
		"expiresIn", response.Data.ExpiresIn)

	return &response.Data, nil
}

// GetOrderList 获取订单列表
func (s *Service) GetOrderList(token, orderTime string) ([]byte, error) {
	headers := map[string]string{
		"accept":             "*/*",
		"accept-language":    "zh-CN,zh;q=0.9,en;q=0.8",
		"authorization":      fmt.Sprintf("Bearer %s", token),
		"cache-control":      "no-cache",
		"client":             "mobile",
		"content-type":       "application/x-www-form-urlencoded",
		"origin":             "https://b.isv6.com",
		"pragma":             "no-cache",
		"priority":           "u=1, i",
		"referer":            "https://b.isv6.com/",
		"sec-ch-ua":          `"Microsoft Edge";v="141", "Not?A_Brand";v="8", "Chromium";v="141"`,
		"sec-ch-ua-mobile":   "?0",
		"sec-ch-ua-platform": `"Windows"`,
		"sec-fetch-dest":     "empty",
		"sec-fetch-mode":     "cors",
		"sec-fetch-site":     "same-origin",
		"user-agent":         "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/141.0.0.0 Safari/537.36 Edg/141.0.0.0",
	}

	url := "https://b.isv6.com/api/api/bank/orderList"
	data := map[string]string{
		"type":            "",
		"pay_way":         "",
		"order_sn":        "",
		"page":            "1",
		"perpage":         "10",
		"create_at_start": orderTime,
		"create_at_end":   orderTime,
	}

	body, resp, err := s.c.PostForm(url, data, headers, nil)
	if err != nil {
		return nil, fmt.Errorf("HTTP 请求失败: %w", err)
	}

	logInfo("GetOrderList HTTP 响应",
		"statusCode", resp.StatusCode,
		"contentType", resp.Header.Get("Content-Type"),
		"bodyLength", len(body))

	if resp.StatusCode != 200 {
		logError("HTTP 响应状态码异常",
			"statusCode", resp.StatusCode,
			"body", string(body))
		return nil, fmt.Errorf("HTTP 状态码异常: %d, 响应: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// CheckPayment 检查支付结果
func (s *Service) CheckPayment(body string, startTime, endTime time.Time, amount string) (bool, *OrderItem, error) {
	// 解析响应
	var response OrderListResponse
	if err := json.Unmarshal([]byte(body), &response); err != nil {
		logError("解析订单列表 JSON 失败",
			"error", err,
			"body", body)
		return false, nil, fmt.Errorf("解析 JSON 失败: %w", err)
	}

	if response.Code != 200 {
		logWarn("订单列表查询失败",
			"code", response.Code,
			"msg", response.Msg)
		return false, nil, fmt.Errorf("查询失败: %s", response.Msg)
	}

	logInfo("开始检查订单列表",
		"订单总数", len(response.Data.Data),
		"目标金额", amount,
		"开始时间", startTime,
		"结束时间", endTime)

	// 将目标金额转换为浮点数
	// 将目标金额转换为 decimal.Decimal 类型
	targetAmount, err := decimal.NewFromString(amount)
	if err != nil {
		logError("目标金额格式错误",
			"amount", amount,
			"error", err)
		return false, nil, fmt.Errorf("目标金额格式错误: %w", err)
	}

	// 遍历订单列表进行检查
	for i, order := range response.Data.Data {
		logDebug("检查订单",
			"索引", i,
			"订单号", order.OrderSn,
			"创建时间", order.CreateAt,
			"实收金额", order.RealAmount,
			"订单状态", order.Status)

		// 1. 检查金额是否匹配
		if !order.RealAmount.Equal(targetAmount) {
			logDebug("金额不匹配",
				"订单号", order.OrderSn,
				"实收金额", order.RealAmount,
				"目标金额", targetAmount)
			continue
		}

		// 2. 检查时间是否在范围内
		if !s.checkTimeRange(order.CreateAt, startTime, endTime) {
			logDebug("时间不在范围内",
				"订单号", order.OrderSn,
				"创建时间", order.CreateAt)
			continue
		}

		// 找到符合条件的订单
		logInfo("找到符合条件的订单",
			"订单号", order.OrderSn,
			"创建时间", order.CreateAt,
			"实收金额", order.RealAmount,
			"订单状态", order.Status,
			"支付方式", order.PayWay)

		return true, &order, nil
	}

	logInfo("未找到符合条件的订单",
		"目标金额", amount,
		"开始时间", startTime,
		"结束时间", endTime)

	return false, nil, nil
}

// checkTimeRange 检查时间是否在指定范围内
func (s *Service) checkTimeRange(orderTimeStr string, startTime, endTime time.Time) bool {
	// 解析订单时间，格式：2025-10-13 15:30:45
	orderTime, err := time.Parse("2006-01-02 15:04:05", orderTimeStr)
	if err != nil {
		// 尝试解析日期格式：2025-10-13
		orderTime, err = time.Parse("2006-01-02", orderTimeStr)
		if err != nil {
			logError("解析订单时间失败",
				"orderTime", orderTimeStr,
				"error", err)
			return false
		}
	}

	// 检查时间是否在范围内（包含边界）
	return (orderTime.Equal(startTime) || orderTime.After(startTime)) &&
		(orderTime.Equal(endTime) || orderTime.Before(endTime))
}

// CheckToken 检查令牌是否有效
func (s *Service) CheckToken(token string) bool {
	// 使用当前日期进行测试查询
	today := time.Now().Format("2006-01-02")
	body, err := s.GetOrderList(token, today)
	if err != nil {
		logError("获取订单列表失败", "error", err)
		return false
	}

	// 检查响应是否包含错误信息
	bodyStr := string(body)
	if strings.Contains(bodyStr, "登录超时") || strings.Contains(bodyStr, "token") {
		return false
	}

	return true
}

// redisCookieKey 生成 Redis 键名
func redisCookieKey(merId string) string {
	return fmt.Sprintf("%s:%s", global.REDIS_PAY_REQUEST_COOKIE, merId)
}

// redisTokenKey 生成 Redis token 键名
func redisTokenKey(merId string) string {
	return fmt.Sprintf("%s:%s:%s", global.REDIS_PAY_REQUEST_TOKEN, global.MER_TYPE_XIANG_XIAN, merId)
}

// SaveToken 保存令牌到 Redis
func (s *Service) SaveToken(token string, merId string, expiresIn int) error {
	if global.GVA_REDIS == nil {
		return fmt.Errorf("Redis 未初始化")
	}

	key := redisTokenKey(merId)
	duration := time.Duration(expiresIn) * time.Second

	err := global.GVA_REDIS.Set(context.Background(), key, token, duration).Err()
	if err != nil {
		logError("保存令牌到 Redis 失败",
			"key", key,
			"error", err)
		return err
	}

	logInfo("成功保存令牌到 Redis",
		"key", key,
		"expiresIn", expiresIn)

	return nil
}

// GetTokenFromRedis 从 Redis 获取令牌
func (s *Service) GetTokenFromRedis(merId string) (string, error) {
	if global.GVA_REDIS == nil {
		return "", fmt.Errorf("Redis 未初始化")
	}

	key := redisTokenKey(merId)
	token, err := global.GVA_REDIS.Get(context.Background(), key).Result()
	if err != nil {
		logWarn("从 Redis 获取令牌失败",
			"key", key,
			"error", err)
		return "", err
	}

	logInfo("成功从 Redis 获取令牌",
		"key", key)

	return token, nil
}
