package example

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/model/example"
	exampleReq "github.com/flipped-aurora/gin-vue-admin/server/model/example/request"
	exampleResp "github.com/flipped-aurora/gin-vue-admin/server/model/example/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type SysUserConfigApi struct{}

// CreateSysUserConfig 创建sysUserConfig表
// @Tags SysUserConfig
// @Summary 创建sysUserConfig表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body example.SysUserConfig true "创建sysUserConfig表"
// @Success 200 {object} response.Response{msg=string} "创建成功"
// @Router /sysUserConfig/createSysUserConfig [post]
func (sysUserConfigApi *SysUserConfigApi) CreateSysUserConfig(c *gin.Context) {
	// 创建业务用Context
	ctx := c.Request.Context()

	var sysUserConfig example.SysUserConfig
	err := c.ShouldBindJSON(&sysUserConfig)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	err = sysUserConfigService.CreateSysUserConfig(ctx, &sysUserConfig)
	if err != nil {
		global.GVA_LOG.Error("创建失败!", zap.Error(err))
		response.FailWithMessage("创建失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

// DeleteSysUserConfig 删除sysUserConfig表
// @Tags SysUserConfig
// @Summary 删除sysUserConfig表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body example.SysUserConfig true "删除sysUserConfig表"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /sysUserConfig/deleteSysUserConfig [delete]
func (sysUserConfigApi *SysUserConfigApi) DeleteSysUserConfig(c *gin.Context) {
	// 创建业务用Context
	ctx := c.Request.Context()

	id := c.Query("id")
	err := sysUserConfigService.DeleteSysUserConfig(ctx, id)
	if err != nil {
		global.GVA_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// DeleteSysUserConfigByIds 批量删除sysUserConfig表
// @Tags SysUserConfig
// @Summary 批量删除sysUserConfig表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{msg=string} "批量删除成功"
// @Router /sysUserConfig/deleteSysUserConfigByIds [delete]
func (sysUserConfigApi *SysUserConfigApi) DeleteSysUserConfigByIds(c *gin.Context) {
	// 创建业务用Context
	ctx := c.Request.Context()

	ids := c.QueryArray("ids[]")
	err := sysUserConfigService.DeleteSysUserConfigByIds(ctx, ids)
	if err != nil {
		global.GVA_LOG.Error("批量删除失败!", zap.Error(err))
		response.FailWithMessage("批量删除失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("批量删除成功", c)
}

// UpdateSysUserConfig 更新sysUserConfig表
// @Tags SysUserConfig
// @Summary 更新sysUserConfig表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body example.SysUserConfig true "更新sysUserConfig表"
// @Success 200 {object} response.Response{msg=string} "更新成功"
// @Router /sysUserConfig/updateSysUserConfig [put]
func (sysUserConfigApi *SysUserConfigApi) UpdateSysUserConfig(c *gin.Context) {
	// 从ctx获取标准context进行业务行为
	ctx := c.Request.Context()

	var sysUserConfig example.SysUserConfig
	err := c.ShouldBindJSON(&sysUserConfig)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	err = sysUserConfigService.UpdateSysUserConfig(ctx, sysUserConfig)
	if err != nil {
		global.GVA_LOG.Error("更新失败!", zap.Error(err))
		response.FailWithMessage("更新失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// FindSysUserConfig 用id查询sysUserConfig表
// @Tags SysUserConfig
// @Summary 用id查询sysUserConfig表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param id query int true "用id查询sysUserConfig表"
// @Success 200 {object} response.Response{data=example.SysUserConfig,msg=string} "查询成功"
// @Router /sysUserConfig/findSysUserConfig [get]
func (sysUserConfigApi *SysUserConfigApi) FindSysUserConfig(c *gin.Context) {
	// 创建业务用Context
	ctx := c.Request.Context()

	id := c.Query("id")
	resysUserConfig, err := sysUserConfigService.GetSysUserConfig(ctx, id)
	if err != nil {
		global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithMessage("查询失败:"+err.Error(), c)
		return
	}
	// 转换为仅包含需要字段的响应结构
	dto := exampleResp.NewSysUserConfigItem(resysUserConfig)
	response.OkWithData(dto, c)
}

// GetSysUserConfigList 分页获取sysUserConfig表列表
// @Tags SysUserConfig
// @Summary 分页获取sysUserConfig表列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query exampleReq.SysUserConfigSearch true "分页获取sysUserConfig表列表"
// @Success 200 {object} response.Response{data=response.PageResult,msg=string} "获取成功"
// @Router /sysUserConfig/getSysUserConfigList [get]
func (sysUserConfigApi *SysUserConfigApi) GetSysUserConfigList(c *gin.Context) {
	// 创建业务用Context
	ctx := c.Request.Context()

	var pageInfo exampleReq.SysUserConfigSearch
	err := c.ShouldBindQuery(&pageInfo)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := sysUserConfigService.GetSysUserConfigInfoList(ctx, pageInfo)
	if err != nil {
		global.GVA_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败:"+err.Error(), c)
		return
	}
	// 转换为仅包含需要字段的响应结构列表
	items := exampleResp.NewSysUserConfigItemList(list)
	response.OkWithDetailed(response.PageResult{
		List:     items,
		Total:    total,
		Page:     pageInfo.Page,
		PageSize: pageInfo.PageSize,
	}, "获取成功", c)
}

// GetSysUserConfigPublic 不需要鉴权的sysUserConfig表接口
// @Tags SysUserConfig
// @Summary 不需要鉴权的sysUserConfig表接口
// @Accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{data=object,msg=string} "获取成功"
// @Router /sysUserConfig/getSysUserConfigPublic [get]
func (sysUserConfigApi *SysUserConfigApi) GetSysUserConfigPublic(c *gin.Context) {
	// 创建业务用Context
	ctx := c.Request.Context()

	// 此接口不需要鉴权
	// 示例为返回了一个固定的消息接口，一般本接口用于C端服务，需要自己实现业务逻辑
	sysUserConfigService.GetSysUserConfigPublic(ctx)
	response.OkWithDetailed(gin.H{
		"info": "不需要鉴权的sysUserConfig表接口信息",
	}, "获取成功", c)
}

// UpdateSysUserConfigValue 仅根据 name+formId 更新 value 字段
// @Tags SysUserConfig
// @Summary 更新指定 name+formId 的 value
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body exampleReq.SysUserConfigUpdateValueReq true "更新指定 name+formId 的 value"
// @Success 200 {object} response.Response{msg=string} "更新成功"
// @Router /sysUserConfig/updateValueByNameForm [put]
func (sysUserConfigApi *SysUserConfigApi) UpdateSysUserConfigValue(c *gin.Context) {
	// 创建业务用Context
	ctx := c.Request.Context()

	var req exampleReq.SysUserConfigUpdateValueReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := sysUserConfigService.UpdateValueByNameForm(ctx, req.Name, req.FormId, req.Value); err != nil {
		global.GVA_LOG.Error("更新失败!", zap.Error(err))
		response.FailWithMessage("更新失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}
