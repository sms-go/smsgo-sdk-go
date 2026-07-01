// Configures the outbound webhook (DLR + replies).
//
//	SMSGO_KEY=yourkey go run ./examples/configure-webhook https://yourapp.com/webhooks/smsgo
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
	if len(os.Args) < 2 {
		log.Fatal("usage: configure-webhook <https-url>")
	}
	url := os.Args[1]

	client := smsgo.New(smsgo.Options{APIKey: key})
	ctx := context.Background()

	cfg, err := client.SetWebhook(ctx, smsgo.WebhookUpdate{URL: &url})
	if err != nil {
		log.Fatal(err)
	}

	if cfg.URL != nil {
		fmt.Println("webhook set to:", *cfg.URL)
	}
	if cfg.Secret != nil {
		// Store this secret; use it to verify X-SMSGo-Signature on each callback.
		fmt.Println("signing secret:", *cfg.Secret)
	}
}
