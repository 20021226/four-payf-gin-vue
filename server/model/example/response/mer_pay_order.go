package response

import (
	"github.com/shopspring/decimal"
	"time"
)

// MerPayOrderStateResponse 用于GetMerPayOrderState接口的精简返回结构，只包含必要字段
type MerPayOrderStateResponse struct {
	Id             *int64           `json:"id"`                // 订单ID
	OrderId        *string          `json:"orderId"`           // 订单id
	State          *int8            `json:"state"`             // 支付状态
	RequestAmmount *decimal.Decimal `json:"requestAmmount"`    // 请求金额
	Ammount        *decimal.Decimal `json:"ammount"`           // 实际收款金额
	PayTime        *time.Time       `json:"payTime"`           // 支付时间
	CreateTime     *time.Time       `json:"createTime"`        // 创建时间
}