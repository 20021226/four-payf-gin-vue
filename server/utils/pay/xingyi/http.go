package xingyi

import (
    "bytes"
    "encoding/json"
    "io"
    "net/http"
    "net/url"
    "time"
)

// Default headers equivalent to Python headers
var DefaultHeaders = map[string]string{
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

// Client wraps http.Client with cookie jar-like simple map
type Client struct {
    hc      *http.Client
    Cookies map[string]string
    Headers map[string]string
}

func NewClient() *Client {
    return &Client{
        hc: &http.Client{Timeout: 20 * time.Second},
        Cookies: map[string]string{
            "$cookiesession1": "678A3E1E7E7825003E9E300211ADC19F",
            "acw_tc":          "0a18d73717600661636627422e2cd59be1384e43ad8fa00a5b4daaa41e9cba",
            "JSESSIONID":      "",
        },
        Headers: DefaultHeaders,
    }
}

func (c *Client) SetHeader(k, v string) { c.Headers[k] = v }
func (c *Client) SetCookie(k, v string) { c.Cookies[k] = v }

func (c *Client) addHeaders(req *http.Request) {
    for k, v := range c.Headers { req.Header.Set(k, v) }
    // simple cookie header
    if len(cookiesToHeader(c.Cookies)) > 0 { req.Header.Set("Cookie", cookiesToHeader(c.Cookies)) }
}

func cookiesToHeader(m map[string]string) string {
    if len(m) == 0 { return "" }
    var b bytes.Buffer
    i := 0
    for k, v := range m {
        if i > 0 { b.WriteString("; ") }
        b.WriteString(k)
        b.WriteString("=")
        b.WriteString(v)
        i++
    }
    return b.String()
}

// PostForm posts application/x-www-form-urlencoded data and returns body bytes
func (c *Client) PostForm(u string, form map[string]string) ([]byte, *http.Response, error) {
    values := url.Values{}
    for k, v := range form { values.Set(k, v) }
    req, err := http.NewRequest(http.MethodPost, u, bytes.NewBufferString(values.Encode()))
    if err != nil { return nil, nil, err }
    c.addHeaders(req)
    resp, err := c.hc.Do(req)
    if err != nil { return nil, resp, err }
    defer resp.Body.Close()
    b, err := io.ReadAll(resp.Body)
    return b, resp, err
}

// Get performs a GET request
func (c *Client) Get(u string) ([]byte, *http.Response, error) {
    req, err := http.NewRequest(http.MethodGet, u, nil)
    if err != nil { return nil, nil, err }
    c.addHeaders(req)
    resp, err := c.hc.Do(req)
    if err != nil { return nil, resp, err }
    defer resp.Body.Close()
    b, err := io.ReadAll(resp.Body)
    return b, resp, err
}

// JSON helper
func ParseJSON[T any](b []byte, v *T) error { return json.Unmarshal(b, v) }