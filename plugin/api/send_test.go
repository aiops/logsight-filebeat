package api

import (
	"testing"
)

func TestLogSender_Close(t *testing.T) {
	type fields struct {
		LogApi *LogApi
	}
	var tests []struct {
		name   string
		fields fields
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			as := LogSender{
				LogApi: tt.fields.LogApi,
			}
			as.Close()
		})
	}
}

func TestLogSender_Send(t *testing.T) {
	type fields struct {
		LogApi *LogApi
	}
	type args struct {
		logs []*Log
	}
	var tests []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			as := LogSender{
				LogApi: tt.fields.LogApi,
			}
			if err := as.Send(tt.args.logs); (err != nil) != tt.wantErr {
				t.Errorf("Send() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
