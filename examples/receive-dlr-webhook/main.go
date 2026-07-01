// Receives delivery-status (DLR) and reply webhooks, verifying the signature.
//
//	SMSGO_WEBHOOK_SECRET=whsec_... go run ./examples/receive-dlr-webhook
//	# then expose it (ngrok / Cloudflare Tunnel) and set the URL via configure-webhook.
package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/sms-go/smsgo-sdk-go"
)

func main() {
	secret := os.Getenv("SMSGO_WEBHOOK_SECRET")
	if secret == "" {
		log.Fatal("set SMSGO_WEBHOOK_SECRET (from configure-webhook / the panel)")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "4000"
	}

	http.HandleFunc("/webhooks/smsgo", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		// Read the RAW body — the signature is over these exact bytes.
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "read error", http.StatusBadRequest)
			return
		}
		if !smsgo.VerifyWebhookSignature(body, r.Header.Get("X-SMSGo-Signature"), secret) {
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}

		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		log.Printf("webhook: %v", payload)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	})

	log.Printf("listening for webhooks on http://localhost:%s/webhooks/smsgo", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
