package utils

import "strings"

// IsHostAllowed 判断请求的 Host/域名/IP 是否包含在配置字符串中
// 配置字符串为逗号分隔的域名或IP，允许带端口，例如: "example.com,api.example.com:8080,192.168.1.10"
func IsHostAllowed(allowStr, reqHost string) bool {
	if allowStr == "" || reqHost == "" {
		return false
	}
	hostOnly := reqHost
	if h, _, ok := strings.Cut(reqHost, ":"); ok {
		hostOnly = h
	}
	for _, token := range strings.Split(allowStr, ";") {
		t := strings.TrimSpace(token)
		if t == "" {
			continue
		}
		if t == reqHost || t == hostOnly {
			return true
		}
	}
	return false
}
