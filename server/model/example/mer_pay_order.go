// 自动生成模板MerPayOrder
package example

import (
	"github.com/shopspring/decimal"
	"time"
)

// merPayOrder表 结构体  MerPayOrder
type MerPayOrder struct {
	Id             *int64           `json:"id" form:"id" gorm:"index;primarykey;column:id;"`                                       //id字段
	SysUserId      *int64           `json:"sysUserId" form:"sysUserId" gorm:"comment:管理ID;column:sys_user_id;"`                    //管理ID
	OrderId        *string          `json:"orderId" form:"orderId" gorm:"index;comment:订单id;column:order_id;size:255;"`            //订单id
	MerName        *string          `json:"merName" form:"merName" gorm:"comment:商户名称;column:mer_name;size:255;"`                  //商户名称
	MerId          *int64           `json:"merId" form:"merId" gorm:"comment:商户id;column:mer_id;"`                                 //商户id
	State          *int8            `json:"state" form:"state" gorm:"comment:支付状态;column:state;default:0;"`                        //支付状态
	IsDel          *string          `json:"isDel" form:"isDel" gorm:"comment:是否删除(1: 删除 0:未删除);column:is_del;size:255;default:0;"` //是否删除(1: 删除 0:未删除)
	RequestAmmount *decimal.Decimal `json:"requestAmmount" form:"requestAmmount" gorm:"type:decimal(10,2);comment:请求金额;column:request_ammount;"`
	Ammount        *decimal.Decimal `json:"ammount" form:"ammount" gorm:"type:decimal(10,2);comment:实际收款金额;column:ammount;"`    //实际收款金额
	PayTime        *time.Time       `json:"payTime" form:"payTime" gorm:"comment:支付时间;column:pay_time;"`                        //支付时间
	CreateTime     *time.Time       `json:"createTime" form:"createTime" gorm:"comment:创建时间;column:create_time;autoCreateTime"` //创建时间
	UpdateTime     *time.Time       `json:"updateTime" form:"updateTime" gorm:"comment:更新时间;column:update_time;autoUpdateTime"` //更新时间
	Remarks        *string          `json:"remarks" form:"remarks" gorm:"comment:订单备注;column:remarks;size:255;"`                //订单备注
	MerType        *string          `json:"merType" form:"merType" gorm:"comment:商户类型;column:mer_type;size:255;"`               //商户类型
}

// TableName merPayOrder表 MerPayOrder自定义表名 mer_pay_order
func (MerPayOrder) TableName() string {
	return "mer_pay_order"
}
