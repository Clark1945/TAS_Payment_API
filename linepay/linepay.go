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
	domain       string = "https://sandbox-api-pay.line.me"
	authPath     string = "/v3/payments/request"
	callbackPath string = "/v3/payments/%s/confirm"
	queryPath    string = "/v3/payments?transactionId=%s&orderId=%s"
	refundPath   string = "/v3/payments/%s/refund"
	chanSec             = "c63b2a81aadfeddc3ea33e572e29942a"
	chanId              = "2007090890"
	merId               = "test_202503193091"
)

// API Handler: 對外提供 `/auth` 端點
func AuthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var linePayRequest LinePayRequest
	err := json.NewDecoder(r.Body).Decode(&linePayRequest)
	if err != nil {
		http.Error(w, "Request Body Error", http.StatusBadRequest)
		return
	}

	url, err := auth(linePayRequest)
	if err != nil {
		http.Error(w, fmt.Sprintf("Auth failed: %v", err), http.StatusInternalServerError)
		return
	}

	// 回應 JSON
	response := map[string]string{"payment_url": url}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func auth(request LinePayRequest) (string, error) {
	// auth API == Payment
	fmt.Println("auth started")

	// 處理規格
	reqBody, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", domain+authPath, bytes.NewReader(reqBody))
	if err != nil { // reqBody Error!
		log.Println(err)
		return "", err
	}
	genEncHeader(req, string(reqBody), authPath)

	// 發送請求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
		return "", err
	}
	defer resp.Body.Close() // Delay disconnect the connection.

	// 解析Response
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return "", err
	}
	var linePayResp LinePayResp
	err = json.Unmarshal(respData, &linePayResp)
	if err != nil {
		log.Println("解析失敗:", err)
	}
	prettyPrint, err := json.MarshalIndent(linePayResp, "", "  ")
	if err != nil {
		log.Println(err)
		return "", err
	}
	fmt.Println(string(prettyPrint))

	fmt.Println("auth end")
	return linePayResp.Info.PaymentUrl.Web, err
}

func genEncHeader(req *http.Request, reqStr string, path string) (*http.Request, error) {
	fmt.Println(reqStr)
	fmt.Println(path)
	fmt.Println(merId, ">", chanId, ">", chanSec) // TODO DEL

	nonce := fmt.Sprint(time.Now().UnixMilli())

	signature, err := signKey(chanSec, fmt.Sprintf("%s%s%s%s", chanSec, path, reqStr, nonce))
	if err != nil {
		return nil, err
	}

	fmt.Println(signature)

	req.Header.Add("X-LINE-ChannelId", chanId)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-LINE-Authorization", signature)
	req.Header.Add("X-LINE-Authorization-Nonce", nonce)
	return req, nil
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
	// 收到callback 應該對LinePay查詢
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	reqMap := map[string]interface{}{
		"amount":   100,
		"currency": "TWD",
	}

	// 處理規格
	reqBody, err := json.Marshal(reqMap)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", domain+fmt.Sprintf(callbackPath, "2025032102277666310"), bytes.NewReader(reqBody))
	if err != nil { // reqBody Error!
		log.Println(err)
		return
	}
	genEncHeader(req, string(reqBody), fmt.Sprintf(callbackPath, "2025032102277666310"))

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

	var respMap map[string]interface{}
	err = json.Unmarshal(respByte, &respMap)
	if err != nil {
		log.Println(err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(respMap)
}

func QueryHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Println("Query started")

	reqByte, err := io.ReadAll(req.Body)
	if err != nil {
		log.Println(err)
		return
	}

	var reqMap map[string]interface{}
	err = json.Unmarshal(reqByte, &reqMap)
	if err != nil {
		log.Println(err)
		return
	}

	fullQueryPath := fmt.Sprintf(domain+queryPath, reqMap["transactionId"], reqMap["orderId"])
	request, err := http.NewRequest("GET", fullQueryPath, nil)
	if err != nil {
		log.Println(err)
		return
	}

	slice := strings.Split(fmt.Sprintf(queryPath, reqMap["transactionId"], reqMap["orderId"]), "?")
	genEncHeader(request, slice[1], slice[0])

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Println("連線異常, ", err)
		return
	}

	respByte, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}

	var respMap map[string]interface{}
	err = json.Unmarshal(respByte, &respMap)
	if err != nil {
		log.Println(err)
	}

	respNewByte, err := json.MarshalIndent(respMap, "", " ")
	if err != nil {
		log.Println(err)
	}
	fmt.Println(string(respNewByte))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(respMap)
	fmt.Println("Query end")
}

func RefundHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Println("Refund started")

	reqByte, err := io.ReadAll(req.Body)
	if err != nil {
		log.Println(err)
		return
	}

	var reqMap map[string]interface{}
	err = json.Unmarshal(reqByte, &reqMap)
	if err != nil {
		log.Println(err)
		return
	}

	path := fmt.Sprintf(refundPath, "2025032102277666310")
	request, err := http.NewRequest("POST", domain+path, bytes.NewReader(reqByte))
	if err != nil {
		log.Println(err)
		return
	}

	genEncHeader(request, string(reqByte), path)

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Println("連線異常, ", err)
		return
	}

	respByte, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}

	var respMap map[string]interface{}
	err = json.Unmarshal(respByte, &respMap)
	if err != nil {
		log.Println(err)
	}

	respNewByte, err := json.MarshalIndent(respMap, "", " ")
	if err != nil {
		log.Println(err)
	}
	fmt.Println(string(respNewByte))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(respMap)
	fmt.Println("Refund end")
}
