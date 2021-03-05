package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/joho/godotenv"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
	"github.com/stripe/stripe-go/v72"

	"github.com/javierlopezdeancos/stipendivm/customers"
	"github.com/javierlopezdeancos/stipendivm/inventory"
	"github.com/javierlopezdeancos/stipendivm/payments"
)

func main() {
	rootDirectory := flag.String("root", "./", "Root directory of the Stipendivm server to Quantvm stripe payments")
	environment := flag.String("env", "dev", "Type of environment to start Stipendivm server")

	flag.Parse()

	if *rootDirectory == "" {
		*rootDirectory = "./"
	}

	if *environment == "" {
		*environment = "dev"
	}

	var err error

	if *environment == environments["development"] {
		err = godotenv.Load(path.Join(*rootDirectory, ".env.development"))
	} else if *environment == environments["production"] {
		err = godotenv.Load(path.Join(*rootDirectory, ".env"))
	}

	if err != nil {
		panic(fmt.Sprintf("error loading .env: %v", err))
	}

	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	if stripe.Key == "" {
		panic("STRIPE_SECRET_KEY must be in environment")
	}

	publicDirectory := path.Join(*rootDirectory, "public")
	server := getServer(publicDirectory, environment)
	server.Logger.Fatal(server.Start(":4567"))
}

type listing struct {
	Data interface{} `json:"data"`
}

// Types to different environments
var Types = map[string]string{
	"test":       "test",
	"production": "prod",
}

var environments = map[string]string{
	"development": "dev",
	"production":  "prod",
}

func getWines(c echo.Context) error {
	wines, err := inventory.ListWines()

	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, listing{wines})
}

func getWineSkus(c echo.Context) error {
	skus, err := inventory.ListSKUs(c.Param("wine_id"))

	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, listing{skus})
}

func getWine(c echo.Context) error {
	wine, err := inventory.RetrieveWine(c.Param("wine_id"))

	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	return c.JSON(http.StatusOK, wine)
}

func getWinePrice(c echo.Context) error {
	price, err := inventory.ListPrices(c.Param("wine_id"))

	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, price)
}

func getPrices(c echo.Context) error {
	prices, err := inventory.ListPrices()

	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, listing{prices})
}

func getPaymentIntent(c echo.Context) error {
	r := new(payments.IntentCreationRequest)
	err := c.Bind(r)

	if err != nil {
		return err
	}

	pi, err := payments.CreateIntent(r)

	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]*stripe.PaymentIntent{
		"paymentIntent": pi,
	})
}

func getPaymentIntentShippingChange(c echo.Context) error {
	r := new(payments.IntentShippingChangeRequest)
	err := c.Bind(r)

	if err != nil {
		return err
	}

	pi, err := payments.UpdateShipping(c.Param("id"), r)

	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]*stripe.PaymentIntent{
		"paymentIntent": pi,
	})
}

func getPaymentIntentStatus(c echo.Context) error {
	pi, err := payments.RetrieveIntent(c.Param("id"))

	if err != nil {
		return err
	}

	p := payments.PaymentIntentsStatus{
		PaymentIntent: payments.PaymentIntentsStatusData{
			Status: string(pi.Status),
		},
	}

	if pi.LastPaymentError != nil {
		// p.PaymentIntent.LastPaymentError = pi.LastPaymentError.Message
	}

	return c.JSON(http.StatusOK, p)
}

func updatePaymentIntentCurrency(c echo.Context) error {
	r := new(payments.IntentCurrencyPaymentMethodsChangeRequest)
	err := c.Bind(r)

	if err != nil {
		return err
	}

	pi, err := payments.UpdateCurrencyPaymentMethod(c.Param("id"), r)

	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]*stripe.PaymentIntent{
		"paymentIntent": pi,
	})
}

func updateCustomer(c echo.Context) error {
	fmt.Println()
	fmt.Println("\n🔵 [INFO] Getting request to create customer...")
	fmt.Println()

	customer := new(customers.Customer)

	if err := c.Bind(customer); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	newAddress := customers.Address{
		City:       customer.Address.City,
		Country:    customer.Address.Country,
		PostalCode: customer.Address.PostalCode,
		Province:   customer.Address.Province,
		Street:     customer.Address.Street,
	}

	newCustomer := customers.Customer{
		Address:   newAddress,
		Company:   customer.Company,
		Email:     customer.Email,
		FirstName: customer.FirstName,
		LastName:  customer.LastName,
		Lgpd:      customer.Lgpd,
		NifCif:    customer.NifCif,
		Phone:     customer.Phone,
	}

	customerCreated, err := customers.Create(newCustomer)

	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, customerCreated)
}

func getServer(publicDirectory string, environment *string) *echo.Echo {
	server := echo.New()

	server.Use(middleware.Logger())

	if *environment == environments["development"] {
		server.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: []string{"*"},
		}))
	}

	server.Logger.SetLevel(log.DEBUG)

	server.File("/", path.Join(publicDirectory, "index.html"))
	server.File("/.well-known/apple-developer-merchantid-domain-association", path.Join(publicDirectory, ".well-known/apple-developer-merchantid-domain-association"))

	server.Static("/javascripts", path.Join(publicDirectory, "javascripts"))
	server.Static("/stylesheets", path.Join(publicDirectory, "stylesheets"))
	server.Static("/images", path.Join(publicDirectory, "images"))

	server.GET("/wines", getWines)
	server.GET("/wines/:wine_id/skus", getWineSkus)
	server.GET("/wines/:wine_id", getWine)
	server.GET("/prices", getPrices)
	server.GET("/prices/:wine_id", getWinePrice)
	server.POST("/payment-intents", getPaymentIntent)
	server.POST("/payment-intents/:id/shipping-change", getPaymentIntentShippingChange)
	server.POST("/payment-intents/:id/currency", updatePaymentIntentCurrency)
	server.GET("/payment-intents/:id/status", getPaymentIntentStatus)
	server.POST("/customers", updateCustomer)

	return server
}
