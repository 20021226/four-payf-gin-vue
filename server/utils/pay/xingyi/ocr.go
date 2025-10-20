package xingyi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// OCRProvider defines an interface to recognize captcha from base64 image
type OCRProvider interface {
	// Recognize takes base64 image string and returns recognized text
	Recognize(base64Image string) (string, error)
}

type ocrResponse struct {
	Data string `json:"data"`
}

type Config struct {
	BaseURL     string
	Probability bool
	PNGFix      bool
	Timeout     time.Duration
}

type XingyiOCR struct {
	config Config
	client *http.Client
}

func NewXingyiOCR(config Config) OCRProvider {
	if config.Timeout == 0 {
		config.Timeout = 10 * time.Second
	}

	config.BaseURL = "http://192.168.0.208:9580/ocr"
	if config.BaseURL == "" {
		config.BaseURL = "http://ocr/ocr"
	}

	return &XingyiOCR{
		config: config,
		client: &http.Client{Timeout: config.Timeout},
	}
}

func (x *XingyiOCR) Recognize(base64Image string) (string, error) {
	form := url.Values{}
	form.Set("image", base64Image)
	form.Set("probability", "False")
	form.Set("png_fix", "False")
	fmt.Printf("请求URL: %s, 请求参数: %v\n", x.config.BaseURL, form)
	resp, err := x.client.PostForm(x.config.BaseURL, form)
	if err != nil {
		return "", fmt.Errorf("OCR请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OCR服务返回错误状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取OCR响应失败: %w", err)
	}

	var out ocrResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return "", fmt.Errorf("解析OCR响应失败: %w, 响应体: %s", err, string(body))
	}

	return out.Data, nil
}
