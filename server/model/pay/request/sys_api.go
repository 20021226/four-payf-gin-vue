package request

type PayQrcodeParms struct {
	MerId       *int64 `json:"merId"`
	PayAmmount  int64  `json:"payAmmount"`
	OrderId     string `json:"orderId"`
	CreateTime  string `json:"createTime"`
	Expires     int64  `json:"expires"`
	CallbackUrl string `json:"callbackUrl"` // 回调URL
}
