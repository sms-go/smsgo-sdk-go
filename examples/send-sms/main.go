// Sends a single SMS.
//
//	SMSGO_KEY=yourkey go run ./examples/send-sms +5511999990000 "Olá do SMSGo"
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/SMSFy/smsgo-sdk-go"
)

func main() {
	key := os.Getenv("SMSGO_KEY")
	if key == "" {
		log.Fatal("set SMSGO_KEY")
	}
	if len(os.Args) < 3 {
		log.Fatal("usage: send-sms <phone> <message>")
	}
	phone, message := os.Args[1], os.Args[2]

	client := smsgo.New(smsgo.Options{APIKey: key})

	res, err := client.Send(context.Background(), smsgo.SendParams{
		Phone:   phone,
		Message: message,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("sent: id=%s status=%s quantity=%d\n", res.ID, res.Status, res.Quantity)
}
