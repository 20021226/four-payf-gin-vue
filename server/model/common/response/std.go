package response

import (
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

// 可扩展的 code->message 映射
var (
	codeMessageMap = map[int]string{
		SUCCESS: "成功",
		ERROR:   "失败",
	}
	codeMessageMu sync.RWMutex
)

// SetCodeMessage 设置单个 code 的默认文案
func SetCodeMessage(code int, msg string) {
	codeMessageMu.Lock()
	codeMessageMap[code] = msg
	codeMessageMu.Unlock()
}

// SetCodeMessages 批量设置/覆盖默认文案
func SetCodeMessages(m map[int]string) {
	if m == nil {
		return
	}
	codeMessageMu.Lock()
	for k, v := range m {
		codeMessageMap[k] = v
	}
	codeMessageMu.Unlock()
}

// Std 标准返回：{code, message, success, data}
// - code: 状态码，默认 0 为成功，7 为失败
// - message: 若未传入，则根据 code 自动设置默认文案；可自定义覆盖
// - success: code == SUCCESS 为 true，否则为 false
// - data: 返回数据；nil 将被替换为空对象
func Std(c *gin.Context, code int, data interface{}, messageOverride ...string) {
	msg := defaultMessage(code)
	if len(messageOverride) > 0 {
		if m := strings.TrimSpace(messageOverride[0]); m != "" {
			msg = m
		}
	}
	success := code == SUCCESS
	if data == nil {
		data = map[string]interface{}{}
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    code,
		"message": msg,
		"success": success,
		"data":    data,
	})
}

// StdOk 成功返回，支持自定义 message
func StdOk(c *gin.Context, data interface{}, messageOverride ...string) {
	Std(c, SUCCESS, data, messageOverride...)
}

// StdFail 失败返回，message 必填
func StdFail(c *gin.Context, message string) {
	Std(c, ERROR, map[string]interface{}{}, message)
}

func StdGeneral(c *gin.Context, message string) {
	Std(c, ERROR, map[string]interface{}{}, message)
}

// defaultMessage 根据 code 返回默认文案
func defaultMessage(code int) string {
	codeMessageMu.RLock()
	msg, ok := codeMessageMap[code]
	codeMessageMu.RUnlock()
	if ok {
		return msg
	}
	return "未知状态"
}
