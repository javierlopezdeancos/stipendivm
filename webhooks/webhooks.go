package webhooks

import (
	"fmt"

	"github.com/stripe/stripe-go/v72"

	"github.com/javierlopezdeancos/stipendivm/inventory"
	"github.com/javierlopezdeancos/stipendivm/payments"
)

// HandlePaymentIntent Handle payment intent
func HandlePaymentIntent(event stripe.Event, pi *stripe.PaymentIntent) (bool, error) {
	switch event.Type {
	case "payment_intent.succeeded":
		fmt.Printf("🔔  Webhook received! Payment for PaymentIntent %s succeeded\n", pi.ID)

		for wineId, wineQuantity := range pi.Metadata {
			inventory.UpdateWineStock(wineId, wineQuantity)
		}

		return true, nil

	case "payment_intent.payment_failed":
		if pi.LastPaymentError.PaymentMethod != nil {
			fmt.Printf(
				"🔔  Webhook received! Payment on %s %s for PaymentIntent %s failed\n",
				"payment_method",
				pi.LastPaymentError.PaymentMethod.ID,
				pi.ID,
			)
		} else {
			fmt.Printf(
				"🔔  Webhook received! Payment on %s %s for PaymentIntent %s failed\n",
				"source",
				pi.LastPaymentError.Source.ID,
				pi.ID,
			)
		}

		return true, nil

	default:
		return false, nil
	}
}

func HandleSource(event stripe.Event, source *stripe.Source) (bool, error) {
	paymentIntent := source.Metadata["paymentIntent"]

	if paymentIntent == "" {
		return false, nil
	}

	if source.Status == "chargeable" {
		fmt.Printf("🔔  Webhook received! The source %s is chargeable\n", source.ID)
		return true, payments.ConfirmIntent(paymentIntent, source)
	} else if source.Status == "failed" || source.Status == "canceled" {
		return true, payments.CancelIntent(paymentIntent)
	}

	return false, nil
}
