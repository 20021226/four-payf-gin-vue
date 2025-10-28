package example

import (
	"context"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/example"
	exampleReq "github.com/flipped-aurora/gin-vue-admin/server/model/example/request"
)

type MerPayOrderService struct{}

// CreateMerPayOrder 创建merPayOrder表记录
// Author [yourname](https://github.com/yourname)
func (merPayOrderService *MerPayOrderService) CreateMerPayOrder(ctx context.Context, merPayOrder *example.MerPayOrder) (err error) {
	err = global.GVA_DB.Create(merPayOrder).Error
	return err
}

// DeleteMerPayOrder 删除merPayOrder表记录
// Author [yourname](https://github.com/yourname)
func (merPayOrderService *MerPayOrderService) DeleteMerPayOrder(ctx context.Context, id string) (err error) {
	err = global.GVA_DB.Delete(&example.MerPayOrder{}, "id = ?", id).Error
	return err
}

// DeleteMerPayOrderByIds 批量删除merPayOrder表记录
// Author [yourname](https://github.com/yourname)
func (merPayOrderService *MerPayOrderService) DeleteMerPayOrderByIds(ctx context.Context, ids []string) (err error) {
	err = global.GVA_DB.Delete(&[]example.MerPayOrder{}, "id in ?", ids).Error
	return err
}

// UpdateMerPayOrder 更新merPayOrder表记录
// Author [yourname](https://github.com/yourname)
func (merPayOrderService *MerPayOrderService) UpdateMerPayOrder(ctx context.Context, merPayOrder example.MerPayOrder) (err error) {
	err = global.GVA_DB.Model(&example.MerPayOrder{}).Where("id = ?", merPayOrder.Id).Updates(&merPayOrder).Error
	return err
}

// GetMerPayOrder 根据id获取merPayOrder表记录
// Author [yourname](https://github.com/yourname)
func (merPayOrderService *MerPayOrderService) GetMerPayOrder(ctx context.Context, id int64) (merPayOrder example.MerPayOrder, err error) {
	err = global.GVA_DB.Where("id = ?", id).First(&merPayOrder).Error
	return
}

// GetMerPayOrderByInfo 获取merPayOrder表记录
// Author [yourname](https://github.com/yourname)
func (merPayOrderService *MerPayOrderService) GetMerPayOrderByInfo(ctx context.Context, info example.MerPayOrder) (merPayOrder example.MerPayOrder, err error) {
	// 创建db
	db := global.GVA_DB.Model(&example.MerPayOrder{})
	// 如果有条件搜索 下方会自动创建搜索语句
	if info.OrderId != nil && *info.OrderId != "" {
		db = db.Where("order_id LIKE ?", "%"+*info.OrderId+"%")
	}
	if info.SysUserId != nil && *info.SysUserId != 0 {
		db = db.Where("sys_user_id = ?", *info.SysUserId)
	}
	if info.MerName != nil && *info.MerName != "" {
		db = db.Where("mer_name LIKE ?", "%"+*info.MerName+"%")
	}
	if info.State != nil {
		db = db.Where("state = ?", *info.State)
	}
	err = db.First(&merPayOrder).Error
	return merPayOrder, err
}

// GetMerPayOrderInfoList 分页获取merPayOrder表记录
// Author [yourname](https://github.com/yourname)
func (merPayOrderService *MerPayOrderService) GetMerPayOrderInfoList(ctx context.Context, info exampleReq.MerPayOrderSearch) (list []example.MerPayOrder, total int64, err error) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)
	// 创建db
	db := global.GVA_DB.Model(&example.MerPayOrder{})
	var merPayOrders []example.MerPayOrder
	// 如果有条件搜索 下方会自动创建搜索语句

	if info.OrderId != nil && *info.OrderId != "" {
		db = db.Where("order_id LIKE ?", "%"+*info.OrderId+"%")
	}
	if info.MerName != nil && *info.MerName != "" {
		db = db.Where("mer_name LIKE ?", "%"+*info.MerName+"%")
	}
	if info.State != nil {
		db = db.Where("state = ?", *info.State)
	}
	if info.RequestAmmount != nil && *info.RequestAmmount != "" {
		db = db.Where("request_ammount = ?", *info.RequestAmmount)
	}
	if info.Ammount != nil {
		db = db.Where("ammount = ?", *info.Ammount)
	}
	if len(info.PayTimeRange) == 2 {
		db = db.Where("pay_time BETWEEN ? AND ? ", info.PayTimeRange[0], info.PayTimeRange[1])
	}
	if len(info.CreateTimeRange) == 2 {
		db = db.Where("create_time BETWEEN ? AND ? ", info.CreateTimeRange[0], info.CreateTimeRange[1])
	}
	if len(info.UpdateTimeRange) == 2 {
		db = db.Where("update_time BETWEEN ? AND ? ", info.UpdateTimeRange[0], info.UpdateTimeRange[1])
	}
	err = db.Count(&total).Error
	if err != nil {
		return
	}

	if limit != 0 {
		db = db.Limit(limit).Offset(offset)
	}

	err = db.Find(&merPayOrders).Error
	return merPayOrders, total, err
}
func (merPayOrderService *MerPayOrderService) GetMerPayOrderPublic(ctx context.Context) {
	// 此方法为获取数据源定义的数据
	// 请自行实现
}
