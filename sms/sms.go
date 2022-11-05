package sms

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var defaultAPIHost string = "https://app.bongasms.co.ke"

func GetAPIURL(host string) string {
	return host + "/api"
}

func GetBalanceURL(host string) string {
	return GetAPIURL(host) + "/check-credits"
}

func GetSMSURL(host string) string {
	return GetAPIURL(host) + "/send-sms-v1"
}

func GetDeliveryReportURL(host string) string {
	return GetAPIURL(host) + "/fetch-delivery"
}

type SendSMSResponse struct {
	Status        int    `json:"status"`
	StatusMessage string `json:"status_message"`
	UniqueId      int    `json:"unique_id"`
	Credits       int    `json:"credits"`
}

type CheckBalanceReponse struct {
	Status        int    `json:"status"`
	StatusMessage string `json:"status_message"`
	ClientName    string `json:"client_name"`
	ClientId      int    `json:"api_client_id"`
	Credits       int    `json:"sms_credits"`
	Threshold     int    `json:"sms_threshold"`
}

type DeliveryReportResponse struct {
	Status                    int       `json:"status"`
	StatusMessage             string    `json:"status_message"`
	MessageId                 int       `json:"unique_id"`
	DeliveryStatusDescription string    `json:"delivery_status_desc"`
	DateReceived              time.Time `json:"date_received"`
	Correlator                string    `json:"correlator"`
	MSISDN                    string    `json:"msisdn"`
}

func (r *DeliveryReportResponse) UnmarshalJSON(p []byte) error {
	var i struct {
		Status                    int    `json:"status"`
		StatusMessage             string `json:"status_message"`
		MessageId                 int    `json:"unique_id"`
		DeliveryStatusDescription string `json:"delivery_status_desc"`
		DateReceived              string `json:"date_received"`
		Correlator                string `json:"correlator"`
		MSISDN                    string `json:"msisdn"`
	}

	err := json.Unmarshal(p, &i)
	if err != nil {
		return err
	}

	nai, err := time.LoadLocation("Africa/Nairobi")
	if err != nil {
		return err
	}

	t, err := time.ParseInLocation("2006-01-02 15:04:05", i.DateReceived, nai)
	if err != nil {
		return err
	}

	r.DateReceived = t
	r.Status = i.Status
	r.StatusMessage = i.StatusMessage
	r.MessageId = i.MessageId
	r.DeliveryStatusDescription = i.DeliveryStatusDescription
	r.Correlator = i.Correlator
	r.MSISDN = i.MSISDN
	return nil
}

type Service struct {
	Key      string
	ClientID int
	Secret   string
	apiHost  string
}

func NewService(key string, clientID int, secret string) Service {
	return Service{
		Key:      key,
		ClientID: clientID,
		Secret:   secret,
		apiHost:  defaultAPIHost,
	}
}

func (service Service) FetchDeliveryReport(messageId int) (DeliveryReportResponse, error) {
	var (
		response DeliveryReportResponse
		err      error
	)
	params := url.Values{}
	params.Add("apiClientID", strconv.Itoa(service.ClientID))
	params.Add("key", service.Key)
	params.Add("unique_id", strconv.Itoa(messageId))

	request, err := http.NewRequest("GET", GetDeliveryReportURL(service.apiHost), nil)
	if err != nil {
		return response, err
	}

	request.URL.RawQuery = params.Encode()
	client := http.Client{}
	res, err := client.Do(request)
	if err != nil {
		return response, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return response, fmt.Errorf(res.Status)
	}

	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return response, err
	}

	if response.Status == 666 {
		return response, fmt.Errorf(response.StatusMessage)
	}
	return response, err
}

func (service Service) CheckBalance() (CheckBalanceReponse, error) {
	var (
		response CheckBalanceReponse
		err      error
	)

	params := url.Values{}
	params.Add("apiClientID", strconv.Itoa(service.ClientID))
	params.Add("key", service.Key)

	request, err := http.NewRequest("GET", GetBalanceURL(service.apiHost), nil)
	if err != nil {
		return response, err
	}
	request.URL.RawQuery = params.Encode()

	client := http.Client{}
	res, err := client.Do(request)
	if err != nil {
		return response, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return response, fmt.Errorf(res.Status)
	}

	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return response, err
	}
	if response.Status == 666 {
		return response, fmt.Errorf(response.StatusMessage)
	}
	return response, err
}

func (service Service) SendSMS(serviceId string, msg string, msisdn string) (SendSMSResponse, error) {
	var (
		response SendSMSResponse
		err      error
	)
	form := url.Values{}
	form.Add("apiClientID", strconv.Itoa(service.ClientID))
	form.Add("key", service.Key)
	form.Add("secret", service.Secret)
	form.Add("txtMessage", msg)
	form.Add("MSISDN", msisdn)

	request, err := http.NewRequest("POST", GetSMSURL(service.apiHost), strings.NewReader(form.Encode()))
	if err != nil {
		return response, err
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	client := http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(request)
	if err != nil {
		return response, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return response, fmt.Errorf(res.Status)
	}

	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return response, err
	}
	if response.Status == 666 {
		return response, fmt.Errorf(response.StatusMessage)
	}
	return response, err
}
