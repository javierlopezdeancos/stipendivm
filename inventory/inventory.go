package inventory

import (
	"fmt"

	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/price"
	"github.com/stripe/stripe-go/v72/product"
	"github.com/stripe/stripe-go/v72/sku"

	"github.com/javierlopezdeancos/stipendivm/config"
)

// Item Intent payment item
type Item struct {
	Parent   string `json:"parent"`
	Quantity int64  `json:"quantity"`
}

// ListWines Wines list
func ListWines() ([]*stripe.Product, error) {
	wines := []*stripe.Product{}

	params := &stripe.ProductListParams{}
	params.Filters.AddFilter("limit", "", "3")
	params.Filters.AddFilter("active", "", "true")

	i := product.List(params)

	for i.Next() {
		wines = append(wines, i.Product())
	}

	err := i.Err()

	if err != nil {
		return nil, fmt.Errorf("inventory: error listing products: %v", err)
	}

	if config.Environment == config.Environments["development"] {
		winesInDevelopmentEnvironment := []*stripe.Product{}

		for _, w := range wines {
			if w.Livemode {
				continue
			}

			winesInDevelopmentEnvironment = append(winesInDevelopmentEnvironment, w)
		}

		return winesInDevelopmentEnvironment, nil
	} else if config.Environment == config.Environments["production"] {
		winesInProductionEnvironment := []*stripe.Product{}

		for _, w := range wines {
			if w.Livemode {
				winesInProductionEnvironment = append(winesInProductionEnvironment, w)
			}
		}

		return winesInProductionEnvironment, nil
	}

	return wines, nil
}

// RetrieveWine Retrieve wine from wine list
func RetrieveWine(wineID string) (*stripe.Product, error) {
	return product.Get(wineID, nil)
}

// UpdateWineStock Update wine stock
func UpdateWineStock(wineID string, stock string) (*stripe.Product, error) {
	wine, err := RetrieveWine(wineID)

	if err != nil {
		return nil, err
	}

	wine.Metadata["quantity"] = stock

	return wine, nil
}

// ListSKUs SKUs list
func ListSKUs(productID string) ([]*stripe.SKU, error) {
	skus := []*stripe.SKU{}

	params := &stripe.SKUListParams{}
	params.Filters.AddFilter("limit", "", "1")

	i := sku.List(params)

	for i.Next() {
		skus = append(skus, i.SKU())
	}

	err := i.Err()

	if err != nil {
		return nil, fmt.Errorf("inventory: error listing SKUs: %v", err)
	}

	return skus, nil

}

// CalculatePaymentAmount Calc payment amount
func CalculatePaymentAmount(items []Item) (int64, error) {
	total := int64(0)

	for _, item := range items {
		prices, err := ListPrices(item.Parent)

		if err != nil {
			return 0, fmt.Errorf("inventory: error getting SKU for price: %v", err)
		}

		total += prices[0].UnitAmount * item.Quantity
	}

	return total, nil
}

// ListPrices Prices list
func ListPrices(args ...string) ([]*stripe.Price, error) {
	prices := []*stripe.Price{}

	params := &stripe.PriceListParams{}
	params.Filters.AddFilter("limit", "", "3")

	if len(args) == 1 {
		wineID := args[0]
		params.Product = stripe.String(wineID)
	}

	i := price.List(params)

	for i.Next() {
		prices = append(prices, i.Price())
	}

	err := i.Err()

	if err != nil {
		return nil, fmt.Errorf("inventory: error listing products: %v", err)
	}

	return prices, nil
}

// RetrievePrice Retrieve wine from wine list
func RetrievePrice(priceID string) (*stripe.Price, error) {
	return price.Get(priceID, nil)
}
