package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/buck3000/wiz/internal/license"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/checkout/session"
	"github.com/stripe/stripe-go/v82/webhook"
)

// tierForPriceID maps a Stripe price ID to a wiz license tier.
func tierForPriceID(cfg config, priceID string) (license.Tier, bool) {
	switch priceID {
	case cfg.StripePriceProID:
		return license.TierPro, true
	case cfg.StripePriceTeamID:
		return license.TierTeam, true
	default:
		return license.TierFree, false
	}
}

// handleCreateCheckoutSession creates a Stripe Checkout session and redirects
// the user to it. Expects a form POST with a "tier" field ("pro" or "team").
func handleCreateCheckoutSession(cfg config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stripe.Key = cfg.StripeSecretKey

		tier := r.FormValue("tier")
		var priceID string
		switch tier {
		case "pro":
			priceID = cfg.StripePriceProID
		case "team":
			priceID = cfg.StripePriceTeamID
		default:
			http.Error(w, "invalid tier", http.StatusBadRequest)
			return
		}

		params := &stripe.CheckoutSessionParams{
			Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
			LineItems: []*stripe.CheckoutSessionLineItemParams{
				{
					Price:    stripe.String(priceID),
					Quantity: stripe.Int64(1),
				},
			},
			SuccessURL: stripe.String(successURL(r)),
			CancelURL:  stripe.String(cancelURL(r)),
		}

		s, err := session.New(params)
		if err != nil {
			log.Printf("error creating checkout session: %v", err)
			http.Error(w, "failed to create checkout session", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, s.URL, http.StatusSeeOther)
	}
}

// handleWebhook processes incoming Stripe webhook events.
func handleWebhook(cfg config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const maxBodyBytes = 65536
		body, err := io.ReadAll(io.LimitReader(r.Body, maxBodyBytes))
		if err != nil {
			http.Error(w, "read error", http.StatusServiceUnavailable)
			return
		}

		event, err := webhook.ConstructEvent(body, r.Header.Get("Stripe-Signature"), cfg.StripeWebhookSecret)
		if err != nil {
			log.Printf("webhook signature verification failed: %v", err)
			http.Error(w, "invalid signature", http.StatusBadRequest)
			return
		}

		switch event.Type {
		case "checkout.session.completed":
			handleCheckoutCompleted(cfg, event)
		case "invoice.paid":
			handleInvoicePaid(cfg, event)
		case "customer.subscription.deleted":
			handleSubscriptionDeleted(event)
		default:
			log.Printf("unhandled event type: %s", event.Type)
		}

		w.WriteHeader(http.StatusOK)
	}
}

// handleCheckoutCompleted processes a completed checkout: generates a license
// key and stores it on the Stripe customer for delivery via receipt email.
func handleCheckoutCompleted(cfg config, event stripe.Event) {
	var sess stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &sess); err != nil {
		log.Printf("error parsing checkout session: %v", err)
		return
	}

	email := sess.CustomerDetails.Email
	if email == "" {
		log.Printf("checkout session %s has no customer email", sess.ID)
		return
	}

	// Determine the tier from the line items. The checkout session object in
	// the webhook payload doesn't always include line items expanded, so we
	// look up the subscription's price via the session metadata or price ID.
	tier := resolveTierFromSession(cfg, &sess)

	// Generate a license key valid for 1 year (renewed on invoice.paid).
	key := license.GenerateKey(email, tier, time.Now().AddDate(1, 0, 0))
	log.Printf("generated %s license for %s (checkout %s)", tier, email, sess.ID)

	// Store the license key on the Stripe customer for retrieval and email delivery.
	if sess.Customer != nil {
		if err := storeLicenseOnCustomer(sess.Customer.ID, key, tier.String()); err != nil {
			log.Printf("error storing license on customer %s: %v", sess.Customer.ID, err)
		}
	}
}

// handleInvoicePaid generates a fresh license key on subscription renewal.
func handleInvoicePaid(cfg config, event stripe.Event) {
	var inv struct {
		CustomerEmail string `json:"customer_email"`
		CustomerID    string `json:"customer"`
		Lines         struct {
			Data []struct {
				Price struct {
					ID string `json:"id"`
				} `json:"price"`
			} `json:"data"`
		} `json:"lines"`
	}
	if err := json.Unmarshal(event.Data.Raw, &inv); err != nil {
		log.Printf("error parsing invoice: %v", err)
		return
	}

	if inv.CustomerEmail == "" || len(inv.Lines.Data) == 0 {
		return
	}

	priceID := inv.Lines.Data[0].Price.ID
	tier, ok := tierForPriceID(cfg, priceID)
	if !ok {
		log.Printf("invoice for unknown price %s, skipping", priceID)
		return
	}

	key := license.GenerateKey(inv.CustomerEmail, tier, time.Now().AddDate(1, 0, 0))
	log.Printf("renewed %s license for %s", tier, inv.CustomerEmail)

	if inv.CustomerID != "" {
		if err := storeLicenseOnCustomer(inv.CustomerID, key, tier.String()); err != nil {
			log.Printf("error updating license on customer %s: %v", inv.CustomerID, err)
		}
	}
}

// handleSubscriptionDeleted logs cancellation. The license key has a built-in
// expiry so it will naturally stop working.
func handleSubscriptionDeleted(event stripe.Event) {
	var sub struct {
		ID       string `json:"id"`
		Customer string `json:"customer"`
	}
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		log.Printf("error parsing subscription deletion: %v", err)
		return
	}
	log.Printf("subscription %s cancelled for customer %s", sub.ID, sub.Customer)
}

// resolveTierFromSession determines the license tier from a checkout session.
func resolveTierFromSession(cfg config, sess *stripe.CheckoutSession) license.Tier {
	// Try expanded line items first.
	if sess.LineItems != nil {
		for _, li := range sess.LineItems.Data {
			if li.Price != nil {
				if t, ok := tierForPriceID(cfg, li.Price.ID); ok {
					return t
				}
			}
		}
	}

	// Fallback: retrieve the session with line items expanded.
	stripe.Key = cfg.StripeSecretKey
	params := &stripe.CheckoutSessionParams{}
	params.AddExpand("line_items")
	expanded, err := session.Get(sess.ID, params)
	if err != nil {
		log.Printf("error expanding checkout session %s: %v", sess.ID, err)
		return license.TierPro // safe default for paid checkout
	}
	if expanded.LineItems != nil {
		for _, li := range expanded.LineItems.Data {
			if li.Price != nil {
				if t, ok := tierForPriceID(cfg, li.Price.ID); ok {
					return t
				}
			}
		}
	}

	return license.TierPro
}

func successURL(r *http.Request) string {
	scheme := "https"
	if r.TLS == nil {
		scheme = "http"
	}
	return fmt.Sprintf("%s://%s/success.html", scheme, r.Host)
}

func cancelURL(r *http.Request) string {
	scheme := "https"
	if r.TLS == nil {
		scheme = "http"
	}
	return fmt.Sprintf("%s://%s/#pricing", scheme, r.Host)
}
