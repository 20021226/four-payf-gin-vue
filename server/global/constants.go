package global

// 集中放置全局共享常量与示例共享变量
// - 常量用于约定统一的键名、默认配置等
// - 变量可在运行时被配置/注入（尽量保持只读或通过配置文件管理）

const (
	MERE_XINGYI_CODE = 0
	MER_RICH_CODE    = 1
	MER_XIANG_XIAN   = 0

	MER_XIANG_XIAN_KEY = "xiangxian"
	MER_XINGYI_KEY     = "xingyi"
	MER_RICH_KEY       = "rich"

	PAY_AMOUNT_USED_KEY = "pay_ammount_used"

	PAY_ORDER_STATE_PENDING = "pending"

	MER_TYPE_XINGYI     = "0"
	MER_TYPE_RICH       = "1"
	MER_TYPE_XIANG_XIAN = "2"

	REDIS_PAY_REQUEST_TOKEN = "pay_request_token"
)

// Shared 演示性的全局共享变量容器（可按需扩展字段并由初始化流程赋值）
// 如需更复杂的全局共享配置，建议继续使用 GVA 的配置系统（global.GVA_CONFIG）
var Shared = struct {
	ServiceName string
}{
	ServiceName: "four-pay-gin-vue-admin-server",
}
