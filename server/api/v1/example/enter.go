package example

import "github.com/flipped-aurora/gin-vue-admin/server/service"

type ApiGroup struct {
	CustomerApi
	FileUploadAndDownloadApi
	AttachmentCategoryApi
	MerUserApi
	SysUserConfigApi
	MerPayOrderApi
}

var (
	customerService              = service.ServiceGroupApp.ExampleServiceGroup.CustomerService
	fileUploadAndDownloadService = service.ServiceGroupApp.ExampleServiceGroup.FileUploadAndDownloadService
	attachmentCategoryService    = service.ServiceGroupApp.ExampleServiceGroup.AttachmentCategoryService
	merUserService               = service.ServiceGroupApp.ExampleServiceGroup.MerUserService
	sysUserConfigService         = service.ServiceGroupApp.ExampleServiceGroup.SysUserConfigService
	merPayOrderService           = service.ServiceGroupApp.ExampleServiceGroup.MerPayOrderService
)
