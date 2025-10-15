package request

type PayQrcodeParms struct {
	MerId      *int32 `json:"merId"`
	PayAmmount int64  `json:"payAmmount"`
	OrderId    string `json:"orderId"`
	MerType    string `json:"merType"`
	CreateTime string `json:"createTime"`
	Expires    int64  `json:"expires"`
}
