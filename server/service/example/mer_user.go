package example

import (
	"context"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/example"
	exampleReq "github.com/flipped-aurora/gin-vue-admin/server/model/example/request"
	exampleRes "github.com/flipped-aurora/gin-vue-admin/server/model/example/response"
	payReq "github.com/flipped-aurora/gin-vue-admin/server/model/pay/request"
)

type MerUserService struct{}

// CreateMerUser 创建merUser表记录
// Author [yourname](https://github.com/yourname)
func (merUserService *MerUserService) CreateMerUser(ctx context.Context, merUser *example.MerUser) (err error) {
	err = global.GVA_DB.WithContext(ctx).Create(merUser).Error
	return err
}

// DeleteMerUser 删除merUser表记录
// Author [yourname](https://github.com/yourname)
func (merUserService *MerUserService) DeleteMerUser(ctx context.Context, id string) (err error) {
	db := global.GVA_DB.WithContext(ctx).Model(&example.MerUser{})
	err = db.Where("id = ?", id).Delete(&example.MerUser{}).Error
	return err
}

// DeleteMerUserByIds 批量删除merUser表记录
// Author [yourname](https://github.com/yourname)
func (merUserService *MerUserService) DeleteMerUserByIds(ctx context.Context, ids []string) (err error) {
	db := global.GVA_DB.WithContext(ctx).Model(&example.MerUser{})
	err = db.Where("id in ?", ids).Delete(&example.MerUser{}).Error
	return err
}

// UpdateMerUser 更新merUser表记录
// Author [yourname](https://github.com/yourname)
func (merUserService *MerUserService) UpdateMerUser(ctx context.Context, merUser example.MerUser) (err error) {
	db := global.GVA_DB.WithContext(ctx).Model(&example.MerUser{})
	err = db.Where("id = ?", merUser.Id).Omit("sys_user_id").Updates(&merUser).Error
	return err
}

// GetMerUser 根据id获取merUser表记录
// Author [yourname](https://github.com/yourname)
func (merUserService *MerUserService) GetMerUser(ctx context.Context, id string) (merUser example.MerUser, err error) {
	db := global.GVA_DB.WithContext(ctx).Model(&example.MerUser{})
	err = db.Where("id = ?", id).First(&merUser).Error
	return
}

// GetMerUser 根据id获取merUser表记录
// Author [yourname](https://github.com/yourname)
func (merUserService *MerUserService) GetNomalMerUser(ctx context.Context, params payReq.PayQrcodeParms, userID uint) (list []example.MerUser, err error) {
	db := global.GVA_DB.WithContext(ctx).Model(&example.MerUser{})
	if params.MerId != nil {
		db = db.Where("id = ?", *params.MerId)
	}
	err = db.Where("state = ? AND is_del = ? AND sys_user_id = ?", 1, 0, userID).Find(&list).Error
	return
}

// GetMerUserInfoList 分页获取merUser表记录
// Author [yourname](https://github.com/yourname)
func (merUserService *MerUserService) GetMerUserInfoList(ctx context.Context, info exampleReq.MerUserSearch) (list []exampleRes.MerUserListItem, total int64, err error) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)
	// 创建db
	db := global.GVA_DB.WithContext(ctx).Model(&example.MerUser{})
	var merUsers []exampleRes.MerUserListItem
	// 如果有条件搜索 下方会自动创建搜索语句
	// 所有权过滤通过全局回调自动生效

	// 添加查询条件
	if info.MerName != nil && *info.MerName != "" {
		db = db.Where("mer_name LIKE ?", "%"+*info.MerName+"%")
	}
	if info.Id != nil {
		db = db.Where("id = ?", *info.Id)
	}
	if info.MerType != nil && *info.MerType != "" {
		db = db.Where("mer_type = ?", *info.MerType)
	}
	if info.UserName != nil && *info.UserName != "" {
		db = db.Where("user_name LIKE ?", "%"+*info.UserName+"%")
	}

	err = db.Count(&total).Error
	if err != nil {
		return
	}

	if limit != 0 {
		db = db.Limit(limit).Offset(offset)
	}

	err = db.Find(&merUsers).Error
	return merUsers, total, err
}
func (merUserService *MerUserService) GetMerUserPublic(ctx context.Context) {
	// 此方法为获取数据源定义的数据
	// 请自行实现
}
