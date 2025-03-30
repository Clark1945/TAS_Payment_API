package linepay

import (
	"TAS_payment/merchant"
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// LinePay SandBox ID:test_202503193091@line.pay
// PW: nCb72ySCQ&
// 通路密鑰:c63b2a81aadfeddc3ea33e572e29942a
// 通路ID:2007090890
// 商店ID:test_202503193091
// 統編:24941093
const (
	domain             string = "https://sandbox-api-pay.line.me"
	authPath           string = "/v3/payments/request"
	callbackPath       string = "/v3/payments/%s/confirm"
	queryPath          string = "/v3/payments?transactionId=%s&orderId=%s"
	refundPath         string = "/v3/payments/%s/refund"
	merId                     = "test_202503193091"
	callbackUrl        string = "http://localhost:8080/callback"
	failCallbackUrl    string = "http://localhost:8080/failCallback"
	StatusSuccess             = "0000"
	StatusAuth                = "AUTH"
	StatusAuthFail            = "AUTH_FAIL"
	StatusPaidSuccess         = "PAID_SUCCESS"
	ErrOnlyPostAllowed        = "Only POST allowed"
	ErrRequestBody            = "Request Body Error"
	ErrResponse               = "Response Error"
)

var tempMap = make(map[string]map[string]string)

// API Handler: 對外提供 `/auth` 端點
func AuthHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		merchant.RespondWithError(w, ErrOnlyPostAllowed, http.StatusMethodNotAllowed)
		return
	}
	fmt.Println("auth started")

	// 原先做法 io.Reader轉成Byte，再由Byte轉換成物件
	// 更好做法 使用json.NewDecoder讀取reqBody並直接轉換成物件
	var merchantReq merchant.MerchantAuthReq
	if err := json.NewDecoder(req.Body).Decode(&merchantReq); err != nil {
		merchant.RespondWithError(w, ErrRequestBody, http.StatusBadRequest)
		return
	}

	chanId := merchantReq.PaymentConfig["channelId"]
	chanSec := merchantReq.PaymentConfig["channelSec"]
	linePayReq := newAuthReq(
		&merchantReq.TransactionId,
		&merchantReq.TotalPrice,
		&merchantReq.Currency,
	)

	linePayResp, err := auth(&linePayReq, &chanId, &chanSec)
	if err != nil {
		merchant.RespondWithError(w, fmt.Sprintf("Auth failed: %v", err), http.StatusInternalServerError)
		return
	}

	merchant.LogWithResponse(linePayResp)

	merResp := merchant.MerchantAuthResp{
		Correlationid: merchantReq.Correlationid,
		TransactionId: merchantReq.TransactionId,
		CallTime:      time.Now().Format("2006-01-02 15:04:05"),
		RespCode:      linePayResp.ReturnCode,
		RespMsg:       linePayResp.ReturnMessage,
	}

	if linePayResp.ReturnCode == StatusSuccess { // Payment Success
		merResp.LegacyId = strconv.Itoa(linePayResp.Info.TxId)
		merResp.PaymentUrl = linePayResp.Info.PaymentUrl.Web
		merResp.Currency = merchantReq.Currency
		merResp.Status = StatusAuth

		// 暫代緩存
		tempMap[merResp.LegacyId] = map[string]string{
			"transactionId": merResp.TransactionId,
			"channelId":     chanId,
			"channelSec":    chanSec,
			"totalPrice":    strconv.Itoa(merchantReq.TotalPrice),
			"currency":      merchantReq.Currency}
	} else {
		merResp.Status = "AUTH_FAIL"
	}

	// 原先做法，將struct轉成byte，再從byte轉成map回傳
	// 後續做法，直接回傳struct
	if err := merchant.RespondWithJSON(w, merResp); err != nil {
		merchant.RespondWithError(w, ErrResponse, http.StatusInternalServerError)
	}
	fmt.Println("auth ended")
}

func CallbackHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Starting LINE Pay callback")
	legacyId := r.URL.Query().Get("transactionId")
	chanId := tempMap[legacyId]["channelId"]
	chanSec := tempMap[legacyId]["channelSec"]
	transactionId := tempMap[legacyId]["transactionId"]

	totalFee, err := strconv.Atoi(tempMap[legacyId]["totalPrice"])
	if err != nil {
		log.Println(err)
		return
	}
	reqMap := map[string]interface{}{
		"amount":   totalFee,
		"currency": tempMap[legacyId]["currency"],
	}

	// 處理規格
	reqBodyByte, err := json.Marshal(reqMap)
	if err != nil {
		return
	}
	req, err := http.NewRequest(http.MethodPost, domain+fmt.Sprintf(callbackPath, legacyId), bytes.NewReader(reqBodyByte))
	if err != nil { // reqBody Error!
		log.Println(err)
		return
	}
	reqBodyStr := string(reqBodyByte)
	appendHeader(req, &reqBodyStr, fmt.Sprintf(callbackPath, legacyId), chanId, chanSec)

	// 發送請求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		respondWithError(w, "Error while sending callback", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close() // Delay disconnect the connection.

	var callbackResp CallbackResp
	if err = json.NewDecoder(resp.Body).Decode(&callbackResp); err != nil {
		respondWithError(w, "Error while parsing callback", http.StatusInternalServerError)
		return
	}

	logWithResponse(callbackResp)

	merchantCallbackResp := NewMerchantCallbackResp(transactionId, legacyId, callbackResp.ReturnCode, callbackResp.ReturnMessage, totalFee)
	respondWithJSON(w, merchantCallbackResp)
}

