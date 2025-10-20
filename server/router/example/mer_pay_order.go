package example

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type MerPayOrderRouter struct{}

// InitMerPayOrderRouter 初始化 merPayOrder表 路由信息
func (s *MerPayOrderRouter) InitMerPayOrderRouter(Router *gin.RouterGroup, PublicRouter *gin.RouterGroup) {
	merPayOrderRouter := Router.Group("merPayOrder").Use(middleware.OperationRecord()).Use(middleware.WithSysUserID())
	merPayOrderRouterWithoutRecord := Router.Group("merPayOrder").Use(middleware.WithSysUserID())
	merPayOrderRouterWithoutAuth := PublicRouter.Group("merPayOrder").Use(middleware.PermanentTokenAuth()).Use(middleware.WithSysUserID())

	//merPayOrderRouterWithoutAuth := PublicRouter.Group("merPayOrder")
	{
		merPayOrderRouter.POST("createMerPayOrder", merPayOrderApi.CreateMerPayOrder)             // 新建merPayOrder表
		merPayOrderRouter.DELETE("deleteMerPayOrder", merPayOrderApi.DeleteMerPayOrder)           // 删除merPayOrder表
		merPayOrderRouter.DELETE("deleteMerPayOrderByIds", merPayOrderApi.DeleteMerPayOrderByIds) // 批量删除merPayOrder表
		merPayOrderRouter.PUT("updateMerPayOrder", merPayOrderApi.UpdateMerPayOrder)              // 更新merPayOrder表
	}
	{
		merPayOrderRouterWithoutRecord.GET("findMerPayOrder", merPayOrderApi.FindMerPayOrder)       // 根据ID获取merPayOrder表
		merPayOrderRouterWithoutRecord.GET("getMerPayOrderList", merPayOrderApi.GetMerPayOrderList) // 获取merPayOrder表列表
	}
	{
		merPayOrderRouterWithoutAuth.GET("getMerPayOrderState", merPayOrderApi.GetMerPayOrderState)  // 根据ID获取merPayOrder状态
		merPayOrderRouterWithoutAuth.GET("cancelMerPayOrderState", merPayOrderApi.CancelMerPayOrder) // 根据ID取消订单
	}
	{
		//merPayOrderRouterWithoutAuth.GET("getMerPayOrderPublic", merPayOrderApi.GetMerPayOrderPublic) // merPayOrder表开放接口
	}
}
