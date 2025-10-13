package response

import "github.com/flipped-aurora/gin-vue-admin/server/model/example"

// SysUserConfigItem 仅用于对外返回的精简字段
type SysUserConfigItem struct {
    Name   *string `json:"name"`
    Title  *string `json:"title"`
    FormId *int32  `json:"formId"`
    Value  *string `json:"value"`
}

// NewSysUserConfigItem 从模型转换为对外返回结构
func NewSysUserConfigItem(m example.SysUserConfig) SysUserConfigItem {
    return SysUserConfigItem{
        Name:   m.Name,
        Title:  m.Title,
        FormId: m.FormId,
        Value:  m.Value,
    }
}

// NewSysUserConfigItemList 批量转换列表
func NewSysUserConfigItemList(ms []example.SysUserConfig) []SysUserConfigItem {
    items := make([]SysUserConfigItem, 0, len(ms))
    for _, m := range ms {
        items = append(items, NewSysUserConfigItem(m))
    }
    return items
}