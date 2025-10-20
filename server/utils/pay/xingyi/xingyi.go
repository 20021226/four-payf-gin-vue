package xingyi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"go.uber.org/zap"
)

// PayListResponse 支付列表响应结构
type PayListResponse struct {
	RSPCOD  string     `json:"RSPCOD"`
	RSPMSG  string     `json:"RSPMSG"`
	ROWLIST []OrderRow `json:"ROWLIST"`
	TOTAL   string     `json:"TOTAL"`
}

// OrderRow 订单行数据结构
type OrderRow struct {
	OrderTime   string `json:"ORDER_TIME"`   // 订单时间，格式：2025-10-08 17:04:35
	RecTxamt    string `json:"REC_TXAMT"`    // 实收金额
	OrderNo     string `json:"ORDER_NO"`     // 订单号
	OrderStatus string `json:"ORDER_STATUS"` // 订单状态
	PayChannel  string `json:"PAY_CHANNEL"`  // 支付渠道
	Txamt       string `json:"TXAMT"`        // 交易金额
}

// redisCookieKey 生成具体的 Redis 键名
func redisCookieKey(merId string) string {
	return fmt.Sprintf("%s:%s", global.REDIS_PAY_REQUEST_COOKIE, merId)
}

// Service orchestrates xingyi flows
type Service struct {
	c       *Client
	ocr     OCRProvider
	cookies Cookies
}

func NewService(c *Client, o OCRProvider, cookies Cookies) *Service {
	if c == nil {
		c = NewClient()
	}
	if o == nil {
		o = NewXingyiOCR(Config{})
	}
	return &Service{c: c, ocr: o, cookies: cookies}
}

// GetCookies simulates GET to acquire initial cookies
func (s *Service) GetCookies() error {
	// 准备 headers
	headers := map[string]string{
		"accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
		"accept-language":           "zh-CN,zh;q=0.9,en;q=0.8",
		"cache-control":             "max-age=0",
		"if-modified-since":         "Thu, 10 Jul 2025 13:02:00 GMT",
		"priority":                  "u=0, i",
		"referer":                   "https://xypc.postar.cn/index.html",
		"sec-ch-ua":                 "\"Microsoft Edge\";v=\"141\", \"Not?A_Brand\";v=\"8\", \"Chromium\";v=\"141\"",
		"sec-ch-ua-mobile":          "?0",
		"sec-ch-ua-platform":        "\"Windows\"",
		"sec-fetch-dest":            "document",
		"sec-fetch-mode":            "navigate",
		"sec-fetch-site":            "same-origin",
		"sec-fetch-user":            "?1",
		"upgrade-insecure-requests": "1",
		"user-agent":                "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/141.0.0.0 Safari/537.36 Edg/141.0.0.0",
	}

	// 准备 cookies（初始为空）
	cookies := GetDefaultCookies()

	body, resp, err := s.c.Get("https://xypc.postar.cn/images/login-iocn/pic_bg_xyf.webp", headers, cookies)
	_ = body
	if err != nil {
		return err
	}

	// 构建 Cookies 结构体并解析响应中的 Set-Cookie
	for _, ck := range resp.Cookies() {
		// 直接从响应 cookies 中提取需要的值
		switch ck.Name {
		case "cookiesession1":
			s.cookies.CookieSession1 = ck.Value
		case "acw_tc":
			s.cookies.AcwTc = ck.Value
		case "JSESSIONID":
			s.cookies.JSessionID = ck.Value
		}
	}

	return nil
}

