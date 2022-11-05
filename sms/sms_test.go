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
	service Service = Service{
		Key:      "123",
		ClientID: 123,
		Secret:   "123",
		apiHost:  "http://127.0.0.1",
	}
)

func TestService_FetchDeliveryReport(t *testing.T) {
	defer gock.Off()

	gock.New("http://127.0.0.1/api").
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

	type args struct {
		messageId int
	}
	tests := []struct {
		name    string
		fields  Service
		args    args
		want    DeliveryReportResponse
		wantErr bool
	}{
		{
			name:   "",
			fields: service,
			args:   args{messageId: 123},
			want: DeliveryReportResponse{
				Status:                    222,
				StatusMessage:             "fetched delivery status",
				MessageId:                 123,
				DeliveryStatusDescription: "DeliveredToTerminal",
				DateReceived:              now,
				Correlator:                "",
				MSISDN:                    "254712345678",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := Service{
				Key:      tt.fields.Key,
				ClientID: tt.fields.ClientID,
				Secret:   tt.fields.Secret,
				apiHost:  tt.fields.apiHost,
			}
			got, err := service.FetchDeliveryReport(tt.args.messageId)
			if (err != nil) != tt.wantErr {
				t.Errorf("Service.FetchDeliveryReport() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Service.FetchDeliveryReport() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_CheckBalance(t *testing.T) {
	defer gock.Off()

	gock.New("http://127.0.0.1/api").
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

	tests := []struct {
		name    string
		fields  Service
		want    CheckBalanceReponse
		wantErr bool
	}{
		{
			name:   "",
			fields: service,
			want: CheckBalanceReponse{
				Status:        222,
				StatusMessage: "fetched balance",
				ClientName:    "Test Client",
				ClientId:      123,
				Credits:       45,
				Threshold:     2,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := Service{
				Key:      tt.fields.Key,
				ClientID: tt.fields.ClientID,
				Secret:   tt.fields.Secret,
				apiHost:  tt.fields.apiHost,
			}
			got, err := service.CheckBalance()
			if (err != nil) != tt.wantErr {
				t.Errorf("Service.CheckBalance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Service.CheckBalance() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_SendSMS(t *testing.T) {
	defer gock.Off()

	gock.New("http://127.0.0.1/api").
		Post("/send-sms-v1").
		Reply(200).
		JSON(map[string]interface{}{
			"status":         222,
			"status_message": "message sent",
			"unique_id":      789879,
			"credits":        127,
		})

	type args struct {
		serviceId string
		msg       string
		msisdn    string
	}
	tests := []struct {
		name    string
		fields  Service
		args    args
		want    SendSMSResponse
		wantErr bool
	}{
		{
			name:   "",
			fields: service,
			args: args{
				serviceId: "",
				msg:       "Test",
				msisdn:    "254712345678",
			},
			want: SendSMSResponse{
				Status:        222,
				StatusMessage: "message sent",
				UniqueId:      789879,
				Credits:       127,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := Service{
				Key:      tt.fields.Key,
				ClientID: tt.fields.ClientID,
				Secret:   tt.fields.Secret,
				apiHost:  tt.fields.apiHost,
			}
			got, err := service.SendSMS(tt.args.serviceId, tt.args.msg, tt.args.msisdn)
			if (err != nil) != tt.wantErr {
				t.Errorf("Service.SendSMS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Service.SendSMS() = %v, want %v", got, tt.want)
			}
		})
	}
}
