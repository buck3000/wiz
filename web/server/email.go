package main

import (
	"fmt"
	"log"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/customer"
)

// storeLicenseOnCustomer saves the license key in the Stripe customer's
// metadata. This makes the key:
//   - Visible in the Stripe Dashboard for support lookup
//   - Included in Stripe-generated receipt/invoice emails via metadata
//   - Retrievable if the customer loses their key
func storeLicenseOnCustomer(customerID, licenseKey, tier string) error {
	params := &stripe.CustomerParams{
		Params: stripe.Params{
			Metadata: map[string]string{
				"wiz_license_key":  licenseKey,
				"wiz_license_tier": tier,
			},
		},
		Description: stripe.String(fmt.Sprintf("Wiz %s subscriber", tier)),
	}

	_, err := customer.Update(customerID, params)
	if err != nil {
		return fmt.Errorf("update customer metadata: %w", err)
	}

	log.Printf("stored %s license key on customer %s", tier, customerID)
	return nil
}
