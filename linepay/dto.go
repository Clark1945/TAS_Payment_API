package linepay

// LinePayRequest 代表發送至 Line Pay API 的請求
type LinePayAuthRequest struct {
	Amount       int          `json:"amount"`
	Currency     string       `json:"currency"`
	OrderID      string       `json:"orderId"`
	Packages     []Package    `json:"packages"`
	RedirectUrls RedirectUrls `json:"redirectUrls"`
}

type Package struct {
	ID       string    `json:"id"`
	Amount   int       `json:"amount"`
	Products []Product `json:"products"`
}

// Product 代表單一商品資訊
type Product struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	ImageURL string `json:"imageUrl"`
	Quantity int    `json:"quantity"`
	Price    int    `json:"price"`
}

// RedirectUrls 代表交易完成或取消後的跳轉網址
type RedirectUrls struct {
	ConfirmUrl string `json:"confirmUrl"`
	CancelUrl  string `json:"cancelUrl"`
}

type LinePayResp struct {
	Info          LinepayRespInfo `json:"info"`
	ReturnCode    string          `json:"returnCode"`
	ReturnMessage string          `json:"returnMessage"`
}

type LinepayRespInfo struct {
	Token      string     `json:"paymentAccessToken"`
	PaymentUrl PaymentUrl `json:"paymentUrl"`
	TxId       int        `json:"transactionId"`
}
type PaymentUrl struct {
	App string `json:"app"`
	Web string `json:"web"`
}

// 定義 Info 結構
type Info struct {
	OrderID       string `json:"orderId"`
	TransactionID int64  `json:"transactionId"`
}

// 定義根結構
type CallbackResp struct {
	ReturnCode    string `json:"returnCode"`
	ReturnMessage string `json:"returnMessage"`
	Info          Info   `json:"info"`
}

type MerchantAuthReq struct {
	TransactionId string            `json:"transactionId"`
	Correlationid string            `json:"correlationid"`
	TotalPrice    int               `json:"totalPrice"`
	Currency      string            `json:"currency"`
	PaymentConfig map[string]string `json:"paymentConfig"`
	ReqTime       string            `json:"reqTime"`
	CallbackUrl   string            `json:"callbackUrl"`
	NotifyUrl     string            `json:"notifyUrl"`
	// 加密簽章參數
}

type MerchantAuthResp struct {
	TransactionId string `json:"transactionId"`
	LegacyId      string `json:"legacyId"`
	Correlationid string `json:"correlationid"`
	PaymentUrl    string `json:"paymentUrl"`
	Currency      string `json:"currency"`
	CallTime      string `json:"callTime"`
	Status        string `json:"status"`
	RespCode      string `json:"respCode"`
	RespMsg       string `json:"respMsg"`
	// 加密簽章參數
}
