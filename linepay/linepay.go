package linepay

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
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
var domain string = "https://sandbox-api-pay.line.me"
var authPath string = "/v3/payments/request"

var reqBodyStr string = `{
	"amount" : 100,
	"currency" : "TWD",
	"orderId" : "MKSI_S_20180904_1000002",
	"packages" : [
	  {
		"id" : "1",
		"amount": 100,
		"products" : [
		  {
			"id" : "PEN-B-001",
			"name" : "Pen Brown",
			"imageUrl" : "https://upload.wikimedia.org/wikipedia/commons/2/2f/Google_2015_logo.svg",
			"quantity" : 2,
			"price" : 50
		  }
		]
	  }
	],
	"redirectUrls" : {
	  "confirmUrl" : "https://www.google.com.tw/index.html",
	  "cancelUrl" : "https://www.google.com.tw/index.html"
	}
  }` //TODO

func Auth() {
	// auth API == Payment
	fmt.Println("auth started")

	reqBody := strings.NewReader(reqBodyStr)

	req, err := http.NewRequest("POST", domain+authPath, reqBody)
	if err != nil { // reqBody Error!
		log.Println(err)
		return
	}
	genEncHeader(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close() // Delay disconnect the connection.

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}
	// var linePayResp linePayResp
	var responseDataStr string = string(respData)
	fmt.Println("responseStr = " + responseDataStr)

	fmt.Println("auth end")
}

func genEncHeader(req *http.Request) (*http.Request, error) {
	merId, chanId, chanSec := extractProps()
	fmt.Println(merId, ">", chanId, ">", chanSec) // TODO DEL
	nonce := fmt.Sprint(time.Now().UnixMilli())
	var signature string
	var err error
	if req.Method == "GET" {
		signature, err = signKey(chanSec, fmt.Sprintf("%s%s%s%s", chanSec, authPath, reqBodyStr, nonce))
		if err != nil {
			return nil, err
		}
	} else if req.Method == "POST" {
		signature, err = signKey(chanSec, fmt.Sprintf("%s%s%s%s", chanSec, authPath, reqBodyStr, nonce))
		if err != nil {
			return nil, err
		}
	}

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

type linePayResp struct {
}

func extractProps() (string, string, string) {
	// extract properties from request
	chanSec := "c63b2a81aadfeddc3ea33e572e29942a"
	chanId := "2007090890"
	merId := "test_202503193091"
	return merId, chanId, chanSec
}

func capture() {
	fmt.Println("capture started")

	fmt.Println("capture end")
}

func query() {
	fmt.Println("Query started")

	fmt.Println("Query end")
}

func refund() {
	fmt.Println("Refund started")

	fmt.Println("Refund end")
}
