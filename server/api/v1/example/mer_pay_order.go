package example

import (
	"fmt"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/model/example"
	exampleReq "github.com/flipped-aurora/gin-vue-admin/server/model/example/request"
	exampleRes "github.com/flipped-aurora/gin-vue-admin/server/model/example/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"strconv"
)

type MerPayOrderApi struct{}

// CreateMerPayOrder 创建merPayOrder表
// @Tags MerPayOrder
// @Summary 创建merPayOrder表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body example.MerPayOrder true "创建merPayOrder表"
// @Success 200 {object} response.Response{msg=string} "创建成功"
// @Router /merPayOrder/createMerPayOrder [post]
func (merPayOrderApi *MerPayOrderApi) CreateMerPayOrder(c *gin.Context) {
	// 创建业务用Context
	ctx := c.Request.Context()

	var merPayOrder example.MerPayOrder
	err := c.ShouldBindJSON(&merPayOrder)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	err = merPayOrderService.CreateMerPayOrder(ctx, &merPayOrder)
	if err != nil {
		global.GVA_LOG.Error("创建失败!", zap.Error(err))
		response.FailWithMessage("创建失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

// DeleteMerPayOrder 删除merPayOrder表
// @Tags MerPayOrder
// @Summary 删除merPayOrder表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body example.MerPayOrder true "删除merPayOrder表"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /merPayOrder/deleteMerPayOrder [delete]
func (merPayOrderApi *MerPayOrderApi) DeleteMerPayOrder(c *gin.Context) {
	// 创建业务用Context
	ctx := c.Request.Context()

	id := c.Query("id")
	err := merPayOrderService.DeleteMerPayOrder(ctx, id)
	if err != nil {
		global.GVA_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// DeleteMerPayOrderByIds 批量删除merPayOrder表
// @Tags MerPayOrder
// @Summary 批量删除merPayOrder表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{msg=string} "批量删除成功"
// @Router /merPayOrder/deleteMerPayOrderByIds [delete]
func (merPayOrderApi *MerPayOrderApi) DeleteMerPayOrderByIds(c *gin.Context) {
	// 创建业务用Context
	ctx := c.Request.Context()

	ids := c.QueryArray("ids[]")
	err := merPayOrderService.DeleteMerPayOrderByIds(ctx, ids)
	if err != nil {
		global.GVA_LOG.Error("批量删除失败!", zap.Error(err))
		response.FailWithMessage("批量删除失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("批量删除成功", c)
}

// UpdateMerPayOrder 更新merPayOrder表
// @Tags MerPayOrder
// @Summary 更新merPayOrder表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body example.MerPayOrder true "更新merPayOrder表"
// @Success 200 {object} response.Response{msg=string} "更新成功"
// @Router /merPayOrder/updateMerPayOrder [put]
func (merPayOrderApi *MerPayOrderApi) UpdateMerPayOrder(c *gin.Context) {
	// 从ctx获取标准context进行业务行为
	ctx := c.Request.Context()

	var merPayOrder example.MerPayOrder
	err := c.ShouldBindJSON(&merPayOrder)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	err = merPayOrderService.UpdateMerPayOrder(ctx, merPayOrder)
	if err != nil {
		global.GVA_LOG.Error("更新失败!", zap.Error(err))
		response.FailWithMessage("更新失败:"+err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{
		"success": true,
	}, "更新成功", c)
}

// FindMerPayOrder 用id查询merPayOrder表
// @Tags MerPayOrder
// @Summary 用id查询merPayOrder表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param id query int true "用id查询merPayOrder表"
// @Success 200 {object} response.Response{data=example.MerPayOrder,msg=string} "查询成功"
// @Router /merPayOrder/findMerPayOrder [get]
func (merPayOrderApi *MerPayOrderApi) FindMerPayOrder(c *gin.Context) {
	// 创建业务用Context
	ctx := c.Request.Context()

	id := c.Query("id")
	remerPayOrder, err := merPayOrderService.GetMerPayOrder(ctx, id)
	if err != nil {
		global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithMessage("查询失败:"+err.Error(), c)
		return
	}
	response.OkWithData(remerPayOrder, c)
}

func (merPayOrderApi *MerPayOrderApi) GetMerPayOrderState(c *gin.Context) {
	// 创建业务用Context
	ctx := c.Request.Context()

	id := c.Query("id")

	// 从gin上下文中获取sys_user_id（由PermanentTokenAuth中间件设置）
	var sysUserIdPtr *int64
	if sysUserIdValue, exists := c.Get("sys_user_id"); exists {
		if sysUserId, ok := sysUserIdValue.(uint); ok {
			sysUserIdInt64 := int64(sysUserId)
			sysUserIdPtr = &sysUserIdInt64
		} else {
			global.GVA_LOG.Error("sys_user_id 类型转换失败")
			response.FailWithMessage("用户ID获取失败", c)
			return
		}
	} else {
		global.GVA_LOG.Error("未找到sys_user_id")
		response.FailWithMessage("用户ID未找到", c)
		return
	}

	remerPayOrder, err := merPayOrderService.GetMerPayOrderByInfo(ctx, example.MerPayOrder{
		OrderId:   &id,
		SysUserId: sysUserIdPtr,
	})
	if err != nil {
		global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithMessage("查询失败:"+err.Error(), c)
		return
	}

	// 创建限制字段的响应结构
	limitedResponse := exampleRes.MerPayOrderStateResponse{
		State:          remerPayOrder.State,
		RequestAmmount: remerPayOrder.RequestAmmount,
		Ammount:        remerPayOrder.Ammount,
		PayTime:        remerPayOrder.PayTime,
	}

	response.OkWithData(limitedResponse, c)
}

func (merPayOrderApi *MerPayOrderApi) CancelMerPayOrder(c *gin.Context) {
	// 创建业务用Context
	ctx := c.Request.Context()
	id := c.Query("id")

	// 从gin上下文中获取sys_user_id（由PermanentTokenAuth中间件设置）
	var sysUserIdPtr *int64
	if sysUserIdValue, exists := c.Get("sys_user_id"); exists {
		if sysUserId, ok := sysUserIdValue.(uint); ok {
			sysUserIdInt64 := int64(sysUserId)
			sysUserIdPtr = &sysUserIdInt64
		} else {
			global.GVA_LOG.Error("sys_user_id 类型转换失败")
			response.FailWithMessage("用户ID获取失败", c)
			return
		}
	} else {
		global.GVA_LOG.Error("未找到sys_user_id")
		response.FailWithMessage("order user not found", c)
		return
	}

	var merPayOrder example.MerPayOrder
	merPayOrder.SysUserId = sysUserIdPtr
	if idInt64, err := strconv.ParseInt(id, 10, 64); err == nil {
		merPayOrder.Id = &idInt64
	} else {
		global.GVA_LOG.Error("订单ID格式错误", zap.Error(err))
		response.FailWithMessage("订单ID格式错误", c)
		return
	}
	// 先查询订单获取完整信息，用于构建Redis键
	remerPayOrder, err := merPayOrderService.GetMerPayOrder(ctx, id)
	if err != nil {
		global.GVA_LOG.Error("查询订单失败!", zap.Error(err))
		response.FailWithMessage("查询订单失败:"+err.Error(), c)
		return
	}

	// 检查订单是否属于当前用户
	if remerPayOrder.SysUserId == nil || *remerPayOrder.SysUserId != *sysUserIdPtr {
		global.GVA_LOG.Error("订单不属于当前用户")
		response.FailWithMessage("无权限操作此订单", c)
		return
	}

	// 检查订单状态，避免重复取消
	if remerPayOrder.State != nil && remerPayOrder.State == global.MER_PAY_ORDER_CANCELED {
		response.OkWithDetailed(gin.H{
			"success": true,
		}, "订单已经是取消状态", c)
		return
	}

	// 构建 Redis 键：pay_amount_used:userID:amount
	// 检查指针是否为nil，防止panic
	if sysUserIdPtr == nil || remerPayOrder.Ammount == nil {
		global.GVA_LOG.Error("构建Redis键失败: sysUserIdPtr或Ammount为nil")
		response.FailWithMessage("订单信息不完整", c)
		return
	}

	// 将decimal转换为字符串格式
	amountStr := remerPayOrder.Ammount.String()
	redisKey := fmt.Sprintf("%s:%d:%s", global.PAY_AMOUNT_USED_KEY, *sysUserIdPtr, amountStr)

	// 更新订单状态为取消
	merPayOrder.State = global.MER_PAY_ORDER_CANCELED
	err = merPayOrderService.UpdateMerPayOrder(ctx, merPayOrder)
	if err != nil {
		global.GVA_LOG.Error("更新订单状态失败!", zap.Error(err))
		response.FailWithMessage("取消订单失败:"+err.Error(), c)
		return
	}

	// 删除Redis中的键和值
	deletedCount, err := global.GVA_REDIS.Del(ctx, redisKey).Result()
	if err != nil {
		global.GVA_LOG.Error("删除Redis键失败!", zap.Error(err))
		// 即使Redis删除失败，也不影响订单取消的成功
	} else {
		global.GVA_LOG.Info("成功删除Redis键", zap.String("key", redisKey), zap.Int64("deletedCount", deletedCount))
	}

	response.OkWithDetailed(gin.H{
		"success": true,
	}, "取消成功", c)
}

// GetMerPayOrderList 分页获取merPayOrder表列表
// @Tags MerPayOrder
// @Summary 分页获取merPayOrder表列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query exampleReq.MerPayOrderSearch true "分页获取merPayOrder表列表"
// @Success 200 {object} response.Response{data=response.PageResult,msg=string} "获取成功"
// @Router /merPayOrder/getMerPayOrderList [get]
func (merPayOrderApi *MerPayOrderApi) GetMerPayOrderList(c *gin.Context) {
	// 创建业务用Context
	ctx := c.Request.Context()

	var pageInfo exampleReq.MerPayOrderSearch
	err := c.ShouldBindQuery(&pageInfo)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := merPayOrderService.GetMerPayOrderInfoList(ctx, pageInfo)
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

// GetMerPayOrderPublic 不需要鉴权的merPayOrder表接口
// @Tags MerPayOrder
// @Summary 不需要鉴权的merPayOrder表接口
// @Accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{data=object,msg=string} "获取成功"
// @Router /merPayOrder/getMerPayOrderPublic [get]
func (merPayOrderApi *MerPayOrderApi) GetMerPayOrderPublic(c *gin.Context) {
	// 创建业务用Context
	ctx := c.Request.Context()

	// 此接口不需要鉴权
	// 示例为返回了一个固定的消息接口，一般本接口用于C端服务，需要自己实现业务逻辑
	merPayOrderService.GetMerPayOrderPublic(ctx)
	response.OkWithDetailed(gin.H{
		"info": "不需要鉴权的merPayOrder表接口信息",
	}, "获取成功", c)
}
