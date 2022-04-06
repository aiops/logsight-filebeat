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

func TestUserApi_Login(t *testing.T) {
	bearerToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9" +
		".eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ" +
		".SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"
	userId := "27596b04-f260-4bc0-ab02-e437a454ef90"
	userUUID, _ := uuid.Parse(userId)
	userEmail := "lenar.hoyt@hyperion.gal"
	userPassword := "hyperion_exile"
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

	expectedUser := &User{
		Id:       userUUID,
		Email:    userEmail,
		Password: userPassword,
	}

	type fields struct {
		LoginApi *LoginApi
	}
	type args struct {
		email    string
		password string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *User
		wantErr bool
	}{
		{
			name:    "pass",
			fields:  fields{LoginApi: &LoginApi{BaseApi: baseApiPass}},
			args:    args{email: userEmail, password: userPassword},
			want:    expectedUser,
			wantErr: false,
		},
		{
			name:    "fail",
			fields:  fields{LoginApi: &LoginApi{BaseApi: baseApiFail}},
			args:    args{email: userEmail, password: userPassword},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UserApi{
				LoginApi: tt.fields.LoginApi,
			}
			got, err := u.Login(tt.args.email, tt.args.password)
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

func TestUser_String(t *testing.T) {
	type fields struct {
		Id       uuid.UUID
		Email    string
		Password string
	}
	uuidRandom, _ := uuid.NewRandom()
	inputFields := fields{
		Id:       uuidRandom,
		Email:    "lenar.hoyt@hyperion.gal",
		Password: "hyperion_exile",
	}
	want := fmt.Sprintf(`{"id": "%v", "email": "%v"}`, uuidRandom.String(), inputFields.Email)
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "pass",
			fields: inputFields,
			want:   want,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := User{
				Id:       tt.fields.Id,
				Email:    tt.fields.Email,
				Password: tt.fields.Password,
			}
			if got := u.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
