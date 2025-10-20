package example

import (
	api "github.com/flipped-aurora/gin-vue-admin/server/api/v1"
)

type RouterGroup struct {
	CustomerRouter
	FileUploadAndDownloadRouter
	AttachmentCategoryRouter
	MerUserRouter
	SysUserConfigRouter
	MerPayOrderRouter
}

var (
	exaCustomerApi              = api.ApiGroupApp.ExampleApiGroup.CustomerApi
	exaFileUploadAndDownloadApi = api.ApiGroupApp.ExampleApiGroup.FileUploadAndDownloadApi
	attachmentCategoryApi       = api.ApiGroupApp.ExampleApiGroup.AttachmentCategoryApi
	merUserApi                  = api.ApiGroupApp.ExampleApiGroup.MerUserApi
	sysUserConfigApi            = api.ApiGroupApp.ExampleApiGroup.SysUserConfigApi
	merPayOrderApi              = api.ApiGroupApp.ExampleApiGroup.MerPayOrderApi
)
