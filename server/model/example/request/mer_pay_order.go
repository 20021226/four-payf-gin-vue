package request

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	"time"
)

type MerPayOrderSearch struct {
	OrderId         *string     `json:"orderId" form:"orderId"`
	MerName         *string     `json:"merName" form:"merName"`
	State           *bool       `json:"state" form:"state"`
	RequestAmmount  *string     `json:"requestAmmount" form:"requestAmmount"`
	Ammount         *float64    `json:"ammount" form:"ammount"`
	PayTimeRange    []time.Time `json:"payTimeRange" form:"payTimeRange[]"`
	CreateTimeRange []time.Time `json:"createTimeRange" form:"createTimeRange[]"`
	UpdateTimeRange []time.Time `json:"updateTimeRange" form:"updateTimeRange[]"`
	request.PageInfo
}
