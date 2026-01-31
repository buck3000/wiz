package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
)

//go:embed static
var staticFiles embed.FS

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	cfg := loadConfig()

	// Serve static files from the embedded filesystem.
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatalf("failed to create static sub-fs: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(staticFS)))
	mux.HandleFunc("POST /api/create-checkout-session", handleCreateCheckoutSession(cfg))
	mux.HandleFunc("POST /api/webhook", handleWebhook(cfg))

	addr := fmt.Sprintf(":%s", port)
	log.Printf("wiz-web listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

// config holds all required environment configuration.
type config struct {
	StripeSecretKey    string
	StripeWebhookSecret string
	StripePriceProID   string
	StripePriceTeamID  string
}

func loadConfig() config {
	mustEnv := func(key string) string {
		v := os.Getenv(key)
		if v == "" {
			log.Fatalf("required environment variable %s is not set", key)
		}
		return v
	}

	return config{
		StripeSecretKey:    mustEnv("STRIPE_SECRET_KEY"),
		StripeWebhookSecret: mustEnv("STRIPE_WEBHOOK_SECRET"),
		StripePriceProID:   mustEnv("STRIPE_PRO_PRICE_ID"),
		StripePriceTeamID:  mustEnv("STRIPE_TEAM_PRICE_ID"),
	}
}
