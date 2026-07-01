// Sends a 6-digit OTP / 2FA code.
//
//	SMSGO_KEY=yourkey go run ./examples/send-otp +5511999990000
package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/sms-go/smsgo-sdk-go"
)

func main() {
	key := os.Getenv("SMSGO_KEY")
	if key == "" {
		log.Fatal("set SMSGO_KEY")
	}
	if len(os.Args) < 2 {
		log.Fatal("usage: send-otp <phone>")
	}
	phone := os.Args[1]

	n, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		log.Fatal(err)
	}
	code := fmt.Sprintf("%06d", n.Int64()+100000)

	client := smsgo.New(smsgo.Options{APIKey: key})

	res, err := client.Send(context.Background(), smsgo.SendParams{
		Phone:   phone,
		Message: fmt.Sprintf("Seu código SMSGo é %s. Válido por 5 minutos.", code),
	})
	if err != nil {
		log.Fatal(err)
	}
	// Store `code` (with a TTL) and compare it on verification.
	fmt.Printf("otp sent: id=%s code=%s\n", res.ID, code)
}
