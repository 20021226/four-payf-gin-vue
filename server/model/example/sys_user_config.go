
// 自动生成模板SysUserConfig
package example
import (
	"time"
)

// sysUserConfig表 结构体  SysUserConfig
type SysUserConfig struct {
  Id  *int32 `json:"id" form:"id" gorm:"uniqueIndex;primarykey;comment:配置id;column:id;"`  //配置id
  SysUserId  *int64 `json:"sysUserId" form:"sysUserId" gorm:"index;comment:用户id;column:sys_user_id;"`  //用户id
  Name  *string `json:"name" form:"name" gorm:"comment:字段名称;column:name;size:255;"`  //字段名称
  Title  *string `json:"title" form:"title" gorm:"comment:字段提示文字;column:title;size:255;"`  //字段提示文字
  FormId  *int32 `json:"formId" form:"formId" gorm:"comment:表单id;column:form_id;"`  //表单id
  Value  *string `json:"value" form:"value" gorm:"comment:值;column:value;size:191;"`  //值
  Status  *bool `json:"status" form:"status" gorm:"comment:是否隐藏;column:status;"`  //是否隐藏
  CreateTime  *time.Time `json:"createTime" form:"createTime" gorm:"comment:创建时间;column:create_time;"`  //创建时间
  UpdateTime  *time.Time `json:"updateTime" form:"updateTime" gorm:"comment:更新时间;column:update_time;"`  //更新时间
}


// TableName sysUserConfig表 SysUserConfig自定义表名 sys_user_config
func (SysUserConfig) TableName() string {
    return "sys_user_config"
}





