package smsgo_test

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"

	"github.com/sms-go/smsgo-sdk-go"
)

// ExampleClient_Send sends a single SMS.
func ExampleClient_Send() {
	client := smsgo.New(smsgo.Options{APIKey: os.Getenv("SMSGO_KEY")})

	res, err := client.Send(context.Background(), smsgo.SendParams{
		Phone:   "+5511999990000",
		Message: "Olá do SMSGo",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(res.ID, res.Status)
}

// Example_otp sends a 6-digit one-time password.
func Example_otp() {
	client := smsgo.New(smsgo.Options{APIKey: os.Getenv("SMSGO_KEY")})

	n, _ := rand.Int(rand.Reader, big.NewInt(900000))
	code := fmt.Sprintf("%06d", n.Int64()+100000)

	_, err := client.Send(context.Background(), smsgo.SendParams{
		Phone:   "+5511999990000",
		Message: fmt.Sprintf("Seu código SMSGo é %s. Válido por 5 minutos.", code),
	})
	if err != nil {
		log.Fatal(err)
	}
	// Store `code` with a TTL and compare it on verification.
}

// ExampleClient_Get queries the status of a send.
func ExampleClient_Get() {
	client := smsgo.New(smsgo.Options{APIKey: os.Getenv("SMSGO_KEY")})
	ctx := context.Background()

	detail, err := client.Get(ctx, "a1b2c3-...")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%d/%d delivered, done=%v\n",
		detail.Summary.Delivered, detail.Summary.Total, detail.Summary.Done)

	// Track a large send by bucket, paginated:
	failed, err := client.GetNumbers(ctx, detail.ID, smsgo.NumbersParams{Status: "failed", Page: 1})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("failed rows:", len(failed.Data))
}

// ExampleClient_GetBalance reads the balance and the SMS-type catalog.
func ExampleClient_GetBalance() {
	client := smsgo.New(smsgo.Options{APIKey: os.Getenv("SMSGO_KEY")})
	ctx := context.Background()

	bal, err := client.GetBalance(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%.2f %s\n", bal.Balance, bal.Currency)

	types, err := client.GetSMSTypes(ctx)
	if err != nil {
		log.Fatal(err)
	}
	for _, t := range types {
		fmt.Printf("#%d %s R$%.2f\n", t.ID, t.Name, t.Price)
	}
}

// ExampleBillingResource_Purchase buys credits off-session.
func ExampleBillingResource_Purchase() {
	client := smsgo.New(smsgo.Options{APIKey: os.Getenv("SMSGO_KEY")})
	ctx := context.Background()

	receipt, err := client.Billing.Purchase(ctx, smsgo.PurchaseParams{Quantity: 5000})
	if err != nil {
		log.Fatal(err)
	}
	// "succeeded" already credited the balance; "processing" confirms via webhook.
	fmt.Println(receipt.Status, receipt.InvoiceUUID)
}

// ExampleClient_SetWebhook configures the outbound webhook.
func ExampleClient_SetWebhook() {
	client := smsgo.New(smsgo.Options{APIKey: os.Getenv("SMSGO_KEY")})
	ctx := context.Background()

	url := "https://yourapp.com/webhooks/smsgo"
	cfg, err := client.SetWebhook(ctx, smsgo.WebhookUpdate{URL: &url})
	if err != nil {
		log.Fatal(err)
	}
	if cfg.Secret != nil {
		fmt.Println("store this secret:", *cfg.Secret)
	}
}

// ExampleVerifyWebhookSignature verifies an incoming DLR webhook.
func ExampleVerifyWebhookSignature() {
	secret := os.Getenv("SMSGO_WEBHOOK_SECRET")

	http.HandleFunc("/webhooks/smsgo", func(w http.ResponseWriter, r *http.Request) {
		// Read the RAW body — the signature is over these exact bytes.
		body, _ := io.ReadAll(r.Body)

		if !smsgo.VerifyWebhookSignature(body, r.Header.Get("X-SMSGo-Signature"), secret) {
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	})
}
