
package request

import (
    "github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
    
)

type SysUserConfigSearch struct{
    request.PageInfo
    FormId int32 `json:"formId" form:"formId"`
}

// SysUserConfigUpdateValueReq 用于只更新 value 字段的请求体
type SysUserConfigUpdateValueReq struct {
    Name   string `json:"name" binding:"required"`
    FormId int32  `json:"formId" binding:"required"`
    Value  string `json:"value" binding:"required"`
}
