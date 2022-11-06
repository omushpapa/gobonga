package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/giantas/gobonga/sms"
)

var (
	clientID  int64
	phone     string
	serviceID string = os.Getenv("API_SERVICE_ID")
	apiKey    string = os.Getenv("API_KEY")
	apiSecret string = os.Getenv("API_SECRET")
)

func main() {
	flag.Int64Var(&clientID, "client-id", 0, "Client ID")
	flag.StringVar(&serviceID, "service-id", "", "Service ID")
	flag.StringVar(&phone, "phone", "", "The recipient phone number")
	flag.Parse()

	if clientID == 0 || phone == "" {
		exitWithError(fmt.Errorf("client ID and/or phone number not provided"))
	}

	if apiKey == "" || apiSecret == "" {
		exitWithError(fmt.Errorf("missing environment variables: API_KEY, API_SECRET"))
	}

	service := sms.NewService(apiKey, int(clientID), apiSecret)

	smsResponse, err := service.SendSMS(serviceID, "Test", "+254701556803")
	if err != nil {
		exitWithError(err)
	}
	fmt.Printf("status: %s\n", smsResponse.Status.Message)

	balanceResponse, err := service.CheckBalance()
	if err != nil {
		exitWithError(err)
	}
	fmt.Printf(
		"status: %s. %d credits remaining for client %d:'%s' \n",
		balanceResponse.Status.Message, balanceResponse.Credits, balanceResponse.ClientId, balanceResponse.ClientName,
	)

	messageId := smsResponse.MessageId
	deliveryResponse, err := service.FetchDeliveryReport(messageId)
	if err != nil {
		exitWithError(err)
	}
	fmt.Printf(
		"status: %s. message of id %d has reached status %s for contact %s as at %s\n",
		deliveryResponse.Status.Message, deliveryResponse.MessageId, deliveryResponse.DeliveryStatusDescription,
		deliveryResponse.MSISDN, deliveryResponse.DateReceived.Format(time.RFC3339),
	)
}

func exitWithError(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(1)
}
