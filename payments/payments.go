package payments

import (
	"fmt"

	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/paymentintent"

	"github.com/javierlopezdeancos/stipendivm/config"
	"github.com/javierlopezdeancos/stipendivm/inventory"
)

// IntentCreationRequest Intent creation request
type IntentCreationRequest struct {
	Currency string           `json:"currency"`
	Items    []inventory.Item `json:"items"`
}

// IntentShippingChangeRequest Intent shipping change request
type IntentShippingChangeRequest struct {
	Items          []inventory.Item         `json:"items"`
	ShippingOption inventory.ShippingOption `json:"shippingOption"`
}

// IntentCurrencyPaymentMethodsChangeRequest Intent currency payment methods change request
type IntentCurrencyPaymentMethodsChangeRequest struct {
	Currency       string   `json:"currency"`
	PaymentMethods []string `json:"payment_methods"`
}

// CreateIntent Create intent
func CreateIntent(r *IntentCreationRequest) (*stripe.PaymentIntent, error) {
	amount, err := inventory.CalculatePaymentAmount(r.Items)

	if err != nil {
		return nil, fmt.Errorf("payments: error computing payment amount: %v", err)
	}

	// build initial payment methods which should exclude currency specific ones
	initPaymentMethods := config.PaymentMethods()
	removeVal(initPaymentMethods, "au_becs_debit")

	params := &stripe.PaymentIntentParams{
		Amount:             stripe.Int64(amount),
		Currency:           stripe.String(r.Currency),
		PaymentMethodTypes: stripe.StringSlice(initPaymentMethods),
	}

	pi, err := paymentintent.New(params)

	if err != nil {
		return nil, fmt.Errorf("payments: error creating payment intent: %v", err)
	}

	return pi, nil
}

// helper function to remove a value from a slice
func removeVal(slice []string, value string) []string {
	for i, other := range slice {
		if other == value {
			return append(slice[:i], slice[i+1:]...)
		}
	}

	return slice
}

// RetrieveIntent Retrieve intent
func RetrieveIntent(paymentIntent string) (*stripe.PaymentIntent, error) {
	pi, err := paymentintent.Get(paymentIntent, nil)

	if err != nil {
		return nil, fmt.Errorf("payments: error fetching payment intent: %v", err)
	}

	return pi, nil
}

// ConfirmIntent Confirm intent
func ConfirmIntent(paymentIntent string, source *stripe.Source) error {
	pi, err := paymentintent.Get(paymentIntent, nil)

	if err != nil {
		return fmt.Errorf("payments: error fetching payment intent for confirmation: %v", err)
	}

	if pi.Status != "requires_payment_method" {
		return fmt.Errorf("payments: PaymentIntent already has a status of %s", pi.Status)
	}

	params := &stripe.PaymentIntentConfirmParams{
		PaymentMethod: stripe.String(source.ID),
	}

	_, err = paymentintent.Confirm(pi.ID, params)

	if err != nil {
		return fmt.Errorf("payments: error confirming PaymentIntent: %v", err)
	}

	return nil
}

// CancelIntent Cancel indent
func CancelIntent(paymentIntent string) error {
	_, err := paymentintent.Cancel(paymentIntent, nil)

	if err != nil {
		return fmt.Errorf("payments: error canceling PaymentIntent: %v", err)
	}

	return nil
}

// UpdateShipping Update shipping
func UpdateShipping(paymentIntent string, r *IntentShippingChangeRequest) (*stripe.PaymentIntent, error) {
	amount, err := inventory.CalculatePaymentAmount(r.Items)

	if err != nil {
		return nil, fmt.Errorf("payments: error computing payment amount: %v", err)
	}

	shippingCost, ok := inventory.GetShippingCost(r.ShippingOption.ID)

	if !ok {
		return nil, fmt.Errorf("payments: no cost found for shipping id %q", r.ShippingOption.ID)
	}

	amount += shippingCost

	params := &stripe.PaymentIntentParams{
		Amount: stripe.Int64(amount),
	}

	pi, err := paymentintent.Update(paymentIntent, params)

	if err != nil {
		return nil, fmt.Errorf("payments: error updating payment intent: %v", err)
	}

	return pi, nil
}

// UpdateCurrencyPaymentMethod Update payment currency
func UpdateCurrencyPaymentMethod(paymentIntent string, r *IntentCurrencyPaymentMethodsChangeRequest) (*stripe.PaymentIntent, error) {
	currency := r.Currency
	paymentMethods := r.PaymentMethods

	params := &stripe.PaymentIntentParams{
		Currency:           stripe.String(currency),
		PaymentMethodTypes: stripe.StringSlice(paymentMethods),
	}

	pi, err := paymentintent.Update(paymentIntent, params)

	if err != nil {
		return nil, fmt.Errorf("payments: error updating payment intent: %v", err)
	}

	return pi, nil
}
