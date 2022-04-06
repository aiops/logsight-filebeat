package api

import (
	"fmt"
	"github.com/google/uuid"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func TestLoginApi_Login(t *testing.T) {
	bearerToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9" +
		".eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ" +
		".SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"
	userId := "27596b04-f260-4bc0-ab02-e437a454ef90"
	userEmail := "hari.seldon@fundation.gal"
	jsonLogin := []byte(fmt.Sprintf(`{"token":"%v","user":{"userId":"%v","email":"%v"}}`,
		bearerToken, userId, userEmail))

	// generate a test server, so we can capture and inspect the request
	testServerPass := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
		_, _ = res.Write(jsonLogin)
	}))
	defer func() { testServerPass.Close() }()
	testServerFail := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusBadRequest)
		_, _ = res.Write([]byte(`{"message":"failed"}`))
	}))
	defer func() { testServerFail.Close() }()

	httpClient := http.DefaultClient
	urlTestServerPass, _ := url.Parse(testServerPass.URL)
	urlTestServerFail, _ := url.Parse(testServerFail.URL)
	baseApiPass := &BaseApi{HttpClient: httpClient, Url: urlTestServerPass}
	baseApiFail := &BaseApi{HttpClient: httpClient, Url: urlTestServerFail}

	userUUID, _ := uuid.Parse(userId)
	expectedLoginResponse := &LoginResponse{
		Token: &bearerToken,
		User: &UserDTO{
			Id:    &userUUID,
			Email: &userEmail,
		},
	}

	loginReq := LoginRequest{
		Email:    "hari.seldon@fundation.gal",
		Password: "foundation_rulez",
	}

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
		{
			name:    "pass",
			fields:  fields{BaseApi: baseApiPass},
			args:    args{loginReq: loginReq},
			want:    expectedLoginResponse,
			wantErr: false,
		},
		{
			name:    "fails",
			fields:  fields{BaseApi: baseApiFail},
			args:    args{loginReq: loginReq},
			want:    nil,
			wantErr: true,
		},
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

func TestLoginApi_unmarshal(t *testing.T) {
	httpClient := http.DefaultClient
	parsedUrl, _ := url.Parse("https://test.org:8080")
	baseApi := &BaseApi{HttpClient: httpClient, Url: parsedUrl}

	type fields struct {
		BaseApi *BaseApi
	}
	bearerToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9" +
		".eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ" +
		".SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"
	userId := "27596b04-f260-4bc0-ab02-e437a454ef90"
	userEmail := "hari.seldon@fundation.gal"
	jsonLogin := fmt.Sprintf(`{"token":"%v","user":{"userId":"%v","email":"%v"}}`,
		bearerToken, userId, userEmail)
	readerPass := ioutil.NopCloser(strings.NewReader(jsonLogin))
	userUUID, _ := uuid.Parse(userId)
	expectedLoginResponse := &LoginResponse{
		Token: &bearerToken,
		User: &UserDTO{
			Id:    &userUUID,
			Email: &userEmail,
		},
	}

	readerFail := ioutil.NopCloser(strings.NewReader(`{"invalidKey":"invalid"}`))

	type args struct {
		body io.ReadCloser
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *LoginResponse
		wantErr bool
	}{
		{
			name:    "pass",
			fields:  fields{BaseApi: baseApi},
			args:    args{body: readerPass},
			want:    expectedLoginResponse,
			wantErr: false,
		},
		{
			name:    "fail invalid json",
			fields:  fields{BaseApi: baseApi},
			args:    args{body: readerFail},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			la := &LoginApi{
				BaseApi: tt.fields.BaseApi,
			}
			got, err := la.unmarshal(tt.args.body)
			if (err != nil) != tt.wantErr {
				t.Errorf("unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("unmarshal() got = %v, want %v", got, tt.want)
			}
		})
	}
}
