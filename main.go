package main

import (
	"TAS_payment/linepay"
	"fmt"
)

func main() {
	fmt.Println("This is my TAS_Payment Init!")
	linepay.Auth()
}
