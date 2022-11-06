package sms

import (
	"reflect"
	"testing"
	"time"

	"github.com/h2non/gock"
)

var (
	nai, _          = time.LoadLocation("Africa/Nairobi")
	now, _          = time.ParseInLocation("2006-01-02 15:04:05", "2022-11-05 15:34:45", nai)
	apiHost         = "http://127.0.0.1"
	apiURL          = apiHost + "/api"
	service Service = Service{
		Key:      "123",
		ClientID: 123,
		Secret:   "123",
		apiHost:  apiHost,
	}
)

func TestService_FetchDeliveryReport(t *testing.T) {
	defer gock.Off()

	gock.New(apiURL).
		Get("/fetch-delivery").
		Reply(200).
		JSON(map[string]interface{}{
			"status":               222,
			"status_message":       "fetched delivery status",
			"unique_id":            123,
			"delivery_status_desc": "DeliveredToTerminal",
			"date_received":        now.Format("2006-01-02 15:04:05"),
			"correlator":           "",
			"msisdn":               "254712345678",
		})

	messageId := 123
	want := DeliveryReportResponse{
		Status:                    Status{Code: 222, Message: "fetched delivery status"},
		MessageId:                 123,
		DeliveryStatusDescription: "DeliveredToTerminal",
		DateReceived:              now,
		Correlator:                "",
		MSISDN:                    "254712345678",
	}

	got, err := service.FetchDeliveryReport(messageId)
	t.Run("", func(t *testing.T) {
		if err != nil {
			t.Errorf("Service.FetchDeliveryReport() error = %v", err)
			return
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Service.FetchDeliveryReport() = %v, want %v", got, want)
		}
	})

}

func TestService_CheckBalance(t *testing.T) {
	defer gock.Off()

	gock.New(apiURL).
		Get("/check-credits").
		Reply(200).
		JSON(map[string]interface{}{
			"status":         222,
			"status_message": "fetched balance",
			"client_name":    "Test Client",
			"api_client_id":  123,
			"sms_credits":    45,
			"sms_threshold":  2,
		})

	want := CheckBalanceReponse{
		Status:     Status{Code: 222, Message: "fetched balance"},
		ClientName: "Test Client",
		ClientId:   123,
		Credits:    45,
		Threshold:  2,
	}

	got, err := service.CheckBalance()
	t.Run("", func(t *testing.T) {
		if err != nil {
			t.Errorf("Service.CheckBalance() error = %v", err)
			return
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Service.CheckBalance() = %v, want %v", got, want)
		}
	})
}

func TestService_SendSMS(t *testing.T) {
	defer gock.Off()

	gock.New(apiURL).
		Post("/send-sms-v1").
		Reply(200).
		JSON(map[string]interface{}{
			"status":         222,
			"status_message": "message sent",
			"unique_id":      789879,
			"credits":        127,
		})

	var (
		serviceId string = ""
		msg       string = "Test"
		msisdn    string = "254712345678"
	)

	want := SendSMSResponse{
		Status:    Status{Code: 222, Message: "message sent"},
		MessageId: 789879,
		Credits:   127,
	}

	got, err := service.SendSMS(serviceId, msg, msisdn)
	t.Run("", func(t *testing.T) {
		if err != nil {
			t.Errorf("Service.SendSMS() error = %v", err)
			return
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Service.SendSMS() = %v, want %v", got, want)
		}
	})
}
