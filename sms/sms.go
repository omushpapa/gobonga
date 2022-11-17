// Package sms provides an interface for interacting with the
// sms resources of the BongaSMS API
package sms

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var defaultAPIHost string = "https://app.bongasms.co.ke"

// GetAPIURL returns the URL for the API.
func GetAPIURL(host string) string {
	return host + "/api"
}

// GetBalanceURL returns the URL for fetching credit balance information.
func GetBalanceURL(host string) string {
	return GetAPIURL(host) + "/check-credits"
}

// GetSMSURL returns the URL for POSTing SMS messages.
func GetSMSURL(host string) string {
	return GetAPIURL(host) + "/send-sms-v1"
}

// GetDeliveryReportURL returns the URL for fetching the delivery report for a message.
func GetDeliveryReportURL(host string) string {
	return GetAPIURL(host) + "/fetch-delivery"
}

// Status is a representation of the status of the action performed on the server side
type Status struct {
	// Code is a custom number indicating whether the action was a success or not.
	// 222 represents success while 666 represents failure.
	Code int `json:"status"`
	// Message is verbose text describing the status of the action.
	Message string `json:"status_message"`
}

// SendSMSResponse is a representation of the response returned when sending an SMS.
type SendSMSResponse struct {
	Status
	// MessageId is a unique number assigned to the SMS.
	MessageId int `json:"unique_id"`
	// Credits represents the remaining balance.
	// This is the number of messages that can be sent.
	Credits int `json:"credits"`
}

// CheckBalanceReponse is a representation of the response returned when requesting credit balance.
type CheckBalanceReponse struct {
	Status
	// ClientName is the name of the client in use.
	ClientName string `json:"client_name"`
	// ClientId is the id assigned to the client in use.
	ClientId int `json:"api_client_id"`
	// Credits represents the remaining balance.
	// This is the number of messages that can be sent.
	Credits int `json:"sms_credits"`
	// Threshold is the number at which an email notification will be sent when Credits go below.
	Threshold int `json:"sms_threshold"`
}

// DeliveryReportResponse is a representation of the response returned when requesting for the delivery
// status of an SMS
type DeliveryReportResponse struct {
	Status
	// MessageId is the unique number assigned to the SMS.
	MessageId int `json:"unique_id"`
	// DeliveryStatusDescription is the delivery state at which the SMS is at.
	DeliveryStatusDescription string `json:"delivery_status_desc"`
	// DateReceived represents the date and time when the state given by the DeliveryStatusDescription
	// was received.
	DateReceived time.Time `json:"date_received"`
	// Correlator is a value provided by the sender which identifies the message on the sender's side.
	Correlator string `json:"correlator"`
	// MSISDN is the contact to which the SMS was sent.
	MSISDN string `json:"msisdn"`
}

// UnmarshalJSON is the method called when JSON is being unmarshalled.
// It is implemented here to convert the date/time string received for DateReceived into and actual time.Time value.
func (r *DeliveryReportResponse) UnmarshalJSON(p []byte) error {
	var i struct {
		Status
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
	r.MessageId = i.MessageId
	r.DeliveryStatusDescription = i.DeliveryStatusDescription
	r.Correlator = i.Correlator
	r.MSISDN = i.MSISDN
	return nil
}

// A Service calls the BongaSMS API resources and parses the response into structured formats.
type Service struct {
	Key      string // The API key given on registration
	Secret   string // The API secret given on registration
	ClientID int    // The ID of the client
	apiHost  string // The link to the host
}

// NewService uses the given credentials and the defaultAPIHost to configure a new Service which is then returned.
func NewService(key string, clientID int, secret string) Service {
	return Service{
		Key:      key,
		ClientID: clientID,
		Secret:   secret,
		apiHost:  defaultAPIHost,
	}
}

// FetchDeliveryReport interacts with the resource for retrieving the delivery report of an SMS
// and parses the response into a DeliveryReportResponse.
func (service Service) FetchDeliveryReport(messageId int) (DeliveryReportResponse, error) {
	var (
		response DeliveryReportResponse
		err      error
	)
	params := url.Values{}
	params.Add("apiClientID", strconv.Itoa(service.ClientID))
	params.Add("key", service.Key)
	params.Add("unique_id", strconv.Itoa(messageId))

	res, err := service.makeRequest(http.MethodGet, GetDeliveryReportURL(service.apiHost), nil, params)
	if err != nil {
		return response, err
	}
	defer res.Body.Close()

	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return response, err
	}

	err = service.checkStatus(response.Status)
	return response, err
}

// CheckBalance interacts with the resource for retrieving the credit balance
// and parses the response into a CheckBalanceReponse.
func (service Service) CheckBalance() (CheckBalanceReponse, error) {
	var (
		response CheckBalanceReponse
		err      error
	)

	params := url.Values{}
	params.Add("apiClientID", strconv.Itoa(service.ClientID))
	params.Add("key", service.Key)

	res, err := service.makeRequest(http.MethodGet, GetBalanceURL(service.apiHost), nil, params)
	if err != nil {
		return response, err
	}
	defer res.Body.Close()

	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return response, err
	}
	err = service.checkStatus(response.Status)
	return response, err
}

// SendSMS interacts with the resource for sending an SMS and parses the response into a SendSMSResponse.
func (service Service) SendSMS(serviceId string, msg string, msisdn string) (SendSMSResponse, error) {
	var (
		response SendSMSResponse
		err      error
	)

	formData := url.Values{}
	formData.Add("apiClientID", strconv.Itoa(service.ClientID))
	formData.Add("key", service.Key)
	formData.Add("secret", service.Secret)
	formData.Add("txtMessage", msg)
	formData.Add("MSISDN", msisdn)

	res, err := service.makeRequest(http.MethodPost, GetSMSURL(service.apiHost), formData, nil)
	if err != nil {
		return response, err
	}
	defer res.Body.Close()

	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return response, err
	}
	err = service.checkStatus(response.Status)
	return response, err
}

func (service Service) checkStatus(status Status) error {
	if status.Code == 666 {
		return fmt.Errorf(status.Message)
	}
	return nil
}

func (service Service) makeRequest(method string, path string, formData url.Values, params url.Values) (*http.Response, error) {
	var body io.Reader = nil
	if len(formData) > 0 {
		body = strings.NewReader(formData.Encode())
	}

	request, err := http.NewRequest(method, path, body)
	if err != nil {
		return nil, err
	}

	if method == "POST" {
		request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}
	if len(params) > 0 {
		request.URL.RawQuery = params.Encode()
	}

	client := http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return res, fmt.Errorf(res.Status)
	}
	return res, nil
}
