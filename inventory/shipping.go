package inventory

// ShippingOption Shipping option
type ShippingOption struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	Detail string `json:"detail"`
	Amount int64  `json:"amount"`
}

// ShippingOptions Shipping options
func ShippingOptions() []ShippingOption {
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
	for _, option := range ShippingOptions() {
		if option.ID == optionID {
			return option.Amount, true
		}
	}

	return 0, false
}
