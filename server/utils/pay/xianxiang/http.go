package xianxiang

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client HTTP 客户端结构体
type Client struct {
	hc *http.Client
}

// NewClient 创建新的 HTTP 客户端
func NewClient() *Client {
	return &Client{
		hc: &http.Client{Timeout: 20 * time.Second},
	}
}

// addHeaders 添加请求头
func (c *Client) addHeaders(req *http.Request, headers map[string]string) {
	for k, v := range headers {
		req.Header.Set(k, v)
	}
}

// PostForm 发送表单数据的 POST 请求
func (c *Client) PostForm(u string, form map[string]string, headers map[string]string, cookies map[string]string) ([]byte, *http.Response, error) {
	// 构建表单数据
	data := url.Values{}
	for k, v := range form {
		data.Set(k, v)
	}

	// 创建请求
	req, err := http.NewRequest("POST", u, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, nil, err
	}

	// 设置默认的 Content-Type
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 添加自定义请求头
	c.addHeaders(req, headers)

	// 添加 cookies（如果有的话）
	if cookies != nil {
		for k, v := range cookies {
			req.AddCookie(&http.Cookie{Name: k, Value: v})
		}
	}

	// 发送请求
	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp, err
	}

	return body, resp, nil
}

// Get 发送 GET 请求
func (c *Client) Get(u string, headers map[string]string, cookies map[string]string) ([]byte, *http.Response, error) {
	// 创建请求
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	// 添加自定义请求头
	c.addHeaders(req, headers)

	// 添加 cookies（如果有的话）
	if cookies != nil {
		for k, v := range cookies {
			req.AddCookie(&http.Cookie{Name: k, Value: v})
		}
	}

	// 发送请求
	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp, err
	}

	return body, resp, nil
}