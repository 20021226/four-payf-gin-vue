import service from '@/utils/request'
// @Tags MerUser
// @Summary 创建merUser表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.MerUser true "创建merUser表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"创建成功"}"
// @Router /merUser/createMerUser [post]
export const createMerUser = (data) => {
  return service({
    url: '/merUser/createMerUser',
    method: 'post',
    data
  })
}

// @Tags MerUser
// @Summary 删除merUser表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.MerUser true "删除merUser表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"删除成功"}"
// @Router /merUser/deleteMerUser [delete]
export const deleteMerUser = (params) => {
  return service({
    url: '/merUser/deleteMerUser',
    method: 'delete',
    params
  })
}

// @Tags MerUser
// @Summary 批量删除merUser表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body request.IdsReq true "批量删除merUser表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"删除成功"}"
// @Router /merUser/deleteMerUser [delete]
export const deleteMerUserByIds = (params) => {
  return service({
    url: '/merUser/deleteMerUserByIds',
    method: 'delete',
    params
  })
}

// @Tags MerUser
// @Summary 更新merUser表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.MerUser true "更新merUser表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"更新成功"}"
// @Router /merUser/updateMerUser [put]
export const updateMerUser = (data) => {
  return service({
    url: '/merUser/updateMerUser',
    method: 'put',
    data
  })
}

// @Tags MerUser
// @Summary 用id查询merUser表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query model.MerUser true "用id查询merUser表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"查询成功"}"
// @Router /merUser/findMerUser [get]
export const findMerUser = (params) => {
  return service({
    url: '/merUser/findMerUser',
    method: 'get',
    params
  })
}

// @Tags MerUser
// @Summary 分页获取merUser表列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query request.PageInfo true "分页获取merUser表列表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"获取成功"}"
// @Router /merUser/getMerUserList [get]
export const getMerUserList = (params) => {
  return service({
    url: '/merUser/getMerUserList',
    method: 'get',
    params
  })
}

// @Tags MerUser
// @Summary 不需要鉴权的merUser表接口
// @Accept application/json
// @Produce application/json
// @Param data query exampleReq.MerUserSearch true "分页获取merUser表列表"
// @Success 200 {object} response.Response{data=object,msg=string} "获取成功"
// @Router /merUser/getMerUserPublic [get]
export const getMerUserPublic = () => {
  return service({
    url: '/merUser/getMerUserPublic',
    method: 'get',
  })
}
