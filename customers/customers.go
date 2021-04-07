package customers

import (
	"fmt"

	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/customer"
)

// Address to a customer structure
type Address struct {
	City       string `json:"city"`
	Country    string `json:"country"`
	PostalCode string `json:"postalCode"`
	Province   string `json:"province"`
	Street     string `json:"street"`
}

// Customer type to a customer
type Customer struct {
	Address   Address `json:"address"`
	Company   string  `json:"company"`
	Email     string  `json:"email"`
	FirstName string  `json:"firstName"`
	LastName  string  `json:"lastName"`
	Lgpd      bool    `json:"lgpd"`
	NifCif    string  `json:"nifCif"`
	Phone     string  `json:"phone"`
}

// Create a new customer in Stripe BBDD
func Create(newCustomer Customer) (*stripe.Customer, error) {
	fmt.Println("\n🔵 [INFO] Creating new customer...")
	fmt.Println()

	name := newCustomer.FirstName + " " + newCustomer.LastName

	addressParams := &stripe.AddressParams{
		Line1:      stripe.String(newCustomer.Address.Street),
		PostalCode: stripe.String(newCustomer.Address.PostalCode),
		State:      stripe.String(newCustomer.Address.Province),
		City:       stripe.String(newCustomer.Address.City),
		Country:    stripe.String(newCustomer.Address.Country),
	}

	shipping := &stripe.CustomerShippingDetailsParams{
		Address: addressParams,
		Name:    stripe.String(name),
		Phone:   stripe.String(newCustomer.Phone),
	}

	params := &stripe.CustomerParams{
		Name:     stripe.String(name),
		Email:    stripe.String(newCustomer.Email),
		Phone:    stripe.String(newCustomer.Phone),
		Address:  addressParams,
		Shipping: shipping,
	}

	metadata := map[string]string{
		"nifCif":  newCustomer.NifCif,
		"company": newCustomer.Company,
	}

	if newCustomer.Lgpd {
		metadata["lgpd"] = "true"
	}

	for key, value := range metadata {
		params.AddMetadata(key, value)
	}

	return customer.New(params)
}
