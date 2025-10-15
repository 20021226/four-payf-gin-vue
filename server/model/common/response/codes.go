package response

// 统一业务响应码定义与默认消息注册
// 建议按模块分段（例如：MerUser 模块使用 100xxx），系统级保留 0/1 等通用码

const (
    // MerUser 模块响应码（示例，可按需扩展/调整）
    CodeMerUserCreateOK     = 100000
    CodeMerUserUnauthorized = 100100
    CodeMerUserParamInvalid = 100200
    CodeMerUserNotFound     = 100404
    CodeMerUserInternalError= 100500
)

// DefaultCodeMessages 集中维护“响应码→默认文案”的映射
// 可在运行时通过 SetCodeMessage/SetCodeMessages 进行覆盖或扩展
var DefaultCodeMessages = map[int]string{
    // 系统级通用码
    SUCCESS: "成功",
    ERROR:   "失败",

    // MerUser 模块默认文案（示例）
    CodeMerUserCreateOK:      "创建成功",
    CodeMerUserUnauthorized:  "未授权调用",
    CodeMerUserParamInvalid:  "参数校验失败",
    CodeMerUserNotFound:      "未找到资源",
    CodeMerUserInternalError: "服务器内部错误",
}

// 在包初始化阶段注册默认文案（可被后续调用覆盖）
func init() {
    SetCodeMessages(DefaultCodeMessages)
}