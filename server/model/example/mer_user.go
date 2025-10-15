// 自动生成模板MerUser
package example

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"strings"
	"time"
)

// merUser表 结构体  MerUser
type MerUser struct {
	Id        *int32  `json:"id" form:"id" gorm:"index;primarykey;autoIncrement;column:id;"`                    //id字段
	SysUserId *int64  `json:"sysUserId" form:"sysUserId" gorm:"index;comment:管理ID;column:sys_user_id;"`         //管理ID
	MerType   *string `json:"merType" form:"merType" gorm:"comment:接入类型(0:星驿,1:富掌柜);column:mer_type;size:255;"` //接入类型(0:星驿,1:富掌柜)
	UserName  *string `json:"userName" form:"userName" gorm:"comment:账号;column:user_name;size:255;"`            //账号
	Password  *string `json:"password" form:"password" gorm:"comment:密码;column:password;size:255;"`             //密码
	State     *bool   `json:"state" form:"state" gorm:"comment:是否启用(1: 启用 0:不启用);column:state;size:255;"`       //是否启用
	//QrCode     *string    `json:"qrCode" form:"qrCode" gorm:"comment:收款码;column:qr_code;"`                             //收款码
	QrCode    *string `json:"qrCode" form:"qrCode" gorm:"type:MEDIUMTEXT;comment:收款码;column:qr_code"`      // 收款码
	Key       *string `json:"key" form:"key" gorm:"comment:请求密钥;column:key;size:255;"`                     //请求密钥
	IsDel     *string `json:"isDel" form:"isDel" gorm:"comment:是否删除(1: 删除 0:未删除);column:is_del;size:255;"` //是否删除(1: 删除 0:未删除)
	MaxAmount *int64  `json:"maxAmount" form:"maxAmount" gorm:"index;comment:最大金额;column:max_amount;"`     //最大金额
	MinAmount *int64  `json:"minAmount" form:"minAmount" gorm:"index;comment:最小金额;column:min_amount;"`     //最小金额

	CreateTime *time.Time `json:"createTime" form:"createTime" gorm:"comment:创建时间;column:create_time;autoCreateTime;"` //创建时间
	UpdateTime *time.Time `json:"updateTime" form:"updateTime" gorm:"comment:更新时间;column:update_time;autoUpdateTime;"` //更新时间
	Remarks    *string    `json:"remarks" form:"remarks" gorm:"comment:备注;column:remarks;size:255;"`                   //备注
}

// TableName merUser表 MerUser自定义表名 mer_user
func (MerUser) TableName() string {
	return "mer_user"
}

const (
	MerTypeXingYi     = "0"
	MerTypeFuZhangGui = "1"
)

// AllowedMerTypes 持有可用的枚举值，便于后续扩展
var AllowedMerTypes = map[string]string{
	MerTypeXingYi:     "星驿",
	MerTypeFuZhangGui: "富掌柜",
}

// RegisterMerType 用于新增允许的枚举值
func RegisterMerType(code, name string) {
	if code == "" {
		return
	}
	AllowedMerTypes[code] = name
}

func (m *MerUser) validateMerType() error {
	if m.MerType == nil {
		return nil
	}
	if _, ok := AllowedMerTypes[*m.MerType]; ok {
		return nil
	}
	var opts []string
	for k, v := range AllowedMerTypes {
		opts = append(opts, fmt.Sprintf("%s(%s)", k, v))
	}
	return errors.New("invalid mer_type, allowed: " + strings.Join(opts, ", "))
}

func (m *MerUser) BeforeCreate(tx *gorm.DB) (err error) {
	return m.validateMerType()
}

func (m *MerUser) BeforeUpdate(tx *gorm.DB) (err error) {
	return m.validateMerType()
}
