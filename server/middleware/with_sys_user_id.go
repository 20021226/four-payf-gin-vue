package middleware

import (
    "context"
    "github.com/flipped-aurora/gin-vue-admin/server/utils"
    "github.com/gin-gonic/gin"
)

// WithSysUserID 从 JWT claims 中提取用户ID，注入到 gin.Context 和 Request.Context
// 在使用该中间件的路由中，后续的处理函数可以通过 utils.GetUserID(c) 获取，
// 同时也可以从 request.Context 中通过 key "sys_user_id" 获取。
func WithSysUserID() gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := utils.GetUserID(c)
        // 注入到 gin.Context（已有 GetUserID 基于 claims，可直接使用）
        // 额外注入到 Request.Context，方便 service 层通过 ctx 访问
        ctx := context.WithValue(c.Request.Context(), "sys_user_id", userID)
        c.Request = c.Request.WithContext(ctx)
        c.Next()
    }
}