// GetVerifyCode requests captcha and runs OCR
func (s *Service) GetVerifyCode() (string, error) {
	// 准备 headers
	headers := map[string]string{
		"accept":             "application/json, text/javascript, */*; q=0.01",
		"accept-language":    "zh-CN,zh;q=0.9,en;q=0.8",
		"content-length":     "0",
		"origin":             "https://xypc.postar.cn",
		"priority":           "u=1, i",
		"referer":            "https://xypc.postar.cn/login.html",
		"sec-ch-ua":          "\"Microsoft Edge\";v=\"141\", \"Not?A_Brand\";v=\"8\", \"Chromium\";v=\"141\"",
		"sec-ch-ua-mobile":   "?0",
		"sec-ch-ua-platform": "\"Windows\"",
		"sec-fetch-dest":     "empty",
		"sec-fetch-mode":     "cors",
		"sec-fetch-site":     "same-origin",
		"user-agent":         "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/141.0.0.0 Safari/537.36 Edg/141.0.0.0",
		"x-requested-with":   "XMLHttpRequest",
	}

	// 将 Cookies 结构体转换为 map

	// POST to 720003.mer
	body, resp, err := s.c.PostForm("https://xypc.postar.cn/720003.mer", map[string]string{}, headers, s.cookies.ToMap())
	if err != nil {
		return "", err
	}

	//for _, ck := range resp.Cookies() {
	//	// 直接从响应 cookies 中提取需要的值
	//	switch ck.Name {
	//	case "cookiesession1":
	//		cookies.CookieSession1 = ck.Value
	//	case "acw_tc":
	//		cookies.AcwTc = ck.Value
	//	case "JSESSIONID":
	//		cookies.JSessionID = ck.Value
	//	}
	//}

	// 更新 cookies（如果需要的话，这里可以返回更新后的 cookies）
	_ = resp

	// parse JSON to get CODE_PIC
	var j struct {
		CodePic string `json:"CODE_PIC"`
	}
	if err := ParseJSON(body, &j); err != nil {
		return "", err
	}
	// OCR
	vc, err := s.ocr.Recognize(j.CodePic)
	if err != nil {
		return "", err
	}
	return vc, nil
}

