package example

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type MerUserRouter struct{}

// InitMerUserRouter 初始化 merUser表 路由信息
func (s *MerUserRouter) InitMerUserRouter(Router *gin.RouterGroup, PublicRouter *gin.RouterGroup) {
	// 在 merUser 私有路由组挂载注入 sys_user_id 中间件
	merUserRouter := Router.Group("merUser").Use(middleware.OperationRecord(), middleware.WithSysUserID())
	merUserRouterWithoutRecord := Router.Group("merUser").Use(middleware.WithSysUserID())
	// merUserRouterWithoutAuth := PublicRouter.Group("merUser")
	{
		merUserRouter.POST("createMerUser", merUserApi.CreateMerUser)             // 新建merUser表
		merUserRouter.DELETE("deleteMerUser", merUserApi.DeleteMerUser)           // 删除merUser表
		merUserRouter.DELETE("deleteMerUserByIds", merUserApi.DeleteMerUserByIds) // 批量删除merUser表
		merUserRouter.PUT("updateMerUser", merUserApi.UpdateMerUser)              // 更新merUser表
		merUserRouter.POST("getPayQrCode", merUserApi.GetPayQrCode)               // 获取收款码
	}
	{
		merUserRouterWithoutRecord.GET("findMerUser", merUserApi.FindMerUser)       // 根据ID获取merUser表
		merUserRouterWithoutRecord.GET("getMerUserList", merUserApi.GetMerUserList) // 获取merUser表列表
	}
	// {
	// 	merUserRouterWithoutAuth.GET("getMerUserPublic", merUserApi.GetMerUserPublic) // merUser表开放接口
	// }

}
