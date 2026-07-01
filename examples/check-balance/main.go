// Prints the account balance and the SMS-type catalog.
//
//	SMSGO_KEY=yourkey go run ./examples/check-balance
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/sms-go/smsgo-sdk-go"
)

func main() {
	key := os.Getenv("SMSGO_KEY")
	if key == "" {
		log.Fatal("set SMSGO_KEY")
	}

	client := smsgo.New(smsgo.Options{APIKey: key})
	ctx := context.Background()

	bal, err := client.GetBalance(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("balance: %.2f %s (%s)\n", bal.Balance, bal.Currency, bal.Company.Name)

	types, err := client.GetSMSTypes(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("SMS types (id -> smsTypeId):")
	for _, t := range types {
		price := t.Price
		if t.Sale != nil {
			price = *t.Sale
		}
		fmt.Printf("  #%d %s — R$ %.4f\n", t.ID, t.Name, price)
	}
}
