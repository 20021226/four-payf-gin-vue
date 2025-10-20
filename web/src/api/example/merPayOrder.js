import service from '@/utils/request'
// @Tags MerPayOrder
// @Summary 创建merPayOrder表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.MerPayOrder true "创建merPayOrder表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"创建成功"}"
// @Router /merPayOrder/createMerPayOrder [post]
export const createMerPayOrder = (data) => {
  return service({
    url: '/merPayOrder/createMerPayOrder',
    method: 'post',
    data
  })
}

// @Tags MerPayOrder
// @Summary 删除merPayOrder表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.MerPayOrder true "删除merPayOrder表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"删除成功"}"
// @Router /merPayOrder/deleteMerPayOrder [delete]
export const deleteMerPayOrder = (params) => {
  return service({
    url: '/merPayOrder/deleteMerPayOrder',
    method: 'delete',
    params
  })
}

// @Tags MerPayOrder
// @Summary 批量删除merPayOrder表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body request.IdsReq true "批量删除merPayOrder表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"删除成功"}"
// @Router /merPayOrder/deleteMerPayOrder [delete]
export const deleteMerPayOrderByIds = (params) => {
  return service({
    url: '/merPayOrder/deleteMerPayOrderByIds',
    method: 'delete',
    params
  })
}

// @Tags MerPayOrder
// @Summary 更新merPayOrder表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.MerPayOrder true "更新merPayOrder表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"更新成功"}"
// @Router /merPayOrder/updateMerPayOrder [put]
export const updateMerPayOrder = (data) => {
  return service({
    url: '/merPayOrder/updateMerPayOrder',
    method: 'put',
    data
  })
}

// @Tags MerPayOrder
// @Summary 用id查询merPayOrder表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query model.MerPayOrder true "用id查询merPayOrder表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"查询成功"}"
// @Router /merPayOrder/findMerPayOrder [get]
export const findMerPayOrder = (params) => {
  return service({
    url: '/merPayOrder/findMerPayOrder',
    method: 'get',
    params
  })
}

// @Tags MerPayOrder
// @Summary 分页获取merPayOrder表列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query request.PageInfo true "分页获取merPayOrder表列表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"获取成功"}"
// @Router /merPayOrder/getMerPayOrderList [get]
export const getMerPayOrderList = (params) => {
  return service({
    url: '/merPayOrder/getMerPayOrderList',
    method: 'get',
    params
  })
}

// @Tags MerPayOrder
// @Summary 不需要鉴权的merPayOrder表接口
// @Accept application/json
// @Produce application/json
// @Param data query exampleReq.MerPayOrderSearch true "分页获取merPayOrder表列表"
// @Success 200 {object} response.Response{data=object,msg=string} "获取成功"
// @Router /merPayOrder/getMerPayOrderPublic [get]
export const getMerPayOrderPublic = () => {
  return service({
    url: '/merPayOrder/getMerPayOrderPublic',
    method: 'get',
  })
}
