package config

import (
	"os"
	"strings"
)

// Environments types map
var Environments = map[string]string{
	"development": "dev",
	"production":  "prod",
}

// Environment execution
var Environment string

// PublicDirectory in server
var PublicDirectory string

// Configuration type to our stripe integration
type Configuration struct {
	StripePublishableKey string           `json:"stripePublishableKey"`
	StripeCountry        string           `json:"stripeCountry"`
	Country              string           `json:"country"`
	Currency             string           `json:"currency"`
	PaymentMethods       []string         `json:"paymentMethods"`
	ShippingOptions      []ShippingOption `json:"shippingOptions"`
}

// GetPaymentMethods get payments methods selected to the stripe integration
func GetPaymentMethods() []string {
	paymentMethodsString := os.Getenv("PAYMENT_METHODS")

	if paymentMethodsString == "" {
		return []string{"card"}
	}

	return strings.Split(paymentMethodsString, ", ")
}

// ShippingOption Shipping option
type ShippingOption struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	Detail string `json:"detail"`
	Amount int64  `json:"amount"`
}

// GetShippingOptions Shipping options
func GetShippingOptions() []ShippingOption {
	return []ShippingOption{
		{
			ID:     "free",
			Label:  "Free Shipping",
			Detail: "Delivery within 5 days",
			Amount: 0,
		},
		{
			ID:     "express",
			Label:  "Express Shipping",
			Detail: "Next day delivery",
			Amount: 500,
		},
	}
}

// GetShippingCost Get shipping cost
func GetShippingCost(optionID string) (int64, bool) {
	for _, option := range GetShippingOptions() {
		if option.ID == optionID {
			return option.Amount, true
		}
	}

	return 0, false
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
		ShippingOptions:      GetShippingOptions(),
	}

	return c
}
