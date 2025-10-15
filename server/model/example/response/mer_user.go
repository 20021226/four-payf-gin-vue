package response

import (
	"time"
)

// MerUserListItem 用于列表接口的精简返回结构，隐藏敏感字段
type MerUserListItem struct {
	Id         *int32     `json:"id" gorm:"column:id"`
	MerType    *string    `json:"merType" gorm:"column:mer_type"`
	UserName   *string    `json:"userName" gorm:"column:user_name"`
	State      *bool      `json:"state" gorm:"column:state"`
	CreateTime *time.Time `json:"createTime" gorm:"column:create_time"`
	UpdateTime *time.Time `json:"updateTime" gorm:"column:update_time"`
	Remarks    *string    `json:"remarks" gorm:"column:remarks"`
	QrCode     *string    `json:"qrCode" gorm:"column:qr_code"`
	Password   *string    `json:"password" gorm:"column:password"`
}

type PaymentQrCodeResponse struct {
	QrcodeCode *string  `json:"qrcodeCode"`
	Amount     *float64 `json:"amount"`
	CreateTime *string  `json:"createTime"`
	OrderId    *string  `json:"orderId"`
}
