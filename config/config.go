package config

import (
	"os"
	"strings"

	"github.com/javierlopezdeancos/stipendivm/inventory"
)

// Configuration type to our stripe integration
type Configuration struct {
	StripePublishableKey string                     `json:"stripePublishableKey"`
	StripeCountry        string                     `json:"stripeCountry"`
	Country              string                     `json:"country"`
	Currency             string                     `json:"currency"`
	PaymentMethods       []string                   `json:"paymentMethods"`
	ShippingOptions      []inventory.ShippingOption `json:"shippingOptions"`
}

// GetPaymentMethods get payments methods selected to the stripe integration
func GetPaymentMethods() []string {
	paymentMethodsString := os.Getenv("PAYMENT_METHODS")

	if paymentMethodsString == "" {
		return []string{"card"}
	}

	return strings.Split(paymentMethodsString, ", ")
}

// Default get default values to stripe integration
func Default() Configuration {
	stripeCountry := os.Getenv("STRIPE_ACCOUNT_COUNTRY")
	country := os.Getenv("COUNTRY")
	currency := os.Getenv("CURRENCY")

	if stripeCountry == "" {
		stripeCountry = "SP"
	}

	if country == "" {
		country = "SP"
	}

	if currency == "" {
		currency = "eur"
	}

	c := Configuration{
		StripePublishableKey: os.Getenv("STRIPE_PUBLISHABLE_KEY"),
		StripeCountry:        stripeCountry,
		Country:              country,
		Currency:             currency,
		PaymentMethods:       GetPaymentMethods(),
		ShippingOptions:      inventory.ShippingOptions(),
	}

	return c
}
