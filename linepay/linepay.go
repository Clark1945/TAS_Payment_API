package linepay

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

// LinePay SandBox ID:test_202503193091@line.pay
// PW: nCb72ySCQ&
// 通路密鑰:c63b2a81aadfeddc3ea33e572e29942a
// 通路ID:2007090890
// 商店ID:test_202503193091
// 統編:24941093
const (
	domain          string = "https://sandbox-api-pay.line.me"
	authPath        string = "/v3/payments/request"
	callbackPath    string = "/v3/payments/%s/confirm"
	queryPath       string = "/v3/payments?transactionId=%s&orderId=%s"
	refundPath      string = "/v3/payments/%s/refund"
	merId                  = "test_202503193091"
	callbackUrl     string = "http://localhost:8080/callback"
	failCallbackUrl string = "http://localhost:8080/failCallback"
)

var tempMap = make(map[string]map[string]string)

// API Handler: 對外提供 `/auth` 端點
func AuthHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	reqByte, err := io.ReadAll(req.Body)
	if err != nil {
		log.Println(err)
		http.Error(w, "Request Body Error", http.StatusBadRequest)
		return
	}
	var merchantReq MerchantAuthReq
	err = json.Unmarshal(reqByte, &merchantReq)
	if err != nil {
		log.Println(err)
		http.Error(w, "Request Body Error", http.StatusBadRequest)
		return
	}

	chanId := merchantReq.PaymentConfig["channelId"]
	chanSec := merchantReq.PaymentConfig["channelSec"]
	linePayReq := newAuthReq(&merchantReq.TransactionId, &merchantReq.TotalPrice, &merchantReq.Currency)

	linePayResp, err := auth(&linePayReq, &chanId, &chanSec)
	if err != nil {
		http.Error(w, fmt.Sprintf("Auth failed: %v", err), http.StatusInternalServerError)
		return
	}

	var merResp MerchantAuthResp
	merResp.Correlationid = merchantReq.Correlationid
	merResp.TransactionId = merchantReq.TransactionId
	merResp.CallTime = time.Now().Format("2006-01-02 15:04:05")
	merResp.RespCode = linePayResp.ReturnCode
	merResp.RespMsg = linePayResp.ReturnMessage
	if linePayResp.ReturnCode == "0000" { // 交易成功
		merResp.LegacyId = strconv.Itoa(linePayResp.Info.TxId)
		merResp.PaymentUrl = linePayResp.Info.PaymentUrl.Web
		merResp.Currency = merchantReq.Currency
		merResp.Status = "AUTH"

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

	respByte, err := json.Marshal(merResp)
	if err != nil {
		http.Error(w, "Response Error", 500)
		return
	}
	var respMap map[string]interface{}
	err = json.Unmarshal(respByte, &respMap)
	if err != nil {
		http.Error(w, "Response Error", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(respMap)
}

// TODO 檢查 請求

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

func auth(request *LinePayAuthRequest, chanId *string, chanSec *string) (*LinePayResp, error) {
	// auth API == Payment
	fmt.Println("auth started")

	// 處理規格 我覺得有點多此一舉
	reqBody, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", domain+authPath, bytes.NewReader(reqBody))
	if err != nil { // reqBody Error!
		log.Println(err)
		return nil, err
	}

	reqStr := string(reqBody)
	err = appendHeader(req, &reqStr, authPath, chanId, chanSec)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// 發送請求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer resp.Body.Close() // Delay disconnect the connection.

	// 解析Response
	respByte, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	var linePayResp LinePayResp
	err = json.Unmarshal(respByte, &linePayResp)
	if err != nil {
		log.Println("解析失敗:", err)
	}
	prettyPrint, err := json.MarshalIndent(linePayResp, "", "  ")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	fmt.Println(string(prettyPrint))

	fmt.Println("auth end")
	return &linePayResp, nil
}

func appendHeader(req *http.Request, reqStr *string, path string, chanId *string, chanSec *string) error {
	nonce := fmt.Sprint(time.Now().UnixMilli())
	signature, err := signKey(*chanSec, fmt.Sprintf("%s%s%s%s", *chanSec, path, *reqStr, nonce))
	if err != nil {
		return err
	}
	req.Header.Add("X-LINE-ChannelId", *chanId)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-LINE-Authorization", signature)
	req.Header.Add("X-LINE-Authorization-Nonce", nonce)
	return nil
}

func signKey(clientKey, msg string) (string, error) {
	// Create an HMAC hasher using SHA-256 and the client key
	h := hmac.New(sha256.New, []byte(clientKey))

	// Write the message to the hasher
	_, err := h.Write([]byte(msg))
	if err != nil {
		return "", err
	}

	// Compute the HMAC digest and encode it as base64
	return base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}

func CallbackHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("callback start")
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
	reqBody, err := json.Marshal(reqMap)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", domain+fmt.Sprintf(callbackPath, legacyId), bytes.NewReader(reqBody))
	if err != nil { // reqBody Error!
		log.Println(err)
		return
	}

	reqDummp := string(reqBody)
	appendHeader(req, &reqDummp, fmt.Sprintf(callbackPath, legacyId), &chanId, &chanSec)

	// 發送請求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close() // Delay disconnect the connection.

	// 解析Response
	respByte, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}

	var callbackResp CallbackResp
	err = json.Unmarshal(respByte, &callbackResp)
	if err != nil {
		log.Println(err)
	}
	respMap := map[string]interface{}{
		"transactionId": transactionId,
		"legacyId":      legacyId,
		"respCode":      callbackResp.ReturnCode,
		"respMsg":       callbackResp.ReturnMessage,
	}
	if callbackResp.ReturnCode == "0000" {
		respMap["totalFee"] = totalFee
		respMap["paidTime"] = time.Now().Format("2006-01-02 15:04:05")
		respMap["status"] = "PAID_SUCCESS"
	} else {
		respMap["status"] = "FAIL"
	}

	prettyPrint, err := json.MarshalIndent(callbackResp, "", "  ")
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println(string(prettyPrint))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(respMap)
}

