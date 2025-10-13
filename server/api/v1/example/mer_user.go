package example

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/model/example"
	exampleReq "github.com/flipped-aurora/gin-vue-admin/server/model/example/request"
	//exampleRes "github.com/flipped-aurora/gin-vue-admin/server/model/example/response"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type MerUserApi struct{}

// CreateMerUser 创建merUser表
// @Tags MerUser
// @Summary 创建merUser表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body example.MerUser true "创建merUser表"
// @Success 200 {object} response.Response{msg=string} "创建成功"
// @Router /merUser/createMerUser [post]
func (merUserApi *MerUserApi) CreateMerUser(c *gin.Context) {
	// 创建业务用Context
	ctx := c.Request.Context()

	var merUser example.MerUser
	err := c.ShouldBindJSON(&merUser)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	err = utils.Verify(merUser, utils.MerUserVerify)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	err = merUserService.CreateMerUser(ctx, &merUser)
	if err != nil {
		global.GVA_LOG.Error("创建失败!", zap.Error(err))
		response.FailWithMessage("创建失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

// DeleteMerUser 删除merUser表
// @Tags MerUser
// @Summary 删除merUser表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body example.MerUser true "删除merUser表"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /merUser/deleteMerUser [delete]
func (merUserApi *MerUserApi) DeleteMerUser(c *gin.Context) {
	// 创建业务用Context
	ctx := c.Request.Context()

	id := c.Query("id")
	err := merUserService.DeleteMerUser(ctx, id)
	if err != nil {
		global.GVA_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// DeleteMerUserByIds 批量删除merUser表
// @Tags MerUser
// @Summary 批量删除merUser表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{msg=string} "批量删除成功"
// @Router /merUser/deleteMerUserByIds [delete]
func (merUserApi *MerUserApi) DeleteMerUserByIds(c *gin.Context) {
	// 创建业务用Context
	ctx := c.Request.Context()

	ids := c.QueryArray("ids[]")
	err := merUserService.DeleteMerUserByIds(ctx, ids)
	if err != nil {
		global.GVA_LOG.Error("批量删除失败!", zap.Error(err))
		response.FailWithMessage("批量删除失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("批量删除成功", c)
}

// UpdateMerUser 更新merUser表
// @Tags MerUser
// @Summary 更新merUser表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body example.MerUser true "更新merUser表"
// @Success 200 {object} response.Response{msg=string} "更新成功"
// @Router /merUser/updateMerUser [put]
func (merUserApi *MerUserApi) UpdateMerUser(c *gin.Context) {
	// 从ctx获取标准context进行业务行为
	ctx := c.Request.Context()

	var merUser example.MerUser
	err := c.ShouldBindJSON(&merUser)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	err = utils.Verify(merUser, utils.MerUserVerify)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	err = merUserService.UpdateMerUser(ctx, merUser)
	if err != nil {
		global.GVA_LOG.Error("更新失败!", zap.Error(err))
		response.FailWithMessage("更新失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// FindMerUser 用id查询merUser表
// @Tags MerUser
// @Summary 用id查询merUser表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param id query int true "用id查询merUser表"
// @Success 200 {object} response.Response{data=example.MerUser,msg=string} "查询成功"
// @Router /merUser/findMerUser [get]
func (merUserApi *MerUserApi) FindMerUser(c *gin.Context) {
	// 创建业务用Context
	ctx := c.Request.Context()

	id := c.Query("id")
	remerUser, err := merUserService.GetMerUser(ctx, id)
	if err != nil {
		global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithMessage("查询失败:"+err.Error(), c)
		return
	}
	response.OkWithData(remerUser, c)
}

// GetMerUserList 分页获取merUser表列表
// @Tags MerUser
// @Summary 分页获取merUser表列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query exampleReq.MerUserSearch true "分页获取merUser表列表"
// @Success 200 {object} response.Response{data=response.PageResult,msg=string} "获取成功"
// @Router /merUser/getMerUserList [get]
func (merUserApi *MerUserApi) GetMerUserList(c *gin.Context) {
	// 创建业务用Context
	ctx := c.Request.Context()
	var pageInfo exampleReq.MerUserSearch
	err := c.ShouldBindQuery(&pageInfo)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := merUserService.GetMerUserInfoList(ctx, pageInfo)
	if err != nil {
		global.GVA_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败:"+err.Error(), c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List:     list,
		Total:    total,
		Page:     pageInfo.Page,
		PageSize: pageInfo.PageSize,
	}, "获取成功", c)
}

// GetMerUserPublic 不需要鉴权的merUser表接口
// @Tags MerUser
// @Summary 不需要鉴权的merUser表接口
// @Accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{data=object,msg=string} "获取成功"
// @Router /merUser/getMerUserPublic [get]
func (merUserApi *MerUserApi) GetMerUserPublic(c *gin.Context) {
	// 创建业务用Context
	ctx := c.Request.Context()

	// 此接口不需要鉴权
	// 示例为返回了一个固定的消息接口，一般本接口用于C端服务，需要自己实现业务逻辑
	merUserService.GetMerUserPublic(ctx)
	response.OkWithDetailed(gin.H{
		"info": "不需要鉴权的merUser表接口信息",
	}, "获取成功", c)
}

// GetMerUserList 分页获取merUser表列表
// @Tags MerUser
// @Summary 分页获取merUser表列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query exampleReq.MerUserSearch true "分页获取merUser表列表"
// @Success 200 {object} response.Response{data=response.PageResult,msg=string} "获取成功"
// @Router /merUser/getMerUserList [get]
func (merUserApi *MerUserApi) CreatePay(c *gin.Context) {
	// 创建业务用Context
	ctx := c.Request.Context()

	var pageInfo exampleReq.MerUserSearch
	err := c.ShouldBindQuery(&pageInfo)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := merUserService.GetMerUserInfoList(ctx, pageInfo)
	if err != nil {
		global.GVA_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败:"+err.Error(), c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List:     list,
		Total:    total,
		Page:     pageInfo.Page,
		PageSize: pageInfo.PageSize,
	}, "获取成功", c)
}
