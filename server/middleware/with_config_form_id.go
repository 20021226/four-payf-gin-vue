package middleware

import (
	"context"
	"strconv"

	"github.com/gin-gonic/gin"
)

// WithConfigFormID 从查询参数提取 formId 注入到 Request.Context
// 示例: /sysUserConfig/getSysUserConfigList?page=1&pageSize=100&formId=1
// 在 service 层可通过 ctx.Value("form_id") 读取并用于筛选
func WithConfigFormID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从查询参数中提取 formId
		formIdStr := c.Query("formId")
		if formIdStr != "" {
			if v, err := strconv.ParseInt(formIdStr, 10, 32); err == nil {
				// 将 form_id(int32) 注入到 Request.Context
				ctx := context.WithValue(c.Request.Context(), "form_id", int32(v))
				c.Request = c.Request.WithContext(ctx)
			}
		}
		c.Next()
	}
}
