package request

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	"time"
)

type MerUserSearch struct {
	request.PageInfo
	MerName  *string `json:"merName" form:"merName"` //商户名称
	Id       *int64  `json:"id" form:"id"`           //id字段
	MerType  *string `json:"merType" form:"merType"`
	UserName *string `json:"userName" form:"userName"` //账号
}

type PaymentQrCodeResponse struct {
	CallBack   *string    `json:"qrcodeCode"`
	Amount     *float64   `json:"amount"`
	CreateTime *time.Time `json:"createTime"`
}