// func QueryHandler(w http.ResponseWriter, req *http.Request) {
// 	fmt.Println("Query started")

// 	reqByte, err := io.ReadAll(req.Body)
// 	if err != nil {
// 		log.Println(err)
// 		return
// 	}

// 	var reqMap map[string]interface{}
// 	err = json.Unmarshal(reqByte, &reqMap)
// 	if err != nil {
// 		log.Println(err)
// 		return
// 	}

// 	fullQueryPath := fmt.Sprintf(domain+queryPath, reqMap["transactionId"], reqMap["orderId"])
// 	request, err := http.NewRequest("GET", fullQueryPath, nil)
// 	if err != nil {
// 		log.Println(err)
// 		return
// 	}

// 	slice := strings.Split(fmt.Sprintf(queryPath, reqMap["transactionId"], reqMap["orderId"]), "?")
// 	appendHeader(request, &slice[1], slice[0])

// 	resp, err := http.DefaultClient.Do(request)
// 	if err != nil {
// 		log.Println("連線異常, ", err)
// 		return
// 	}

// 	respByte, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		log.Println(err)
// 	}

// 	var respMap map[string]interface{}
// 	err = json.Unmarshal(respByte, &respMap)
// 	if err != nil {
// 		log.Println(err)
// 	}

// 	respNewByte, err := json.MarshalIndent(respMap, "", " ")
// 	if err != nil {
// 		log.Println(err)
// 	}
// 	fmt.Println(string(respNewByte))
// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(respMap)
// 	fmt.Println("Query end")
// }

// func RefundHandler(w http.ResponseWriter, req *http.Request) {
// 	fmt.Println("Refund started")

// 	reqByte, err := io.ReadAll(req.Body)
// 	if err != nil {
// 		log.Println(err)
// 		return
// 	}

// 	var reqMap map[string]interface{}
// 	err = json.Unmarshal(reqByte, &reqMap)
// 	if err != nil {
// 		log.Println(err)
// 		return
// 	}

// 	path := fmt.Sprintf(refundPath, "2025032102277666310")
// 	request, err := http.NewRequest("POST", domain+path, bytes.NewReader(reqByte))
// 	if err != nil {
// 		log.Println(err)
// 		return
// 	}

// 	reqDummp := string(reqByte)
// 	appendHeader(request, &reqDummp, path)

// 	resp, err := http.DefaultClient.Do(request)
// 	if err != nil {
// 		log.Println("連線異常, ", err)
// 		return
// 	}

// 	respByte, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		log.Println(err)
// 	}

// 	var respMap map[string]interface{}
// 	err = json.Unmarshal(respByte, &respMap)
// 	if err != nil {
// 		log.Println(err)
// 	}

// 	respNewByte, err := json.MarshalIndent(respMap, "", " ")
// 	if err != nil {
// 		log.Println(err)
// 	}
// 	fmt.Println(string(respNewByte))
// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(respMap)
// 	fmt.Println("Refund end")
// }
