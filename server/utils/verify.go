package utils

var (
	IdVerify       = Rules{"ID": []string{NotEmpty()}}
	ApiVerify      = Rules{"Path": {NotEmpty()}, "Description": {NotEmpty()}, "ApiGroup": {NotEmpty()}, "Method": {NotEmpty()}}
	MenuVerify     = Rules{"Path": {NotEmpty()}, "Name": {NotEmpty()}, "Component": {NotEmpty()}, "Sort": {Ge("0")}}
	MenuMetaVerify = Rules{"Title": {NotEmpty()}}
	LoginVerify    = Rules{"Username": {NotEmpty()}, "Password": {NotEmpty()}}
	RegisterVerify = Rules{"Username": {NotEmpty()}, "NickName": {NotEmpty()}, "Password": {NotEmpty()}, "AuthorityId": {NotEmpty()}}
	PageInfoVerify = Rules{"Page": {NotEmpty()}, "PageSize": {NotEmpty()}}
	CustomerVerify = Rules{"CustomerName": {NotEmpty()}, "CustomerPhoneData": {NotEmpty()}}
	MerUserVerify  = Rules{
		"Username":         {NotEmpty()},
		"Password":         {NotEmpty()},
		"QrCode":           {NotEmpty()},
		"MerType":          {NotEmpty()},
		"MaxDecimalAmount": {Ge("1"), Le("99")}, // 1 <= 值 <= 99
		"MinDecimalAmount": {Ge("1"), Le("99")}, // 1 <= 值 <= 99
		"MaxAmount":        {Ge("1")},           // ≥1 (大于0)
		"MinAmount":        {Ge("1")},           // ≥1 (大于0)
	}
	PayQrcodeParmsVerify = Rules{
		"MerId":       {OptionalPositiveNumber()}, // 商户ID可以为空，不为空时必须是大于0的数字
		"PayAmmount":  {Ge("1")},                  // 支付金额必须大于等于1
		"OrderId":     {NotEmpty()},               // 订单ID不能为空
		"CreateTime":  {NotEmpty()},               // 创建时间不能为空
		"Expires":     {OptionalPositiveNumber()}, // 过期时间可以为空，不为空时必须是大于0的数字
		"CallbackUrl": {NotEmpty()},               // 回调URL不能为空
	}
	AutoCodeVerify         = Rules{"Abbreviation": {NotEmpty()}, "StructName": {NotEmpty()}, "PackageName": {NotEmpty()}}
	AutoPackageVerify      = Rules{"PackageName": {NotEmpty()}}
	AuthorityVerify        = Rules{"AuthorityId": {NotEmpty()}, "AuthorityName": {NotEmpty()}}
	AuthorityIdVerify      = Rules{"AuthorityId": {NotEmpty()}}
	OldAuthorityVerify     = Rules{"OldAuthorityId": {NotEmpty()}}
	ChangePasswordVerify   = Rules{"Password": {NotEmpty()}, "NewPassword": {NotEmpty()}}
	SetUserAuthorityVerify = Rules{"AuthorityId": {NotEmpty()}}
)
