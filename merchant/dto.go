package merchant

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
	LegacyId      string `json:"legacyId,omitempty"`
	Correlationid string `json:"correlationid"`
	PaymentUrl    string `json:"paymentUrl,omitempty"`
	Currency      string `json:"currency,omitempty"`
	CallTime      string `json:"callTime"`
	Status        string `json:"status"`
	RespCode      string `json:"respCode"`
	RespMsg       string `json:"respMsg"`
	// 加密簽章參數
}

type MerchantQueryReq struct {
	LegacyID      string            `json:"legacyId"`
	TransactionID string            `json:"transactionId"`
	PaymentConfig map[string]string `json:"paymentConfig"`
}

type MerchantRefundReq struct {
	TransctionId  string            `json:"transactionId"`
	LegacyId      string            `json:"legacyId"`
	Correlationid string            `json:"correlationid"`
	RefundNo      string            `json:"refundNo"`
	RefundDate    string            `json:"refundDate"`
	RefundPrice   int               `json:"refundPrice"`
	PaymentConfig map[string]string `json:"paymentConfig"`
}

type MerchantCallbackResp struct {
	TransactionId string `json:"transactionId"`
	LegacyId      string `json:"legacyId"`
	RespCode      string `json:"respCode"`
	RespMsg       string `json:"respMsg"`
	TotalFee      int    `json:"totalFee,omitempty"` // if null not return
	PaidTime      string `json:"paidTime,omitempty"`
	Status        string `json:"status"`
}
