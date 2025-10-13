package example

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type SysUserConfigRouter struct{}

// InitSysUserConfigRouter 初始化 sysUserConfig表 路由信息
func (s *SysUserConfigRouter) InitSysUserConfigRouter(Router *gin.RouterGroup, PublicRouter *gin.RouterGroup) {
    sysUserConfigRouter := Router.Group("sysUserConfig").Use(middleware.OperationRecord(), middleware.WithSysUserID(), middleware.WithConfigFormID())
    // sysUserConfigRouter := Router.Group("sysUserConfig").Use(middleware.OperationRecord(), middleware.WithSysUserID())
    sysUserConfigRouterWithoutRecord := Router.Group("sysUserConfig").Use(middleware.WithSysUserID(), middleware.WithConfigFormID())
    sysUserConfigRouterWithoutAuth := PublicRouter.Group("sysUserConfig")
	{
		sysUserConfigRouter.POST("createSysUserConfig", sysUserConfigApi.CreateSysUserConfig)             // 新建sysUserConfig表
		sysUserConfigRouter.DELETE("deleteSysUserConfig", sysUserConfigApi.DeleteSysUserConfig)           // 删除sysUserConfig表
		sysUserConfigRouter.DELETE("deleteSysUserConfigByIds", sysUserConfigApi.DeleteSysUserConfigByIds) // 批量删除sysUserConfig表
		sysUserConfigRouter.PUT("updateSysUserConfig", sysUserConfigApi.UpdateSysUserConfig)              // 更新sysUserConfig表
		sysUserConfigRouter.PUT("updateValueByNameForm", sysUserConfigApi.UpdateSysUserConfigValue)       // 仅更新指定 name+formId 的 value
	}
	{
		sysUserConfigRouterWithoutRecord.GET("findSysUserConfig", sysUserConfigApi.FindSysUserConfig)       // 根据ID获取sysUserConfig表
		sysUserConfigRouterWithoutRecord.GET("getSysUserConfigList", sysUserConfigApi.GetSysUserConfigList) // 获取sysUserConfig表列表
	}
	{
		sysUserConfigRouterWithoutAuth.GET("getSysUserConfigPublic", sysUserConfigApi.GetSysUserConfigPublic) // sysUserConfig表开放接口
	}
}
