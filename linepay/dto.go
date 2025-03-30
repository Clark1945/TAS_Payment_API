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

type LinePayAuthResponse struct {
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

type QueryResponse struct {
	ReturnCode    string      `json:"returnCode"`
	ReturnMessage string      `json:"returnMessage"`
	QueryInfo     []QueryInfo `json:"info"`
}

type QueryInfo struct {
	AuthorizationExpireDate string         `json:"authorizationExpireDate"`
	Currency                string         `json:"currency"`
	OrderID                 string         `json:"orderId"`
	PayStatus               string         `json:"payStatus"`
	ProductName             string         `json:"productName"`
	TransactionDate         string         `json:"transactionDate"`
	TransactionID           float64        `json:"transactionId"` // 用 string 避免數字精度丟失
	TransactionType         string         `json:"transactionType"`
	Info                    []QueryPayInfo `json:"payInfo"`
	Packages                []QueryPackage `json:"packages"`
}

type QueryPayInfo struct {
	Amount int    `json:"amount"`
	Method string `json:"method"`
}

type QueryPackage struct {
	Amount        int       `json:"amount"`
	ID            string    `json:"id"`
	Products      []Product `json:"products"`
	UserFeeAmount int       `json:"userFeeAmount"`
}

type RefundRespInfo struct {
	Info       RefundInfo `json:"info"`
	ReturnCode string     `json:"returnCode"`
	ReturnMsg  string     `json:"returnMessage"`
}
type RefundInfo struct {
	RefundTransactionDate string  `json:"refundTransactionDate"`
	RefundTransactionId   float64 `json:"refundTransactionId"`
}
