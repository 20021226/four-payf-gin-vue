# 永久Token使用说明

## 功能概述

为了支持第三方系统调用API，我们实现了永久token机制。该机制允许用户生成长期有效的token，第三方系统可以使用这些token来调用指定的API接口，而不需要使用传统的JWT登录流程。

## 主要特性

1. **永久有效**: token有效期设置为10年，适合长期使用
2. **独立验证**: 不影响原有的JWT验证机制
3. **单一token机制**: 每个用户只能拥有一个有效的永久token，生成新token时会自动撤销所有旧token
4. **安全管理**: 支持token的生成、查看和撤销
5. **Redis存储**: token信息存储在Redis中，支持高并发访问

## API接口

### 1. 生成永久token
- **接口**: `POST /api/merUser/generatePermanentToken`
- **认证**: 需要JWT认证
- **说明**: 每个用户只能拥有一个有效的永久token，生成新token时会自动撤销所有旧token
- **响应示例**:
```json
{
  "code": 0,
  "data": {
    "token": "perm_a1b2c3d4e5f6...",
    "user_id": 123,
    "created_at": 1640995200,
    "notice": "新token已生成，所有旧token已自动失效"
  },
  "msg": "生成成功，旧token已失效"
}
```

### 2. 获取永久token列表
- **接口**: `GET /api/merUser/getPermanentTokens`
- **认证**: 需要JWT认证
- **响应示例**:
```json
{
  "code": 0,
  "data": {
    "tokens": [
      {
        "user_id": 123,
        "token": "perm_a1b2c3d4e5f6...",
        "created_at": 1640995200,
        "is_active": true
      }
    ],
    "total": 1
  },
  "msg": "获取成功"
}
```

### 3. 撤销永久token
- **接口**: `POST /api/merUser/revokePermanentToken`
- **认证**: 需要JWT认证
- **请求参数**:
```json
{
  "token": "perm_a1b2c3d4e5f6..."
}
```

## 第三方API调用

### 支持永久token的接口

目前支持永久token验证的接口：
- `POST /api/merUser/getPayQrCode` - 获取收款码

### 调用方式

第三方系统可以通过以下两种方式传递永久token：

#### 方式1: Authorization Header
```bash
curl -X POST "http://your-domain/api/merUser/getPayQrCode" \
  -H "Authorization: Bearer perm_a1b2c3d4e5f6..." \
  -H "Content-Type: application/json" \
  -d '{
    "payAmmount": 100,
    "orderId": "ORDER123",
    "merType": "0"
  }'
```

#### 方式2: Query Parameter
```bash
curl -X POST "http://your-domain/api/merUser/getPayQrCode?token=perm_a1b2c3d4e5f6..." \
  -H "Content-Type: application/json" \
  -d '{
    "payAmmount": 100,
    "orderId": "ORDER123",
    "merType": "0"
  }'
```

## 技术实现

### 中间件支持

1. **PermanentTokenAuth()**: 仅支持永久token验证
2. **PermanentTokenOrJWT()**: 支持永久token或JWT验证（优先永久token）

### 路由配置示例

```go
// 第三方API路由组，使用永久token验证
merUserThirdPartyRouter := PublicRouter.Group("merUser").Use(middleware.PermanentTokenAuth())
{
    merUserThirdPartyRouter.POST("getPayQrCode", merUserApi.GetPayQrCode)
}
```

### Token格式

- 永久token以 `perm_` 前缀开头
- 使用MD5哈希生成，包含用户ID、时间戳和随机字符串
- 示例: `perm_a1b2c3d4e5f6789012345678901234567890`

## 安全注意事项

1. **Token保护**: 永久token具有长期有效性，请妥善保管
2. **单一token机制**: 系统确保每个用户只能拥有一个有效的永久token，生成新token时会自动撤销所有旧token，避免token泄露风险
3. **定期轮换**: 建议定期生成新token，系统会自动撤销旧token
4. **权限控制**: token与用户绑定，继承用户的权限设置
5. **监控日志**: 系统会记录token的生成、使用和撤销日志

## 扩展其他接口

如需为其他接口添加永久token支持，请按以下步骤操作：

1. 在路由中使用 `middleware.PermanentTokenAuth()` 中间件
2. 在API处理函数中使用 `middleware.GetUserIDFromTokenOrJWT(c)` 获取用户ID
3. 确保接口逻辑支持从上下文获取用户信息

示例：
```go
// 路由配置
apiRouter := PublicRouter.Group("api").Use(middleware.PermanentTokenAuth())
{
    apiRouter.POST("yourEndpoint", yourApi.YourMethod)
}

// API处理函数
func (api *YourApi) YourMethod(c *gin.Context) {
    userID := middleware.GetUserIDFromTokenOrJWT(c)
    // 业务逻辑...
}
```