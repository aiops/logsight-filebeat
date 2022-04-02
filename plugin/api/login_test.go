package api

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestLoginApi_Login(t *testing.T) {
	// generate a test server so we can capture and inspect the request
	testServerOK := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
		_, _ = res.Write([]byte("body"))
	}))
	defer func() { testServerOK.Close() }()

	type fields struct {
		BaseApi *BaseApi
	}
	type args struct {
		loginReq LoginRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *LoginResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			la := &LoginApi{
				BaseApi: tt.fields.BaseApi,
			}
			got, err := la.Login(tt.args.loginReq)
			if (err != nil) != tt.wantErr {
				t.Errorf("Login() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Login() got = %v, want %v", got, tt.want)
			}
		})
	}
}
