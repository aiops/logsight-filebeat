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

func TestAutoCreateMissingApplication_getApplicationByName(t *testing.T) {
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
			got, err := aca.getApplicationByName(tt.args.appName)
			if (err != nil) != tt.wantErr {
				t.Errorf("getApplicationByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getApplicationByName() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAutoCreateMissingApplication_handleMissingApplication(t *testing.T) {
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
			got, err := aca.handleMissingApplication(tt.args.appName)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleMissingApplication() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("handleMissingApplication() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestErrorOnMissingApplication_getApplicationByName(t *testing.T) {
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
			got, err := ea.getApplicationByName(tt.args.appName)
			if (err != nil) != tt.wantErr {
				t.Errorf("getApplicationByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getApplicationByName() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestErrorOnMissingApplication_handleMissingApplication(t *testing.T) {
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
			got, err := ea.handleMissingApplication(tt.args.appName)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleMissingApplication() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("handleMissingApplication() got = %v, want %v", got, tt.want)
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

func Test_getApplicationByName(t *testing.T) {
	type args struct {
		appName string
		api     ApplicationApiInterface
	}
	tests := []struct {
		name    string
		args    args
		want    *Application
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getApplicationByName(tt.args.appName, tt.args.api)
			if (err != nil) != tt.wantErr {
				t.Errorf("getApplicationByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getApplicationByName() got = %v, want %v", got, tt.want)
			}
		})
	}
}
