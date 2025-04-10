# TAJ_Payment_API

# 後端SideProject的一部分，預計串接：
1. LinePay API
2. PayPal API
3. 軟銀 API

功能包含但不限於：
* 授權(付款)
* 取消授權
* 請款
* 退款
* 查詢

TAS == Taiwan + America + Japan

TechStack:
* Golang
* API Design
* Asynchronous 
* Message Queue
* Redis Cache
* Test

### LinePay Payment Example
---
#### Request
```
{
    "transactionId" : "test126",
    "correlationid" : "track001",
	"totalPrice" : 100,
	"currency" : "TWD",
    "paymentConfig": {
        "channelId":"2007090890",
        "channelSec":"c63b2a81aadfeddc3ea33e572e29942a"
    },
    "callbackUrl":"https://www.google.com.tw/index.html",
    "notifyUrl":"https://www.google.com.tw/index.html"
  }
```
#### Response(Success)
```
{
    "callTime": "2025-03-23 13:51:13",
    "correlationid": "track001",
    "currency": "TWD",
    "legacyId": "2025032302277946610",
    "paymentUrl": "https://sandbox-web-pay.line.me/web/payment/wait?transactionReserveId=cFI3NUdpajF5TmRZWGNha2tBVU1LUVkycllQc1RoTFd5TkZMdW8wcUNNVkJEVlFJWDJIS2srdWxrc2kwc0QwaQ",
    "respCode": "0000",
    "respMsg": "Success.",
    "status": "AUTH",
    "transactionId": "test126"
}
```
#### Response(Fail)
```
{
    "transactionId": "test127",
    "correlationid": "track001",
    "callTime": "2025-03-30 09:19:16",
    "status": "AUTH_FAIL",
    "respCode": "1172",
    "respMsg": "Existing same orderId."
}
```
#### Asynchronous Callback (Response only)
```
{
  "legacyId": "2025032302277946610",
  "paidTime": "2025-03-23 13:51:40",
  "respCode": "0000",
  "respMsg": "Success.",
  "status": "PAID_SUCCESS",
  "totalFee": 100,
  "transactionId": "test126"
}
```

### Query Request
---
#### Request
```
{
    "legacyId": "2025032302277954210",
    "transactionId": "test127",
    "paymentConfig": {
        "channelId":"2007090890",
        "channelSec":"c63b2a81aadfeddc3ea33e572e29942a"
    }
}
```
#### Response(Success)
```
{
    "legacyId": "2025032302277954210",
    "payStatus": "CAPTURE",
    "respCode": "0000",
    "respMsg": "Success.",
    "transactionDate": "2025-03-23T11:02:14Z",
    "transactionId": "test127"
}
```
#### Response(Fail)
```
{
    "legacyId": "2025032302277954210",
    "payStatus": "CAPTURE",
    "respCode": "0000",
    "respMsg": "Success.",
    "transactionDate": "2025-03-23T11:02:14Z",
    "transactionId": "test127"
}
```

### Refund 
#### Request
```
{
    "transactionId" : "test126",
    "correlationid" : "track001",
    "legacyId":"2025032302277946610",
    "refundNo":"1234",
    "refundPrice": 20,
    "refundDate": "2025-03-23 17:00:00",
    "paymentConfig": {
        "channelId":"2007090890",
        "channelSec":"c63b2a81aadfeddc3ea33e572e29942a"
    }
}
```
#### Response (Succuess)
```
{
    "legacyId": "2025032302277946610",
    "refundAmount": 20,
    "refundLegacyId": 2025032302277962000,
    "refundNo": "1234",
    "refundTransactionDate": "2025-03-23T13:10:10Z",
    "respCode": "0000",
    "respMsg": "Success.",
    "status": "REFUND",
    "transactionId": "test126"
}
```
#### Response (Fail)
```
{
    "legacyId": "2025032302277946610",
    "refundAmount": 20,
    "refundLegacyId": 2025032302277962000,
    "refundNo": "1234",
    "refundTransactionDate": "2025-03-23T13:10:10Z",
    "respCode": "0000",
    "respMsg": "Success.",
    "status": "REFUND",
    "transactionId": "test126"
}
```