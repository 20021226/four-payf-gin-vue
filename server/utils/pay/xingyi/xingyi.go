package xingyi

import (
    "context"
    "encoding/json"
    "fmt"
    "net/url"
    "strings"
    "time"

    "github.com/flipped-aurora/gin-vue-admin/server/global"
)

// RedisCookieKeyPrefix 统一的 Redis 键前缀，可集中修改
const RedisCookieKeyPrefix = "xingyi_cookie"

// redisCookieKey 生成具体的 Redis 键名
func redisCookieKey(merId string) string {
    return fmt.Sprintf("%s:%s", RedisCookieKeyPrefix, merId)
}

// Service orchestrates xingyi flows
type Service struct {
	c   *Client
	ocr OCRProvider
}

func NewService(c *Client, o OCRProvider) *Service {
	if c == nil {
		c = NewClient()
	}
	if o == nil {
		o = NoopOCR{}
	}
	return &Service{c: c, ocr: o}
}

// GetCookies simulates GET to acquire initial cookies
func (s *Service) GetCookies() (map[string]string, error) {
	s.c.SetHeader("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	s.c.SetHeader("referer", "https://xypc.postar.cn/index.html")
	s.c.SetHeader("sec-fetch-dest", "document")
	s.c.SetHeader("sec-fetch-mode", "navigate")
	s.c.SetHeader("sec-fetch-site", "same-origin")
	s.c.SetHeader("sec-fetch-user", "?1")
	s.c.SetHeader("upgrade-insecure-requests", "1")
	body, resp, err := s.c.Get("https://xypc.postar.cn/bankLogin.html")
	_ = body
	if err != nil {
		return nil, err
	}
	// Parse Set-Cookie from response headers and merge
	for _, ck := range resp.Cookies() {
		s.c.SetCookie(ck.Name, ck.Value)
	}
	return s.c.Cookies, nil
}

// GetVerifyCode requests captcha and runs OCR
func (s *Service) GetVerifyCode() (string, error) {
	// Minimal headers for captcha endpoint
	s.c.SetHeader("accept", "application/json, text/javascript, */*; q=0.01")
	s.c.SetHeader("origin", "https://xypc.postar.cn")
	s.c.SetHeader("referer", "https://xypc.postar.cn/login.html")
	// POST to 720003.mer
	body, resp, err := s.c.PostForm("https://xypc.postar.cn/720003.mer", map[string]string{})
	if err != nil {
		return "", err
	}
	// merge response cookies
	for _, ck := range resp.Cookies() {
		s.c.SetCookie(ck.Name, ck.Value)
	}
	// parse JSON to get CODE_PIC
	var j struct {
		CodePic string `json:"CODE_PIC"`
	}
	if err := ParseJSON(body, &j); err != nil {
		return "", err
	}
	// OCR
	return s.ocr.Recognize(j.CodePic)
}

// GetLoginResult performs encrypted login and returns PWD_RESPDATA
func (s *Service) GetLoginResult(username, password, merId string) (string, error) {
	// ensure cookies fetched
	if _, err := s.GetCookies(); err != nil {
		return "", err
	}
	vc, err := s.GetVerifyCode()
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
    body, resp, err := s.c.PostForm("https://xypc.postar.cn/720005.mer", form)
    if err != nil {
        return "", err
    }
    // 合并响应中的 Cookie 到客户端，以保持本次会话
    for _, ck := range resp.Cookies() {
        s.c.SetCookie(ck.Name, ck.Value)
    }

    // 将本次会话 Cookie 保存到 Redis，键：xingyi_cookie:<merId>
    // 使用 Cookies 结构体统一存储格式
    cookies := Cookies{
        CookieSession1: s.c.Cookies["$cookiesession1"],
        AcwTc:          s.c.Cookies["acw_tc"],
        JSessionID:     s.c.Cookies["JSESSIONID"],
    }
    // 仅当 Redis 已初始化时写入
    if global.GVA_REDIS != nil {
        b, _ := json.Marshal(cookies)
        _ = global.GVA_REDIS.Set(context.Background(), redisCookieKey(merId), string(b), 0).Err()
    }
	var j map[string]any
	if err := ParseJSON(body, &j); err != nil {
		return string(body), nil
	}
	if v, ok := j["PWD_RESPDATA"].(string); ok {
		return v, nil
	}
	return string(body), nil
}

// GetAccessToken replicates get_at: submit sText to get ACCESS_TOKEN
func (s *Service) GetAccessToken(sText string) (string, error) {
	encryKey := "NTM0NzJGNjM2NTMzNDgzOTc1NEQzOTY5MzE1QTZDNEUyQjQxNjU0NjMzNjI0NTY0NkI3MTczNTc0RDM1NEE0RDM1NzI0ODU0MzYyQjQzNkI0MjZCNTM1NDU1Nzk1NzZCNzI0MjQ4MkI2NTczNTQ0RjU4MkY2RDMyNTI0RTZCNzYzMzRBNkMzNjMxNDE0RTc1NDQ0Rjc3NkE0RTc3Njk0NDc3NDU0RjM0NDk2RjMxMzQ0Nzc5NDI2RjM5NEM0QzZFNEQ2QTUzNzk3NjU1NkE3ODMzNjEzNDRCN0E1NDJGNDI1NTRBNjI1MTREMzI2NjMzMkI1OTY3NTQ0RDQ0NTg0RDZDNTM3MDQzNkM3MTY1NTQ1QTY5NkI2QTQ0NDg3NjVBMzg0QzYyNzY3MTZGMkI0MTQ3MzYyRjQyNzI1QTZENTE0ODYyN0E2RDdBNkM1NjQ1NEM3NjZGM0Q="
	// urlencode sText
	escaped := url.QueryEscape(sText)
	form := map[string]string{
		"encryKey": encryKey,
		"sText":    escaped,
		"sPwds":    "",
		"HY_TYPE":  "01",
	}
	body, _, err := s.c.PostForm("https://xypc.postar.cn/720001.mer", form)
	if err != nil {
		return "", err
	}
	var j map[string]any
	if err := ParseJSON(body, &j); err != nil {
		return "", err
	}
	if v, ok := j["ACCESS_TOKEN"].(string); ok {
		return v, nil
	}
	return "", fmt.Errorf("ACCESS_TOKEN not found: %s", string(body))
}

// GetPayList replicates mergeCodeAndCardTradeList2.mers query
func (s *Service) GetPayList(token string, startTime string, endTime string, merId string) ([]byte, error) {
    // 在请求之前，按 merId 从 Redis 读取并应用 Cookie，确保使用同一会话
    if global.GVA_REDIS != nil {
        val, err := global.GVA_REDIS.Get(context.Background(), redisCookieKey(merId)).Result()
        if err == nil && val != "" {
            var cookies Cookies
            if jsonErr := json.Unmarshal([]byte(val), &cookies); jsonErr == nil {
                cookies.ApplyTo(s.c)
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
	body, _, err := s.c.PostForm("https://xypc.postar.cn/mergeCodeAndCardTradeList2.mers", form)
	if err != nil {
		return nil, err
	}
	return body, nil
}
