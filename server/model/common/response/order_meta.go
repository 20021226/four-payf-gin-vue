package response

import (
	"encoding/json"
)

// OrderMeta 订单元数据结构体
type OrderMeta struct {
	OrderId      *string `json:"orderId" redis:"orderId"`           // 订单ID
	MerId        *int64  `json:"merId" redis:"merId"`               // 商户ID
	QrCode       *string `json:"qrCode" redis:"qrCode"`             // 二维码
	CreateTime   string  `json:"createTime" redis:"createTime"`     // 创建时间
	UserID       uint    `json:"userID" redis:"userID"`             // 用户ID
	State        string  `json:"state" redis:"state"`               // 订单状态
	TotalAmmount string  `json:"totalAmmount" redis:"totalAmmount"` // 订单金额
	MerType      string  `json:"merType" redis:"merType"`           // 商户类型
}

// ToJSON 将结构体序列化为JSON字节数组
func (om *OrderMeta) ToJSON() ([]byte, error) {
	return json.Marshal(om)
}

// FromJSON 从JSON字节数组反序列化为结构体
func (om *OrderMeta) FromJSON(data []byte) error {
	return json.Unmarshal(data, om)
}

// NewOrderMeta 创建新的订单元数据实例
func NewOrderMeta(
	orderId *string,
	merId *int64,
	qrCode *string,
	createTime string,
	userID uint,
	state string,
	totalAmmount string,
	merType string,
) *OrderMeta {
	return &OrderMeta{
		OrderId:      orderId,
		MerId:        merId,
		QrCode:       qrCode,
		CreateTime:   createTime,
		UserID:       userID,
		State:        state,
		TotalAmmount: totalAmmount,
		MerType:      merType,
	}
}

// GetCacheKey 获取Redis缓存键
func (om *OrderMeta) GetCacheKey() string {
	if om.OrderId != nil {
		return "ORDER_" + *om.OrderId
	}
	return ""
}

// IsValid 验证订单元数据是否有效
func (om *OrderMeta) IsValid() bool {
	return om.OrderId != nil && *om.OrderId != ""
}
