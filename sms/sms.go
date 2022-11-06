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

type Status struct {
	Code    int    `json:"status"`
	Message string `json:"status_message"`
}

type SendSMSResponse struct {
	Status
	MessageId int `json:"unique_id"`
	Credits   int `json:"credits"`
}

type CheckBalanceReponse struct {
	Status
	ClientName string `json:"client_name"`
	ClientId   int    `json:"api_client_id"`
	Credits    int    `json:"sms_credits"`
	Threshold  int    `json:"sms_threshold"`
}

type DeliveryReportResponse struct {
	Status
	MessageId                 int       `json:"unique_id"`
	DeliveryStatusDescription string    `json:"delivery_status_desc"`
	DateReceived              time.Time `json:"date_received"`
	Correlator                string    `json:"correlator"`
	MSISDN                    string    `json:"msisdn"`
}

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

	res, err := service.makeRequest("GET", GetDeliveryReportURL(service.apiHost), nil, params)
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

func (service Service) CheckBalance() (CheckBalanceReponse, error) {
	var (
		response CheckBalanceReponse
		err      error
	)

	params := url.Values{}
	params.Add("apiClientID", strconv.Itoa(service.ClientID))
	params.Add("key", service.Key)

	res, err := service.makeRequest("GET", GetBalanceURL(service.apiHost), nil, params)
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

	res, err := service.makeRequest("POST", GetSMSURL(service.apiHost), formData, nil)
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