// TODO 檢查 請求

// // 將交易資訊存入資料庫/緩存的函數
// func storeTransactionData(legacyID string, req MerchantAuthReq, chanID, chanSec string) {
// 	// 使用互斥鎖保護 tempMap 操作，或更好的方式是使用資料庫/Redis等
// 	tempMapMutex.Lock()
// 	defer tempMapMutex.Unlock()

// 	tempMap[legacyID] = map[string]string{
// 		"transactionId": req.TransactionId,
// 		"channelId":     chanID,
// 		"channelSec":    chanSec,
// 		"totalPrice":    strconv.Itoa(req.TotalPrice),
// 		"currency":      req.Currency,
// 	}
// }

// 查詢API
func QueryHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Println("Query started")

	var merchantQueryReq merchant.MerchantQueryReq
	if err := json.NewDecoder(req.Body).Decode(&merchantQueryReq); err != nil {
		merchant.RespondWithError(w, ErrRequestBody, http.StatusBadRequest)
		return
	}

	fullQueryPath := fmt.Sprintf(domain+queryPath, merchantQueryReq.LegacyID, merchantQueryReq.TransactionID)
	request, err := http.NewRequest("GET", fullQueryPath, nil)
	if err != nil {
		merchant.RespondWithError(w, ErrRequestBody, http.StatusBadRequest)
		return
	}

	slice := strings.Split(fmt.Sprintf(queryPath, merchantQueryReq.LegacyID, merchantQueryReq.TransactionID), "?")
	appendHeader(request, &slice[1], slice[0], merchantQueryReq.PaymentConfig["channelId"], merchantQueryReq.PaymentConfig["channelSec"])

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Println("連線異常, ", err)
		return
	}
	defer resp.Body.Close()

	var queryResp QueryResponse
	if err = json.NewDecoder(resp.Body).Decode(&queryResp); err != nil {
		merchant.RespondWithError(w, "Error while parsing response", http.StatusBadRequest)
		return
	}
	merchant.LogWithResponse(queryResp)
	respMap := map[string]interface{}{
		"transactionId": merchantQueryReq.TransactionID,
		"legacyId":      merchantQueryReq.LegacyID,
		"respCode":      queryResp.ReturnCode,
		"respMsg":       queryResp.ReturnMessage,
	}
	if queryResp.ReturnCode == StatusSuccess {
		respMap["status"] = "SUCCESS"
		respMap["payStatus"] = queryResp.QueryInfo[0].PayStatus
		respMap["transactionDate"] = queryResp.QueryInfo[0].TransactionDate
	}
	merchant.RespondWithJSON(w, respMap)
	fmt.Println("Query end")
}

// 退款API
func RefundHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Println("Refund started")

	var merchantRefundReq merchant.MerchantRefundReq
	if err := json.NewDecoder(req.Body).Decode(&merchantRefundReq); err != nil {
		merchant.RespondWithError(w, ErrRequestBody, http.StatusBadRequest)
		return
	}

	reqMap := map[string]int{
		"refundAmount": merchantRefundReq.RefundPrice,
	}
	reqByte, err := json.Marshal(reqMap)
	if err != nil {
		merchant.RespondWithError(w, ErrRequestBody, http.StatusBadRequest)
	}

	path := fmt.Sprintf(refundPath, merchantRefundReq.LegacyId)
	request, err := http.NewRequest(http.MethodPost, domain+path, bytes.NewReader(reqByte))
	if err != nil {
		log.Println(err)
		return
	}

	reqBodyStr := string(reqByte)
	chanId := merchantRefundReq.PaymentConfig["channelId"]
	chanSec := merchantRefundReq.PaymentConfig["channelSec"]
	appendHeader(request, &reqBodyStr, path, chanId, chanSec)

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Println("連線異常, ", err)
		return
	}

	var refundRespInfo RefundRespInfo
	if err = json.NewDecoder(resp.Body).Decode(&refundRespInfo); err != nil {
		merchant.RespondWithError(w, "Error with parsing response", http.StatusInternalServerError)
	}
	merchant.LogWithResponse(refundRespInfo)

	respMap := map[string]interface{}{
		"transactionId": merchantRefundReq.TransctionId,
		"legacyId":      merchantRefundReq.LegacyId,
		"refundAmount":  merchantRefundReq.RefundPrice,
		"refundNo":      merchantRefundReq.RefundNo,
		"respCode":      refundRespInfo.ReturnCode,
		"respMsg":       refundRespInfo.ReturnMsg,
	}
	if refundRespInfo.ReturnCode == StatusSuccess {
		respMap["refundTransactionDate"] = refundRespInfo.Info.RefundTransactionDate
		respMap["refundLegacyId"] = refundRespInfo.Info.RefundTransactionId
		respMap["status"] = "REFUND"
	} else {
		respMap["status"] = "FAIL"
	}

	fmt.Println("Refund end")
	merchant.RespondWithJSON(w, respMap)

}

