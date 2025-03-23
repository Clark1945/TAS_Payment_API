package main

import (
	"TAS_payment/linepay"
	"fmt"
	"log"
	"net/http"
)

func main() {
	fmt.Println("This is my TAS_Payment Init!")
	http.HandleFunc("/auth", linepay.AuthHandler)
	http.HandleFunc("/callback", linepay.CallbackHandler)
	// http.HandleFunc("/query", linepay.QueryHandler)
	// http.HandleFunc("/refund", linepay.RefundHandler)

	fmt.Println("Server running at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

}
