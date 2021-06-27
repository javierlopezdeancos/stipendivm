package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/webhook"

	"github.com/javierlopezdeancos/stipendivm/config"
	"github.com/javierlopezdeancos/stipendivm/customers"
	"github.com/javierlopezdeancos/stipendivm/inventory"
	"github.com/javierlopezdeancos/stipendivm/payments"
	"github.com/javierlopezdeancos/stipendivm/webhooks"
	"github.com/javierlopezdeancos/stipendivm/wine"
)

func main() {
	rootDirectory := flag.String("root", "./", "Root directory of the Stipendivm server to Quantvm stripe payments")
	environment := flag.String("env", "dev", "Type of environment to start Stipendivm server")
	config.Environment = *environment

	flag.Parse()

	if *rootDirectory == "" {
		*rootDirectory = "./"
	}

	if config.Environment == "" {
		config.Environment = config.Environments["development"]
	}

	var err error

	if config.Environment == config.Environments["development"] {
		err = godotenv.Load(path.Join(*rootDirectory, ".env.development"))
	} else if config.Environment == config.Environments["production"] {
		err = godotenv.Load(path.Join(*rootDirectory, ".env"))
	}

	if err != nil {
		panic(fmt.Sprintf("error loading .env: %v", err))
	}

	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	if stripe.Key == "" {
		panic("STRIPE_SECRET_KEY must be in environment")
	}

	config.PublicDirectory = path.Join(*rootDirectory, "public")
	server := getServer()
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

type RequestErrorMetaWine struct {
	Id    string
	Stock string
}

type RequestErrorMeta struct {
	Wines []RequestErrorMetaWine
}

type RequestCustomError struct {
	Message string
	Meta    RequestErrorMeta
}

func hasWineAtLeastOneBottle(quantity int64) bool {
	return quantity > 0
}
func getPaymentIntent(c echo.Context) error {
	/* **********************************************************
	// mock the error to no wine bottle stock

	noStockError := &RequestCustomError{
		Message: "A wine in shopping cart has quantity great than stock",
		Meta: RequestErrorMeta{
			Wines: []RequestErrorMetaWine{
				{
					Id:    "product-wine-bottle-75cl-cristal-sel-d-aiz-yenda-albarinio-godello",
					Stock: "5",
				},
			},
		},
	}

	return c.JSON(http.StatusNotAcceptable, noStockError)

	***********************************************************/

	ir := new(payments.IntentCreationRequest)
	err := c.Bind(ir)

	if err != nil {
		return err
	}

	var wines []inventory.Item = ir.Items

	for _, w := range wines {
		product, err := inventory.RetrieveWine(w.Parent)

		if err != nil {
			return err
		}

		wineHasAtLeastOneBottle := hasWineAtLeastOneBottle(w.Quantity)

		if !wineHasAtLeastOneBottle {
			noMoreThanOneBottleSelectedError := &RequestCustomError{
				Message: fmt.Sprint(
					"Sorry,the wine ", product.Name, " there are not bottles selected to create a paymen intent",
				),
			}

			return c.JSON(http.StatusNotAcceptable, noMoreThanOneBottleSelectedError)
		}

		jsonMetadataProduct, err := json.Marshal(product.Metadata)

		if err != nil {
			return err
		}

		wineMetadata := wine.Metadata{}

		if err := json.Unmarshal(jsonMetadataProduct, &wineMetadata); err != nil {
			return err
		}

		quantity, err := strconv.Atoi(wineMetadata.Quantity)

		if err != nil {
			return err
		}

		if quantity < int(w.Quantity) {
			noWineStockError := &RequestCustomError{
				Message: fmt.Sprint(
					"Sorry, the wine ",
					product.Name,
					", not have stock enough to create your payment order with ",
					w.Quantity,
					" bottles",
				),
			}

			return c.JSON(http.StatusNotAcceptable, noWineStockError)
		}
	}

	pi, err := payments.CreateIntent(ir)

	if err != nil {
		return err
	}

	return c.JSON(
		http.StatusOK,
		map[string]*stripe.PaymentIntent{
			"paymentIntent": pi,
		},
	)
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

	/*
		if pi.LastPaymentError != nil {
			p.PaymentIntent.LastPaymentError = pi.LastPaymentError.Message
		}
	*/

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

func handleWebhook(c echo.Context) error {
	request := c.Request()
	payload, err := ioutil.ReadAll(request.Body)

	if err != nil {
		return err
	}

	var event stripe.Event
	webhookSecret := os.Getenv("STRIPE_SHOPPING_CART_WEBHOOK_SECRET")

	if webhookSecret != "" {
		event, err = webhook.ConstructEvent(payload, request.Header.Get("Stripe-Signature"), webhookSecret)

		if err != nil {
			return err
		}
	} else {
		err := json.Unmarshal(payload, &event)

		if err != nil {
			return err
		}
	}

	eventDataObjectType := event.Data.Object["object"].(string)

	var handled bool

	switch eventDataObjectType {
	case "payment_intent":
		var pi *stripe.PaymentIntent
		err = json.Unmarshal(event.Data.Raw, &pi)
		if err != nil {
			return err
		}

		handled, err = webhooks.HandlePaymentIntent(event, pi)
	case "source":
		var source *stripe.Source
		err := json.Unmarshal(event.Data.Raw, &source)
		if err != nil {
			return err
		}

		handled, err = webhooks.HandleSource(event, source)
	}

	if err != nil {
		return err
	}

	if !handled {
		fmt.Printf("🔔  Webhook received and not handled! %s\n", event.Type)
	}

	return nil
}

func getServer() *echo.Echo {
	server := echo.New()

	server.Use(middleware.Logger())

	if config.Environment == config.Environments["development"] {
		server.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: []string{"*"},
		}))
	}

	server.Logger.SetLevel(log.DEBUG)

	server.File("/", path.Join(config.PublicDirectory, "index.html"))
	server.File("/.well-known/apple-developer-merchantid-domain-association", path.Join(config.PublicDirectory, ".well-known/apple-developer-merchantid-domain-association"))

	server.Static("/javascripts", path.Join(config.PublicDirectory, "javascripts"))
	server.Static("/stylesheets", path.Join(config.PublicDirectory, "stylesheets"))
	server.Static("/images", path.Join(config.PublicDirectory, "images"))

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

	server.POST("/webhook/shopping-cart", handleWebhook)

	return server
}