// GetLoginResult performs encrypted login and returns PWD_RESPDATA
func (s *Service) GetLoginResult(username, password, merId string) (string, error) {
	// ensure cookies fetched
	err := s.GetCookies()
	if err != nil {
		return "", err
	}
	vc, err := s.GetVerifyCode()
	fmt.Println("识别到的验证码:", vc)
	if err != nil {
		return "", err
	}
	// Build src string
	now := time.Now().Unix()
	// Python used lower() on VARI_CODE
	src := fmt.Sprintf("CUST_LOGIN=%s&VARI_CODE=%s&NOW_TIME=%d&SYSCOD=3000&USRPWD=", username, strings.ToLower(vc), now-100)

	// encrypt and urlencode
	encSrc, err := EncryptAES_ECB_PKCS7(src)
	if err != nil {
		return "", err
	}
	encSrc = url.QueryEscape(encSrc)
	encPwd, err := EncryptAES_ECB_PKCS7(password)
	if err != nil {
		return "", err
	}
	encPwd = url.QueryEscape(encPwd)
	form := map[string]string{
		"key":      "3DD892F76F0461549B67FA00",
		"srcStr":   encSrc,
		"NOW_TIME": fmt.Sprintf("%d", now),
		"usrpwd":   encPwd,
	}

	// 准备 headers 和 cookies
	headers := map[string]string{
		"accept":             "application/json, text/javascript, */*; q=0.01",
		"accept-language":    "zh-CN,zh;q=0.9,en;q=0.8",
		"content-type":       "application/x-www-form-urlencoded",
		"origin":             "https://xypc.postar.cn",
		"priority":           "u=0, i",
		"referer":            "https://xypc.postar.cn/login.html",
		"sec-ch-ua":          "\"Microsoft Edge\";v=\"141\", \"Not?A_Brand\";v=\"8\", \"Chromium\";v=\"141\"",
		"sec-ch-ua-mobile":   "?0",
		"sec-ch-ua-platform": "\"Windows\"",
		"sec-fetch-dest":     "empty",
		"sec-fetch-mode":     "cors",
		"sec-fetch-site":     "same-origin",
		"user-agent":         "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/141.0.0.0 Safari/537.36 Edg/141.0.0.0",
		"x-requested-with":   "XMLHttpRequest",
	}

	body, resp, err := s.c.PostForm("https://xypc.postar.cn/720005.mer", form, headers, s.cookies.ToMap())
	if err != nil {
		return "", fmt.Errorf("HTTP 请求失败: %w", err)
	}

	// 记录响应状态和内容用于调试
	println("GetLoginResult HTTP 响应", body)

	// 检查 HTTP 状态码
	if resp.StatusCode != 200 {
		println("HTTP 响应状态码异常", string(body))
		return "", fmt.Errorf("HTTP 状态码异常: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 构建 Cookies 结构体并合并响应中的 Cookie
	for _, ck := range resp.Cookies() {
		// 直接从响应 cookies 中提取需要的值
		switch ck.Name {
		case "cookiesession1":
			s.cookies.CookieSession1 = ck.Value
		case "acw_tc":
			s.cookies.AcwTc = ck.Value
		case "JSESSIONID":
			s.cookies.JSessionID = ck.Value
		}
	}
	// 仅当 Redis 已初始化时写入
	if global.GVA_REDIS != nil {
		b, _ := json.Marshal(s.cookies)
		_ = global.GVA_REDIS.Set(context.Background(), redisCookieKey(merId), string(b), 0).Err()
	}

	// 检查响应内容
	bodyStr := string(body)
	println(bodyStr)
	if strings.Contains(bodyStr, "<html>") || strings.Contains(bodyStr, "登录") {
		global.GVA_LOG.Error("GetLoginResult 响应包含 HTML 内容",
			zap.String("merId", merId),
			zap.String("body", bodyStr))
		return bodyStr, fmt.Errorf("响应包含 HTML 内容，可能是登录失败")
	}

	var j map[string]any
	if err := ParseJSON(body, &j); err != nil {
		global.GVA_LOG.Warn("GetLoginResult JSON 解析失败，返回原始响应",
			zap.String("merId", merId),
			zap.Error(err),
			zap.String("body", bodyStr))
		return bodyStr, nil
	}

	if v, ok := j["PWD_RESPDATA"].(string); ok {
		println("成功获取 PWD_RESPDATA", merId)
		return v, nil
	}

	global.GVA_LOG.Warn("响应中未找到 PWD_RESPDATA",
		zap.String("merId", merId),
		zap.Any("response", j))
	return bodyStr, nil
}

// GetAccessToken replicates get_at: submit sText to get ACCESS_TOKEN
func (s *Service) GetAccessToken(sText string, merId string) (string, error) {
	// 在请求之前，按 merId 从 Redis 读取 Cookie，确保使用同一会话
	if global.GVA_REDIS != nil {
		val, err := global.GVA_REDIS.Get(context.Background(), redisCookieKey(merId)).Result()
		if err == nil && val != "" {
			if jsonErr := json.Unmarshal([]byte(val), &s.cookies); jsonErr == nil {
				global.GVA_LOG.Info("成功从 Redis 读取 cookie",
					zap.String("merId", merId),
					zap.String("cookieKey", redisCookieKey(merId)))
			} else {
				global.GVA_LOG.Warn("解析 cookie JSON 失败",
					zap.String("merId", merId),
					zap.Error(jsonErr))
				return "", jsonErr
			}
		} else {
			global.GVA_LOG.Warn("从 Redis 读取 cookie 失败",
				zap.String("merId", merId),
				zap.String("cookieKey", redisCookieKey(merId)),
				zap.Error(err))
			return "", fmt.Errorf("从 Redis 读取 cookie 失败: %w", err)

		}
	}

	encryKey := "NTM0NzJGNjM2NTMzNDgzOTc1NEQzOTY5MzE1QTZDNEUyQjQxNjU0NjMzNjI0NTY0NkI3MTczNTc0RDM1NEE0RDM1NzI0ODU0MzYyQjQzNkI0MjZCNTM1NDU1Nzk1NzZCNzI0MjQ4MkI2NTczNTQ0RjU4MkY2RDMyNTI0RTZCNzYzMzRBNkMzNjMxNDE0RTc1NDQ0Rjc3NkE0RTc3Njk0NDc3NDU0RjM0NDk2RjMxMzQ0Nzc5NDI2RjM5NEM0QzZFNEQ2QTUzNzk3NjU1NkE3ODMzNjEzNDRCN0E1NDJGNDI1NTRBNjI1MTREMzI2NjMzMkI1OTY3NTQ0RDQ0NTg0RDZDNTM3MDQzNkM3MTY1NTQ1QTY5NkI2QTQ0NDg3NjVBMzg0QzYyNzY3MTZGMkI0MTQ3MzYyRjQyNzI1QTZENTE0ODYyN0E2RDdBNkM1NjQ1NEM3NjZGM0Q="
	// urlencode sText
	escaped := url.QueryEscape(sText)
	form := map[string]string{
		"encryKey": encryKey,
		"sText":    escaped,
		"sPwds":    "",
		"HY_TYPE":  "01",
	}

	// 准备 headers 和 cookies
	headers := GetDefaultHeaders()
	cookieMap := s.cookies.ToMap()
	fmt.Println(s.cookies.ToMap())

	body, resp, err := s.c.PostForm("https://xypc.postar.cn/720001.mer", form, headers, cookieMap)
	if err != nil {
		return "", fmt.Errorf("HTTP 请求失败: %w", err)
	}

	// 记录响应状态和内容用于调试
	global.GVA_LOG.Info("GetAccessToken HTTP 响应",
		zap.String("merId", merId),
		zap.Int("statusCode", resp.StatusCode),
		zap.String("contentType", resp.Header.Get("Content-Type")),
		zap.Int("bodyLength", len(body)))

	// 检查 HTTP 状态码
	if resp.StatusCode != 200 {
		global.GVA_LOG.Error("HTTP 响应状态码异常",
			zap.String("merId", merId),
			zap.Int("statusCode", resp.StatusCode),
			zap.String("body", string(body)))
		return "", fmt.Errorf("HTTP 状态码异常: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 检查响应内容类型
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") && !strings.Contains(string(body), "{") {
		global.GVA_LOG.Error("响应不是 JSON 格式",
			zap.String("merId", merId),
			zap.String("contentType", contentType),
			zap.String("body", string(body)))
		return "", fmt.Errorf("服务器返回非 JSON 响应，可能是会话过期或重定向: %s", string(body))
	}

	// 检查是否包含登录相关的 HTML 内容
	bodyStr := string(body)
	if strings.Contains(bodyStr, "<html>") || strings.Contains(bodyStr, "登录") || strings.Contains(bodyStr, "login") {
		global.GVA_LOG.Error("响应包含登录页面内容，会话可能已过期",
			zap.String("merId", merId),
			zap.String("body", bodyStr))
		return "", fmt.Errorf("会话已过期，需要重新登录")
	}

	var j map[string]any
	if err := ParseJSON(body, &j); err != nil {
		global.GVA_LOG.Error("JSON 解析失败",
			zap.String("merId", merId),
			zap.Error(err),
			zap.String("body", string(body)))
		return "", fmt.Errorf("JSON 解析失败: %w, 响应内容: %s", err, string(body))
	}

	if v, ok := j["ACCESS_TOKEN"].(string); ok {
		tokenPreview := v
		if len(v) > 20 {
			tokenPreview = v[:20] + "..."
		}
		global.GVA_LOG.Info("成功获取 ACCESS_TOKEN",
			zap.String("merId", merId),
			zap.String("token", tokenPreview))
		return v, nil
	}

	global.GVA_LOG.Error("响应中未找到 ACCESS_TOKEN",
		zap.String("merId", merId),
		zap.Any("response", j))
	return "", fmt.Errorf("ACCESS_TOKEN not found: %s", string(body))
}

// GetPayList replicates mergeCodeAndCardTradeList2.mers query
func (s *Service) GetPayList(token string, startTime string, endTime string, merId string) ([]byte, error) {
	// 在请求之前，按 merId 从 Redis 读取 Cookie，确保使用同一会话
	if global.GVA_REDIS != nil {
		val, err := global.GVA_REDIS.Get(context.Background(), redisCookieKey(merId)).Result()
		if err == nil && val != "" {
			if jsonErr := json.Unmarshal([]byte(val), &s.cookies); jsonErr != nil {
				return []byte{}, jsonErr
			}
		}
	}
	form := map[string]string{
		"PAGENUM":                "1",
		"NUMPERPAG":              "10",
		"ORDER_NO":               "",
		"CUST_ID":                "",
		"BUS_CUST_ID":            "",
		"BUS_NAME":               "",
		"PAY_CHANNEL":            "",
		"ORDER_STATUS2":          "1",
		"TER_CODE":               "",
		"CODE_NO":                "",
		"TRAN_TYPE":              "",
		"T_SALE_NAME":            "",
		"MIN_AMT":                "",
		"MAX_AMT":                "",
		"IS_CHAIN":               "",
		"CHAIN_CUST_IDS":         "",
		"T_PAY_NO":               "",
		"RE_MERCID":              "",
		"WB_ID":                  "",
		"THREE_ORDER_NO":         "",
		"TXNLOGID":               "",
		"OPEN_ID":                "",
		"OLD_ORDER_NO":           "",
		"T_ORDER_NO":             "",
		"code_name_id":           "",
		"code_cust_id":           "",
		"IS_TERMFEE":             "",
		"TRANSACTION_BEGIN_TIME": startTime,
		"TRANSACTION_END_TIME":   endTime,
		"ACCESS_TOKEN":           token,
	}

	// 准备 headers 和 cookies
	headers := GetDefaultHeaders()

	body, _, err := s.c.PostForm("https://xypc.postar.cn/mergeCodeAndCardTradeList2.mers", form, headers, s.cookies.ToMap())
	if err != nil {
		return nil, err
	}
	return body, nil
}

// GetCheckResult 检查支付列表中是否有符合条件的订单
// 检查条件：1. ORDER_TIME 在 startTime 和 endTime 之间  2. amount 等于 REC_TXAMT
func (s *Service) GetCheckResult(body string, startTime, endTime time.Time, amount string) (bool, error) {
	// 先检查 JSON 中是否包含 ROWLIST 字段
	if !strings.Contains(body, "ROWLIST") {
		global.GVA_LOG.Warn("JSON 数据中缺少 ROWLIST 字段", zap.String("body", body))
		return false, nil
	}

	// 解析 JSON 数据
	var response PayListResponse
	if err := json.Unmarshal([]byte(body), &response); err != nil {
		global.GVA_LOG.Error("解析支付列表 JSON 失败", zap.Error(err), zap.String("body", body))
		return false, fmt.Errorf("解析 JSON 失败: %w", err)
	}

	// 检查响应状态
	if response.RSPCOD != "00000" {
		global.GVA_LOG.Warn("支付列表查询失败",
			zap.String("RSPCOD", response.RSPCOD),
			zap.String("RSPMSG", response.RSPMSG))
		return false, fmt.Errorf("查询失败: %s", response.RSPMSG)
	}

	global.GVA_LOG.Info("开始检查支付列表",
		zap.Int("订单总数", len(response.ROWLIST)),
		zap.String("目标金额", amount),
		zap.Time("开始时间", startTime),
		zap.Time("结束时间", endTime))

	// 遍历订单列表进行检查
	for i, order := range response.ROWLIST {
		global.GVA_LOG.Debug("检查订单",
			zap.Int("索引", i),
			zap.String("订单号", order.OrderNo),
			zap.String("订单时间", order.OrderTime),
			zap.String("实收金额", order.RecTxamt),
			zap.String("订单状态", order.OrderStatus))

		// 1. 检查金额是否匹配
		if !s.checkAmount(order.RecTxamt, amount) {
			global.GVA_LOG.Debug("金额不匹配",
				zap.String("订单号", order.OrderNo),
				zap.String("实收金额", order.RecTxamt),
				zap.String("目标金额", amount))
			continue
		}

		// 2. 检查时间是否在范围内
		if !s.checkTimeRange(order.OrderTime, startTime, endTime) {
			global.GVA_LOG.Debug("时间不在范围内",
				zap.String("订单号", order.OrderNo),
				zap.String("订单时间", order.OrderTime))
			continue
		}

		// 找到符合条件的订单
		global.GVA_LOG.Info("找到符合条件的订单",
			zap.String("订单号", order.OrderNo),
			zap.String("订单时间", order.OrderTime),
			zap.String("实收金额", order.RecTxamt),
			zap.String("订单状态", order.OrderStatus),
			zap.String("支付渠道", order.PayChannel))

		return true, nil
	}

	global.GVA_LOG.Info("未找到符合条件的订单",
		zap.String("目标金额", amount),
		zap.Time("开始时间", startTime),
		zap.Time("结束时间", endTime))

	return false, nil
}

// checkAmount 检查金额是否匹配
func (s *Service) checkAmount(recTxamt, targetAmount string) bool {
	// 直接比较字符串
	return recTxamt == targetAmount
}

// checkTimeRange 检查时间是否在指定范围内
func (s *Service) checkTimeRange(orderTimeStr string, startTime, endTime time.Time) bool {
	// 解析订单时间，格式：2025-10-08 17:04:35
	orderTime, err := time.Parse("2006-01-02 15:04:05", orderTimeStr)
	if err != nil {
		global.GVA_LOG.Error("解析订单时间失败",
			zap.String("orderTime", orderTimeStr),
			zap.Error(err))
		return false
	}

	// 检查时间是否在范围内（包含边界）
	return (orderTime.Equal(startTime) || orderTime.After(startTime)) &&
		(orderTime.Equal(endTime) || orderTime.Before(endTime))
}

// GetPayList replicates mergeCodeAndCardTradeList2.mers query
func (s *Service) CheckToken(token string, merId string) bool {
	body, err := s.GetPayList(token, "20251008170435", "20251008170435", merId, Cookies{})
	if err != nil {
		global.GVA_LOG.Error("获取支付列表失败", zap.Error(err))
		return false
	}
	if strings.Contains(string(body), "登录超时，请重新登录") {
		return false
	}
	if strings.Contains(string(body), "异地登录") {
		return false
	}
	return true
}
