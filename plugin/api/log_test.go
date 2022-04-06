package api

import (
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
)

func TestLog_validateLevel(t *testing.T) {
	type fields struct {
		Timestamp string
		Message   string
		Level     string
		Metadata  string
	}
	testPass1 := fields{
		Timestamp: "2022-04-04T09:00:35+00:00",
		Message:   "Test message",
		Level:     "INFO",
		Metadata:  "",
	}
	testPass2 := fields{
		Timestamp: "2022-04-04T09:00:35+00:00",
		Message:   "Test message",
		Level:     "ERR",
		Metadata:  "",
	}
	testFailLower1 := fields{
		Timestamp: "2022-04-04T09:00:35+00:00",
		Message:   "Test message",
		Level:     "info",
		Metadata:  "",
	}
	testFailLower2 := fields{
		Timestamp: "2022-04-04T09:00:35+00:00",
		Message:   "Test message",
		Level:     "err",
		Metadata:  "",
	}

	testFail1 := fields{
		Timestamp: "2022-04-04T09:00:35+00:00",
		Message:   "Test message",
		Level:     "errerr",
		Metadata:  "",
	}
	testFail2 := fields{
		Timestamp: "2022-04-04T09:00:35+00:00",
		Message:   "Test message",
		Level:     "ERROR!",
		Metadata:  "",
	}
	testFail3 := fields{
		Timestamp: "2022-04-04T09:00:35+00:00",
		Message:   "Test message",
		Level:     "",
		Metadata:  "",
	}
	testFail4 := fields{
		Timestamp: "2022-04-04T09:00:35+00:00",
		Message:   "Test message",
		Level:     "BoGus",
		Metadata:  "",
	}

	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:    "pass 1",
			fields:  testPass1,
			wantErr: false,
		},
		{
			name:    "pass 2",
			fields:  testPass2,
			wantErr: false,
		},
		{
			name:    "fail lower 1",
			fields:  testFailLower1,
			wantErr: true,
		},
		{
			name:    "fail lower 2",
			fields:  testFailLower2,
			wantErr: true,
		},
		{
			name:    "fail 1",
			fields:  testFail1,
			wantErr: true,
		},
		{
			name:    "fail 2",
			fields:  testFail2,
			wantErr: true,
		},
		{
			name:    "fail 3",
			fields:  testFail3,
			wantErr: true,
		},
		{
			name:    "fail 4",
			fields:  testFail4,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Log{
				Timestamp: tt.fields.Timestamp,
				Message:   tt.fields.Message,
				Level:     tt.fields.Level,
				Metadata:  tt.fields.Metadata,
			}
			if err := l.validateLevel(); (err != nil) != tt.wantErr {
				t.Errorf("validateLevel() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLog_validateTimestamp(t *testing.T) {
	type fields struct {
		Timestamp string
		Message   string
		Level     string
		Metadata  string
	}
	testPass1 := fields{
		Timestamp: "2022-04-04T09:00:35+00:00",
		Message:   "Test message",
		Level:     "INFO",
		Metadata:  "",
	}
	testPass2 := fields{
		Timestamp: "2022-04-04T09:00:35.1111+00:00",
		Message:   "Test message",
		Level:     "INFO",
		Metadata:  "",
	}
	testPass3 := fields{
		Timestamp: "2022-04-04T09:00:35.1111",
		Message:   "Test message",
		Level:     "INFO",
		Metadata:  "",
	}
	testPass4 := fields{
		Timestamp: "2022-04-04T09:00:35",
		Message:   "Test message",
		Level:     "INFO",
		Metadata:  "",
	}
	testPass5 := fields{
		Timestamp: "2022-04-04T09:00:35Z",
		Message:   "Test message",
		Level:     "INFO",
		Metadata:  "",
	}
	testPass6 := fields{
		Timestamp: "2022-04-04T09:00:35.111Z",
		Message:   "Test message",
		Level:     "INFO",
		Metadata:  "",
	}

	testFail1 := fields{
		Timestamp: "2022-04-04T09:00",
		Message:   "Test message",
		Level:     "INFO",
		Metadata:  "",
	}
	testFail2 := fields{
		Timestamp: "2022-04-04T09:00:35Z+02:00",
		Message:   "Test message",
		Level:     "INFO",
		Metadata:  "",
	}

	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:    "pass 1",
			fields:  testPass1,
			wantErr: false,
		},
		{
			name:    "pass 2",
			fields:  testPass2,
			wantErr: false,
		},
		{
			name:    "pass 3",
			fields:  testPass3,
			wantErr: false,
		},
		{
			name:    "pass 4",
			fields:  testPass4,
			wantErr: false,
		},
		{
			name:    "pass 5",
			fields:  testPass5,
			wantErr: false,
		},
		{
			name:    "pass 6",
			fields:  testPass6,
			wantErr: false,
		},
		{
			name:    "fail 1",
			fields:  testFail1,
			wantErr: true,
		},
		{
			name:    "fail 2",
			fields:  testFail2,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Log{
				Timestamp: tt.fields.Timestamp,
				Message:   tt.fields.Message,
				Level:     tt.fields.Level,
				Metadata:  tt.fields.Metadata,
			}
			if err := l.validateTimestamp(); (err != nil) != tt.wantErr {
				t.Errorf("validateTimestamp() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLogApi_SendLogBatch(t *testing.T) {
	idStr := "27596b04-f260-4bc0-ab02-e437a454ef90"
	idUUID, _ := uuid.Parse(idStr)
	logBatchRequest := &LogBatchRequest{
		ApplicationId: idUUID,
		Tag:           "default",
		Logs: []*Log{{
			Timestamp: "2022-04-04T09:00:35",
			Message:   "Test message",
			Level:     "INFO",
			Metadata:  "",
		}},
	}
	logReceitp := &LogReceipt{
		ReceiptId:     idUUID,
		LogsCount:     1,
		Source:        "restBatch",
		ApplicationId: idUUID,
	}

	type fields struct {
		BaseApi *BaseApi
		User    *User
	}
	type args struct {
		logBatchReq *LogBatchRequest
	}
	jsonLogReceiptValid := []byte(fmt.Sprintf(
		`{"receiptId":"%v","logsCount":1,"source":"restBatch","applicationId":"%v"}`, idStr, idUUID))
	jsonLogReceiptInvalid := []byte(fmt.Sprintf(
		`{"receiptId":"%v","logsCount":"1","source":"restBatch","applicationId":"%v"}`, idStr, idUUID))

	// generate a test server, so we can capture and inspect the request
	testServerPassValidReceipt := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
		_, _ = res.Write(jsonLogReceiptValid)
	}))
	defer func() { testServerPassValidReceipt.Close() }()
	testServerPassInvalidReceipt := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
		_, _ = res.Write(jsonLogReceiptInvalid)
	}))
	defer func() { testServerPassInvalidReceipt.Close() }()
	testServerFail := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusBadRequest)
		_, _ = res.Write([]byte(`{"message":"failed"}`))
	}))
	defer func() { testServerFail.Close() }()
	testServerFailRetry1 := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusConflict)
		_, _ = res.Write([]byte(`{"message":"failed"}`))
	}))
	defer func() { testServerFailRetry1.Close() }()
	testServerFailRetry2 := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusNotFound)
		_, _ = res.Write([]byte(`{"message":"failed"}`))
	}))
	defer func() { testServerFailRetry2.Close() }()

	httpClient := http.DefaultClient
	urlTestServerPassValidReceipt, _ := url.Parse(testServerPassValidReceipt.URL)
	urlTestServerPassInvalidReceipt, _ := url.Parse(testServerPassInvalidReceipt.URL)
	urlTestServerFail, _ := url.Parse(testServerFail.URL)
	urlTestServerFailRetry1, _ := url.Parse(testServerFailRetry1.URL)
	urlTestServerFailRetry2, _ := url.Parse(testServerFailRetry2.URL)
	baseApiPassValidReceipt := &BaseApi{HttpClient: httpClient, Url: urlTestServerPassValidReceipt}
	baseApiPassInvalidReceipt := &BaseApi{HttpClient: httpClient, Url: urlTestServerPassInvalidReceipt}
	baseApiFail := &BaseApi{HttpClient: httpClient, Url: urlTestServerFail}
	baseApiFailRetry1 := &BaseApi{HttpClient: httpClient, Url: urlTestServerFailRetry1}
	baseApiFailRetry2 := &BaseApi{HttpClient: httpClient, Url: urlTestServerFailRetry2}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *LogReceipt
		wantErr bool
	}{
		{
			name:    "pass valid receipt",
			fields:  fields{User: &User{}, BaseApi: baseApiPassValidReceipt},
			args:    args{logBatchReq: logBatchRequest},
			want:    logReceitp,
			wantErr: false,
		},
		{
			name:    "pass invalid receipt",
			fields:  fields{User: &User{}, BaseApi: baseApiPassInvalidReceipt},
			args:    args{logBatchReq: logBatchRequest},
			want:    nil,
			wantErr: false,
		},
		{
			name:    "fail",
			fields:  fields{User: &User{}, BaseApi: baseApiFail},
			args:    args{logBatchReq: logBatchRequest},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "fail retry1",
			fields:  fields{User: &User{}, BaseApi: baseApiFailRetry1},
			args:    args{logBatchReq: logBatchRequest},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "fail retry2",
			fields:  fields{User: &User{}, BaseApi: baseApiFailRetry2},
			args:    args{logBatchReq: logBatchRequest},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			la := &LogApi{
				BaseApi: tt.fields.BaseApi,
				User:    tt.fields.User,
			}
			got, err := la.SendLogBatch(tt.args.logBatchReq)
			if (err != nil) != tt.wantErr {
				t.Errorf("SendLogBatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SendLogBatch() got = %v, want %v", got, tt.want)
			}
		})
	}
}
