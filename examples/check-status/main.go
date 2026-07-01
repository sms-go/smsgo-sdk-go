// Sends a bulk batch and then queries its delivery status.
//
//	SMSGO_KEY=yourkey go run ./examples/check-status
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

	sent, err := client.SendBulk(ctx, smsgo.SendBulkParams{
		Messages: []smsgo.BulkMessage{
			{Phone: "+5511999990000", Message: "Oi, Ana!"},
			{Phone: "+5521988887777", Message: "Oi, Bruno!"},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("batch %s (%d messages)\n", sent.ID, sent.Quantity)

	detail, err := client.Get(ctx, sent.ID)
	if err != nil {
		log.Fatal(err)
	}
	s := detail.Summary
	fmt.Printf("delivered=%d failed=%d inProgress=%d done=%v\n",
		s.Delivered, s.Failed, s.InProgress, s.Done)

	failed, err := client.GetNumbers(ctx, sent.ID, smsgo.NumbersParams{Status: "failed", Page: 1})
	if err != nil {
		log.Fatal(err)
	}
	for _, n := range failed.Data {
		fmt.Printf("failed: %s (%s)\n", n.Phone, n.Status)
	}
}
