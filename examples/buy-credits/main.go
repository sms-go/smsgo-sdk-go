// Buys credits off-session and configures automatic recharge.
//
//	SMSGO_KEY=yourkey go run ./examples/buy-credits
//
// Idempotency: each Purchase creates a new charge. On timeout, query
// Billing.Invoices before retrying — do NOT blindly retry.
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

	client := smsgo.New(smsgo.Options{APIKey: key})
	ctx := context.Background()

	plans, err := client.Billing.Plans(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%d plans available\n", len(plans))

	cards, err := client.Billing.Cards(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if len(cards) == 0 {
		log.Fatal("no saved card — add one in the panel first")
	}

	receipt, err := client.Billing.Purchase(ctx, smsgo.PurchaseParams{Quantity: 5000})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("purchase %s: R$ %.2f for %d credits (invoice %s)\n",
		receipt.Status, receipt.Total, receipt.Quantity, receipt.InvoiceUUID)

	// Keep the balance topped up automatically.
	enabled := true
	threshold := 5.0
	planQty := 5000
	cardID := cards[0].ID
	alertEnabled := true
	alertThreshold := 15.0

	cfg, err := client.SetAutoRecharge(ctx, smsgo.AutoRechargeUpdate{
		Enabled:        &enabled,
		Threshold:      &threshold,
		PlanQuantity:   &planQty,
		CardID:         &cardID,
		AlertEnabled:   &alertEnabled,
		AlertThreshold: &alertThreshold,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("auto-recharge enabled=%v threshold=R$%.2f\n", cfg.Enabled, cfg.Threshold)
}
