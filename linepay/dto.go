package linepay

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

// LinePayRequest 代表發送至 Line Pay API 的請求
type LinePayRequest struct {
	Amount       int          `json:"amount"`
	Currency     string       `json:"currency"`
	OrderID      string       `json:"orderId"`
	Packages     []Package    `json:"packages"`
	RedirectUrls RedirectUrls `json:"redirectUrls"`
}

// Package 代表訂單內的商品包
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
