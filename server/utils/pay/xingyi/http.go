package xingyi

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"
)

// GetDefaultHeaders 返回默认的 headers
func GetDefaultHeaders() map[string]string {
	return map[string]string{
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
}

// GetDefaultCookies 返回默认的 cookies
func GetDefaultCookies() map[string]string {
	return map[string]string{}
}

// Client wraps http.Client without state
type Client struct {
	hc *http.Client
}

func NewClient() *Client {
	return &Client{
		hc: &http.Client{Timeout: 20 * time.Second},
	}
}

func (c *Client) addHeaders(req *http.Request, headers map[string]string, cookies map[string]string) {
	// 添加传入的 headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	// 添加 cookie header
	if len(cookiesToHeader(cookies)) > 0 {
		req.Header.Set("Cookie", cookiesToHeader(cookies))
	}
}

func cookiesToHeader(m map[string]string) string {
	if len(m) == 0 {
		return ""
	}
	var b bytes.Buffer
	i := 0
	for k, v := range m {
		if i > 0 {
			b.WriteString("; ")
		}
		b.WriteString(k)
		b.WriteString("=")
		b.WriteString(v)
		i++
	}
	return b.String()
}

// PostForm posts application/x-www-form-urlencoded data and returns body bytes
func (c *Client) PostForm(u string, form map[string]string, headers map[string]string, cookies map[string]string) ([]byte, *http.Response, error) {
	values := url.Values{}
	for k, v := range form {
		values.Set(k, v)
	}
	req, err := http.NewRequest(http.MethodPost, u, bytes.NewBufferString(values.Encode()))
	if err != nil {
		return nil, nil, err
	}
	c.addHeaders(req, headers, cookies)
	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, resp, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	return b, resp, err
}

// Get performs a GET request
func (c *Client) Get(u string, headers map[string]string, cookies map[string]string) ([]byte, *http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}
	c.addHeaders(req, headers, cookies)
	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, resp, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	return b, resp, err
}

// JSON helper
func ParseJSON[T any](b []byte, v *T) error { return json.Unmarshal(b, v) }
