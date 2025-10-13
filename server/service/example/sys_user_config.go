package example

import (
	"context"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/example"
	exampleReq "github.com/flipped-aurora/gin-vue-admin/server/model/example/request"
)

type SysUserConfigService struct{}

// CreateSysUserConfig 创建sysUserConfig表记录
// Author [yourname](https://github.com/yourname)
func (sysUserConfigService *SysUserConfigService) CreateSysUserConfig(ctx context.Context, sysUserConfig *example.SysUserConfig) (err error) {
	db := global.GVA_DB.WithContext(ctx)
	err = db.Create(sysUserConfig).Error
	return err
}

// DeleteSysUserConfig 删除sysUserConfig表记录
// Author [yourname](https://github.com/yourname)
func (sysUserConfigService *SysUserConfigService) DeleteSysUserConfig(ctx context.Context, id string) (err error) {
	db := global.GVA_DB.WithContext(ctx).Model(&example.SysUserConfig{})
	err = db.Where("id = ?", id).Delete(&example.SysUserConfig{}).Error
	return err
}

// DeleteSysUserConfigByIds 批量删除sysUserConfig表记录
// Author [yourname](https://github.com/yourname)
func (sysUserConfigService *SysUserConfigService) DeleteSysUserConfigByIds(ctx context.Context, ids []string) (err error) {
	db := global.GVA_DB.WithContext(ctx).Model(&example.SysUserConfig{})
	err = db.Where("id in ?", ids).Delete(&[]example.SysUserConfig{}).Error
	return err
}

// UpdateSysUserConfig 更新sysUserConfig表记录
// Author [yourname](https://github.com/yourname)
func (sysUserConfigService *SysUserConfigService) UpdateSysUserConfig(ctx context.Context, sysUserConfig example.SysUserConfig) (err error) {
	db := global.GVA_DB.WithContext(ctx).Model(&example.SysUserConfig{})
	err = db.Where("form_id = ?", sysUserConfig.FormId).
		Where("name = ?", sysUserConfig.Name).
		Omit("sys_user_id").
		Updates(&sysUserConfig).Error
	return err
}

// UpdateValueByNameForm 更新指定 name+formId 的 value 字段
func (sysUserConfigService *SysUserConfigService) UpdateValueByNameForm(ctx context.Context, name string, formId int32, value string) error {
	db := global.GVA_DB.WithContext(ctx).Model(&example.SysUserConfig{})
	return db.Where("name = ? AND form_id = ?", name, formId).Update("value", value).Error
}

// GetSysUserConfig 根据id获取sysUserConfig表记录
// Author [yourname](https://github.com/yourname)
func (sysUserConfigService *SysUserConfigService) GetSysUserConfig(ctx context.Context, id string) (sysUserConfig example.SysUserConfig, err error) {
	db := global.GVA_DB.WithContext(ctx)
	err = db.Where("id = ?", id).First(&sysUserConfig).Error
	return
}

// GetSysUserConfigInfoList 分页获取sysUserConfig表记录
// Author [yourname](https://github.com/yourname)
func (sysUserConfigService *SysUserConfigService) GetSysUserConfigInfoList(ctx context.Context, info exampleReq.SysUserConfigSearch) (list []example.SysUserConfig, total int64, err error) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)
	db := global.GVA_DB.WithContext(ctx).Model(&example.SysUserConfig{})
	var sysUserConfigs []example.SysUserConfig
	// 如果有条件搜索 下方会自动创建搜索语句
	// 优先使用查询参数中的 formId，其次从上下文读取
	var formId int32
	if info.FormId > 0 {
		formId = info.FormId
	} else if v := ctx.Value("form_id"); v != nil {
		if fid, ok := v.(int32); ok && fid > 0 {
			formId = fid
		}
	}
	if formId > 0 {
		db = db.Where("form_id = ?", formId)
	}

	err = db.Count(&total).Error
	if err != nil {
		return
	}

	if limit != 0 {
		db = db.Limit(limit).Offset(offset)
	}

	err = db.Find(&sysUserConfigs).Error
	return sysUserConfigs, total, err
}
func (sysUserConfigService *SysUserConfigService) GetSysUserConfigPublic(ctx context.Context) {
	// 此方法为获取数据源定义的数据
	// 请自行实现
}
