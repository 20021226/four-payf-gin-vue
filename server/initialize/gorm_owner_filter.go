package initialize

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"gorm.io/gorm"
)

// RegisterOwnerFilter 注册基于请求上下文中的 sys_user_id 的自动所有权过滤与写入
// - Query/Update/Delete：自动追加 Where("sys_user_id = ?")
// - Create：在模型包含 SysUserId 字段的情况下，自动写入 sys_user_id
func RegisterOwnerFilter() {
	db := global.GVA_DB
	if db == nil {
		return
	}

	// 在查询/更新/删除前追加过滤条件
	db.Callback().Query().Before("gorm:query").Register("owner:filter", func(tx *gorm.DB) {
		addOwnerFilter(tx)
	})
	db.Callback().Update().Before("gorm:update").Register("owner:filter", func(tx *gorm.DB) {
		addOwnerFilter(tx)
	})
	db.Callback().Delete().Before("gorm:delete").Register("owner:filter", func(tx *gorm.DB) {
		addOwnerFilter(tx)
	})

	// 在创建前自动写入归属字段
	db.Callback().Create().Before("gorm:create").Register("owner:set", func(tx *gorm.DB) {
		setOwnerOnCreate(tx)
	})
}

func addOwnerFilter(tx *gorm.DB) {
	if tx == nil || tx.Statement == nil || tx.Statement.Schema == nil {
		return
	}
	v := tx.Statement.Context.Value("sys_user_id")
	var uid int64
	switch vv := v.(type) {
	case uint:
		uid = int64(vv)
	case int64:
		uid = vv
	default:
		return
	}
	if uid == 0 {
		return
	}
	if _, ok := tx.Statement.Schema.FieldsByDBName["sys_user_id"]; !ok {
		return
	}
	tx.Where("sys_user_id = ?", uid)
}

func setOwnerOnCreate(tx *gorm.DB) {
	if tx == nil || tx.Statement == nil || tx.Statement.Schema == nil {
		return
	}
	v := tx.Statement.Context.Value("sys_user_id")
	var uid int64
	switch vv := v.(type) {
	case uint:
		uid = int64(vv)
	case int64:
		uid = vv
	default:
		return
	}
	if uid == 0 {
		return
	}
	// 仅当模型包含 SysUserId 字段时才写入
	if f := tx.Statement.Schema.LookUpField("SysUserId"); f != nil {
		tx.Statement.SetColumn("SysUserId", uid, true)
	}
}
