package request

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	"time"
)

type MerUserSearch struct {
	request.PageInfo
}

type PaymentQrCodeResponse struct {
	CallBack   *string    `json:"qrcodeCode"`
	Amount     *float64   `json:"amount"`
	CreateTime *time.Time `json:"createTime"`
}
