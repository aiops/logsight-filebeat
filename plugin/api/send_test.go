package api

import (
	"github.com/google/uuid"
	"reflect"
	"testing"
)

func TestApplicationNotFoundError_Error(t *testing.T) {
	type fields struct {
		applicationName string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := ApplicationNotFoundError{
				applicationName: tt.fields.applicationName,
			}
			if got := e.Error(); got != tt.want {
				t.Errorf("Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAutoCreateMissingApplication_createMissingApplication(t *testing.T) {
	type fields struct {
		MissingApplicationHandler MissingApplicationHandler
		ApplicationApi            ApplicationApiInterface
	}
	type args struct {
		appName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Application
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aca := AutoCreateMissingApplication{
				MissingApplicationHandler: tt.fields.MissingApplicationHandler,
				ApplicationApi:            tt.fields.ApplicationApi,
			}
			got, err := aca.createMissingApplication(tt.args.appName)
			if (err != nil) != tt.wantErr {
				t.Errorf("createMissingApplication() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createMissingApplication() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAutoCreateMissingApplication_handleApplication(t *testing.T) {
	type fields struct {
		MissingApplicationHandler MissingApplicationHandler
		ApplicationApi            ApplicationApiInterface
	}
	type args struct {
		appName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Application
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aca := AutoCreateMissingApplication{
				MissingApplicationHandler: tt.fields.MissingApplicationHandler,
				ApplicationApi:            tt.fields.ApplicationApi,
			}
			got, err := aca.handleApplication(tt.args.appName)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleApplication() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("handleApplication() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestErrorOnMissingApplication_handleApplication(t *testing.T) {
	type fields struct {
		MissingApplicationHandler MissingApplicationHandler
		ApplicationApi            ApplicationApiInterface
	}
	type args struct {
		appName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Application
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ea := ErrorOnMissingApplication{
				MissingApplicationHandler: tt.fields.MissingApplicationHandler,
				ApplicationApi:            tt.fields.ApplicationApi,
			}
			got, err := ea.handleApplication(tt.args.appName)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleApplication() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("handleApplication() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLogBatch_ToLogBatchRequest(t *testing.T) {
	type fields struct {
		ApplicationName string
		Tag             string
		Logs            []*Log
	}
	type args struct {
		applicationId uuid.UUID
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *LogBatchRequest
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lb := &LogBatch{
				ApplicationName: tt.fields.ApplicationName,
				Tag:             tt.fields.Tag,
				Logs:            tt.fields.Logs,
			}
			if got := lb.ToLogBatchRequest(tt.args.applicationId); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToLogBatchRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLogSender_Close(t *testing.T) {
	type fields struct {
		LogApi            *LogApi
		MissingAppHandler MissingApplicationHandler
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			as := LogSender{
				LogApi:            tt.fields.LogApi,
				MissingAppHandler: tt.fields.MissingAppHandler,
			}
			as.Close()
		})
	}
}

func TestLogSender_Send(t *testing.T) {
	type fields struct {
		LogApi            *LogApi
		MissingAppHandler MissingApplicationHandler
	}
	type args struct {
		logBatch *LogBatch
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			as := LogSender{
				LogApi:            tt.fields.LogApi,
				MissingAppHandler: tt.fields.MissingAppHandler,
			}
			if err := as.Send(tt.args.logBatch); (err != nil) != tt.wantErr {
				t.Errorf("Send() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}