import service from '@/utils/request'
// @Tags SysUserConfig
// @Summary 创建sysUserConfig表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.SysUserConfig true "创建sysUserConfig表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"创建成功"}"
// @Router /sysUserConfig/createSysUserConfig [post]
export const createSysUserConfig = (data) => {
  return service({
    url: '/sysUserConfig/createSysUserConfig',
    method: 'post',
    data
  })
}

// @Tags SysUserConfig
// @Summary 删除sysUserConfig表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.SysUserConfig true "删除sysUserConfig表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"删除成功"}"
// @Router /sysUserConfig/deleteSysUserConfig [delete]
export const deleteSysUserConfig = (params) => {
  return service({
    url: '/sysUserConfig/deleteSysUserConfig',
    method: 'delete',
    params
  })
}

// @Tags SysUserConfig
// @Summary 批量删除sysUserConfig表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body request.IdsReq true "批量删除sysUserConfig表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"删除成功"}"
// @Router /sysUserConfig/deleteSysUserConfig [delete]
export const deleteSysUserConfigByIds = (params) => {
  return service({
    url: '/sysUserConfig/deleteSysUserConfigByIds',
    method: 'delete',
    params
  })
}

// @Tags SysUserConfig
// @Summary 更新sysUserConfig表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.SysUserConfig true "更新sysUserConfig表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"更新成功"}"
// @Router /sysUserConfig/updateSysUserConfig [put]
export const updateSysUserConfig = (data) => {
  return service({
    url: '/sysUserConfig/updateSysUserConfig',
    method: 'put',
    data
  })
}

// @Tags SysUserConfig
// @Summary 用id查询sysUserConfig表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query model.SysUserConfig true "用id查询sysUserConfig表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"查询成功"}"
// @Router /sysUserConfig/findSysUserConfig [get]
export const findSysUserConfig = (params) => {
  return service({
    url: '/sysUserConfig/findSysUserConfig',
    method: 'get',
    params
  })
}

// @Tags SysUserConfig
// @Summary 分页获取sysUserConfig表列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query request.PageInfo true "分页获取sysUserConfig表列表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"获取成功"}"
// @Router /sysUserConfig/getSysUserConfigList [get]
export const getSysUserConfigList = (params) => {
  return service({
    url: '/sysUserConfig/getSysUserConfigList',
    method: 'get',
    params
  })
}

// @Tags SysUserConfig
// @Summary 不需要鉴权的sysUserConfig表接口
// @Accept application/json
// @Produce application/json
// @Param data query exampleReq.SysUserConfigSearch true "分页获取sysUserConfig表列表"
// @Success 200 {object} response.Response{data=object,msg=string} "获取成功"
// @Router /sysUserConfig/getSysUserConfigPublic [get]
export const getSysUserConfigPublic = () => {
  return service({
    url: '/sysUserConfig/getSysUserConfigPublic',
    method: 'get',
  })
}
