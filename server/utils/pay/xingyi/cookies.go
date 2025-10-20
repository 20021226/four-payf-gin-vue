package xingyi

// Cookies 表示需要的 Cookie 项，支持生成 map 并应用到客户端
type Cookies struct {
	CookieSession1 string `json:"cookiesession1"`
	AcwTc          string `json:"acw_tc"`
	JSessionID     string `json:"JSESSIONID"`
}

// ToMap 将结构体转换为 Cookie 键值对，键名与服务端一致
func (c Cookies) ToMap() map[string]string {
	return map[string]string{
		"acw_tc":         c.AcwTc,
		"cookiesession1": c.CookieSession1,
		"JSESSIONID":     c.JSessionID,
	}
}
