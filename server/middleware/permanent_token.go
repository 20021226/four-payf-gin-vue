package middleware

import (
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// PermanentTokenAuth 永久token验证中间件
func PermanentTokenAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取token
		token := c.GetHeader("Authorization")
		if token == "" {
			// 也可以从查询参数获取token
			token = c.Query("token")
		}

		if token == "" {
			global.GVA_LOG.Error("永久token验证失败：未提供token")
			response.FailWithMessage("未提供认证token", c)
			c.Abort()
			return
		}

		// 移除Bearer前缀（如果存在）
		if strings.HasPrefix(token, "Bearer ") {
			token = strings.TrimPrefix(token, "Bearer ")
		}

		// 验证永久token
		permanentToken, err := utils.ValidatePermanentToken(token)
		if err != nil {
			global.GVA_LOG.Error("永久token验证失败",
				zap.String("token", token),
				zap.Error(err))
			response.FailWithMessage("token验证失败: "+err.Error(), c)
			c.Abort()
			return
		}

		// 将用户ID存储到上下文中，供后续处理使用
		c.Set("sys_user_id", permanentToken.UserID)
		c.Set("token_type", "permanent")
		c.Set("permanent_token", permanentToken)

		global.GVA_LOG.Info("api token验证成功",
			zap.Uint("sys_user_id", permanentToken.UserID),
			zap.String("token", token))

		c.Next()
	}
}

// PermanentTokenOrJWT 支持永久token或JWT验证的中间件
// 优先验证永久token，如果失败则尝试JWT验证
func PermanentTokenOrJWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取token
		token := c.GetHeader("Authorization")
		if token == "" {
			// 也可以从查询参数获取token
			token = c.Query("token")
		}

		if token == "" {
			global.GVA_LOG.Error("认证失败：未提供token")
			response.FailWithMessage("未提供认证token", c)
			c.Abort()
			return
		}

		// 移除Bearer前缀（如果存在）
		originalToken := token
		if strings.HasPrefix(token, "Bearer ") {
			token = strings.TrimPrefix(token, "Bearer ")
		}

		// 首先尝试验证永久token
		if strings.HasPrefix(token, "perm_") {
			permanentToken, err := utils.ValidatePermanentToken(token)
			if err != nil {
				global.GVA_LOG.Error("永久token验证失败",
					zap.String("token", token),
					zap.Error(err))
				response.FailWithMessage("永久token验证失败: "+err.Error(), c)
				c.Abort()
				return
			}

			// 永久token验证成功
			c.Set("user_id", permanentToken.UserID)
			c.Set("token_type", "permanent")
			c.Set("permanent_token", permanentToken)

			global.GVA_LOG.Info("永久token验证成功",
				zap.Uint("user_id", permanentToken.UserID))

			c.Next()
			return
		}

		// 如果不是永久token，则尝试JWT验证
		// 恢复原始token格式进行JWT验证
		c.Request.Header.Set("Authorization", originalToken)

		// 调用原有的JWT中间件逻辑
		jwtAuth := JWTAuth()
		jwtAuth(c)
	}
}

// GetUserIDFromContext 从上下文中获取用户ID
func GetUserIDFromContext(c *gin.Context) (uint, bool) {
	if userID, exists := c.Get("sys_user_id"); exists {
		if uid, ok := userID.(uint); ok {
			return uid, true
		}
	}
	return 0, false
}

// GetTokenTypeFromContext 从上下文中获取token类型
func GetTokenTypeFromContext(c *gin.Context) (string, bool) {
	if tokenType, exists := c.Get("token_type"); exists {
		if tType, ok := tokenType.(string); ok {
			return tType, true
		}
	}
	return "", false
}

// IsPermanentToken 检查当前请求是否使用永久token
func IsPermanentToken(c *gin.Context) bool {
	tokenType, exists := GetTokenTypeFromContext(c)
	return exists && tokenType == "permanent"
}

// GetUserIDFromTokenOrJWT 统一获取用户ID，支持永久token和JWT
func GetUserIDFromTokenOrJWT(c *gin.Context) uint {
	// 首先尝试从永久token中获取用户ID
	if userID, exists := GetUserIDFromContext(c); exists {
		return userID
	}

	// 如果没有永久token，则从JWT中获取用户ID
	return utils.GetUserID(c)
}