// appendHeader 添加 LINE Pay 需要的認證頭
func appendHeader(req *http.Request, body *string, path, channelID, channelSecret string) error {
	nonce := merchant.GenerateNonce()
	timestamp := fmt.Sprintf("%d", time.Now().Unix())

	// 組合簽名字符串
	signatureBase := fmt.Sprintf("%s%s%s%s", channelSecret, path, *body, nonce)

	// 計算 HMAC-SHA256
	mac := hmac.New(sha256.New, []byte(channelSecret))
	mac.Write([]byte(signatureBase))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	// 設置請求頭
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-LINE-ChannelId", channelID)
	req.Header.Set("X-LINE-Authorization-Nonce", nonce)
	req.Header.Set("X-LINE-Authorization", signature)
	req.Header.Set("X-LINE-Authorization-Timestamp", timestamp)

	return nil
}

// NewAuthService 創建新的授權服務
func NewAuthService(baseURL string) AuthService {
	return &DefaultAuthService{
		client: &http.Client{
			Timeout:   10 * time.Second,
			Transport: nil,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				panic("TODO")
			},
			Jar: nil,
		},
		baseURL: baseURL,
	}
}

// AuthService 定義了授權相關功能的接口
type AuthService interface {
	Authorize(request *LinePayAuthRequest, channelID, channelSecret string) (*LinePayAuthResponse, error)
}

// DefaultAuthService 是 AuthService 的默認實現
type DefaultAuthService struct {
	client  *http.Client
	baseURL string
}

// Authorize 處理 LINE Pay 授權請求
func (s *DefaultAuthService) Authorize(request *LinePayAuthRequest, channelID, channelSecret string) (*LinePayAuthResponse, error) {
	ctx := context.Background()
	return s.AuthorizeWithContext(ctx, request, channelID, channelSecret)
}

// AuthorizeWithContext 帶有上下文的授權請求處理
func (s *DefaultAuthService) AuthorizeWithContext(ctx context.Context, request *LinePayAuthRequest, channelID, channelSecret string) (*LinePayAuthResponse, error) {
	log.Println("Starting LINE Pay authorization")

	// 1. 序列化請求體
	reqBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 2. 創建 HTTP 請求
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+authPath, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 3. 添加請求頭
	reqStr := string(reqBody)
	if err := appendHeader(req, &reqStr, authPath, channelID, channelSecret); err != nil {
		return nil, fmt.Errorf("failed to append headers: %w", err)
	}

	// 4. 發送請求
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// 5. 檢查 HTTP 狀態碼
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	// 6. 解析回應
	var linePayResp LinePayAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&linePayResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &linePayResp, nil
}

func newAuthReq(orderId *string, amount *int, currency *string) LinePayAuthRequest {
	// LinePay 付款結構，忽略商品結構
	return LinePayAuthRequest{
		Currency: *currency,
		OrderID:  *orderId,
		Amount:   *amount,
		Packages: []Package{
			{
				ID:     "123",
				Amount: *amount,
				Products: []Product{
					{
						ID:       "Demo1",
						Name:     "DemoProduct",
						Quantity: 1,
						Price:    *amount,
						ImageURL: "",
					},
				},
			},
		},
		RedirectUrls: RedirectUrls{
			ConfirmUrl: callbackUrl,
			CancelUrl:  failCallbackUrl,
		},
	}
}

func auth(request *LinePayAuthRequest, chanId *string, chanSec *string) (*LinePayAuthResponse, error) {
	service := NewAuthService(domain)
	return service.Authorize(request, *chanId, *chanSec)
}

// 回傳給特店的Callback
func NewMerchantCallbackResp(transactionId, legacyId, respCode, respMsg string, totalFee int) *merchant.MerchantCallbackResp {
	dto := &merchant.MerchantCallbackResp{
		TransactionId: transactionId,
		LegacyId:      legacyId,
		RespCode:      respCode,
		RespMsg:       respMsg,
	}

	// 如果成功，添加额外字段
	if respCode == StatusSuccess {
		dto.TotalFee = totalFee
		dto.PaidTime = time.Now().Format("2006-01-02 15:04:05")
		dto.Status = StatusPaidSuccess
	} else {
		dto.Status = "FAIL"
	}

	return dto
}
