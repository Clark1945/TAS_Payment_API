package merchant

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

var (
	ContentTypeJSON = "application/json"
)

// 統一的回應輔助函數
func RespondWithError(w http.ResponseWriter, message string, statusCode int) {
	log.Printf("Error: %s", message)
	http.Error(w, message, statusCode)
}

// 回傳方法
func RespondWithJSON(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", ContentTypeJSON)
	return json.NewEncoder(w).Encode(data)
}

// generateNonce 生成隨機字符串作為 nonce
func GenerateNonce() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("%x", b)
}

func LogWithResponse(input interface{}) {
	respNewByte, err := json.MarshalIndent(input, "", " ")
	if err != nil {
		log.Println(err)
	}
	fmt.Println(string(respNewByte))
}
