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

func TestApplicationApi_CreateApplication(t *testing.T) {
	type fields struct {
		BaseApi *BaseApi
		User    *User
	}
	type args struct {
		createAppReq CreateApplicationRequest
	}
	idStr := "27596b04-f260-4bc0-ab02-e437a454ef90"
	idUUID, _ := uuid.Parse(idStr)
	appName := "heighliner"
	jsonAppValid := []byte(fmt.Sprintf(
		`{"applicationId":"%v","applicationName":"%v"}`, idStr, appName))
	jsonAppInvalid := []byte(fmt.Sprintf(`{"asaa":"%v","as":"%v"}`, idStr, appName))

	// generate a test server, so we can capture and inspect the request
	testServerValid := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
		_, _ = res.Write(jsonAppValid)
	}))
	defer func() { testServerValid.Close() }()
	testServerInvalid := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
		_, _ = res.Write(jsonAppInvalid)
	}))
	defer func() { testServerInvalid.Close() }()
	testServerErr := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusInternalServerError)
		_, _ = res.Write([]byte(`{"message":"failed"}`))
	}))
	defer func() { testServerInvalid.Close() }()

	httpClient := http.DefaultClient
	urlTestServerValid, _ := url.Parse(testServerValid.URL)
	urlTestServerInvalid, _ := url.Parse(testServerInvalid.URL)
	urlTestServerErr, _ := url.Parse(testServerErr.URL)
	baseApiValid := &BaseApi{HttpClient: httpClient, Url: urlTestServerValid}
	baseApiInvalid := &BaseApi{HttpClient: httpClient, Url: urlTestServerInvalid}
	baseApiErr := &BaseApi{HttpClient: httpClient, Url: urlTestServerErr}

	appExpected := &Application{Id: &idUUID, Name: &appName}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Application
		wantErr bool
	}{
		{
			name:    "pass valid resp",
			fields:  fields{BaseApi: baseApiValid, User: &User{}},
			args:    args{createAppReq: CreateApplicationRequest{Name: appName}},
			want:    appExpected,
			wantErr: false,
		},
		{
			name:    "fail invalid resp",
			fields:  fields{BaseApi: baseApiInvalid, User: &User{}},
			args:    args{createAppReq: CreateApplicationRequest{Name: appName}},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "fail err resp",
			fields:  fields{BaseApi: baseApiErr, User: &User{}},
			args:    args{createAppReq: CreateApplicationRequest{Name: appName}},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aa := &ApplicationApi{
				BaseApi: tt.fields.BaseApi,
				User:    tt.fields.User,
			}
			got, err := aa.CreateApplication(tt.args.createAppReq)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateApplication() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateApplication() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestApplicationApi_GetApplicationByName(t *testing.T) {
	type fields struct {
		BaseApi *BaseApi
		User    *User
	}
	type args struct {
		name string
	}

	idStr := "27596b04-f260-4bc0-ab02-e437a454ef90"
	idUUID, _ := uuid.Parse(idStr)
	appName := "heighliner"
	jsonAppValid := []byte(fmt.Sprintf(
		`{"applications":[{"applicationId":"%v","name":"%v"}]}`, idStr, appName))
	jsonAppInvalid := []byte(fmt.Sprintf(`{"asaa":"%v","as":"%v"}`, idStr, appName))

	// generate a test server, so we can capture and inspect the request
	testServerValid := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
		_, _ = res.Write(jsonAppValid)
	}))
	defer func() { testServerValid.Close() }()
	testServerInvalid := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
		_, _ = res.Write(jsonAppInvalid)
	}))
	defer func() { testServerInvalid.Close() }()
	testServerErr := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusNotFound)
		_, _ = res.Write([]byte(`{"message":"failed"}`))
	}))
	defer func() { testServerErr.Close() }()

	httpClient := http.DefaultClient
	urlTestServerValid, _ := url.Parse(testServerValid.URL)
	urlTestServerInvalid, _ := url.Parse(testServerInvalid.URL)
	urlTestServerErr, _ := url.Parse(testServerErr.URL)
	baseApiValid := &BaseApi{HttpClient: httpClient, Url: urlTestServerValid}
	baseApiInvalid := &BaseApi{HttpClient: httpClient, Url: urlTestServerInvalid}
	baseApiErr := &BaseApi{HttpClient: httpClient, Url: urlTestServerErr}

	appExpected := &Application{Id: &idUUID, Name: &appName}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Application
		wantErr bool
	}{
		{
			name:    "pass valid resp",
			fields:  fields{BaseApi: baseApiValid, User: &User{}},
			args:    args{name: appName},
			want:    appExpected,
			wantErr: false,
		},
		{
			name:    "fail invalid resp",
			fields:  fields{BaseApi: baseApiInvalid, User: &User{}},
			args:    args{name: appName},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "fail err resp",
			fields:  fields{BaseApi: baseApiErr, User: &User{}},
			args:    args{name: appName},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aa := &ApplicationApi{
				BaseApi: tt.fields.BaseApi,
				User:    tt.fields.User,
			}
			got, err := aa.GetApplicationByName(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetApplicationByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetApplicationByName() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestApplicationApi_GetApplications(t *testing.T) {
	type fields struct {
		BaseApi *BaseApi
		User    *User
	}
	idUUID1, _ := uuid.NewRandom()
	appName1 := "heighliner1"
	idUUID2, _ := uuid.NewRandom()
	appName2 := "heighliner2"
	jsonAppValid1 := []byte(fmt.Sprintf(
		`{"applications":[{"applicationId":"%v","name":"%v"}]}`, idUUID1.String(), appName1))
	jsonAppValid2 := []byte(fmt.Sprintf(
		`{"applications":[{"applicationId":"%v","name":"%v"},{"applicationId":"%v","name":"%v"}]}`,
		idUUID1.String(), appName1, idUUID2.String(), appName2))
	jsonAppValidEmpty := []byte(fmt.Sprintf(`{"applications":[]}`))
	jsonAppInvalid := []byte(fmt.Sprintf(`{"asaa":"12","as":"1221"}`))

	// generate a test server, so we can capture and inspect the request
	testServerValid1 := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
		_, _ = res.Write(jsonAppValid1)
	}))
	defer func() { testServerValid1.Close() }()
	testServerValid2 := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
		_, _ = res.Write(jsonAppValid2)
	}))
	defer func() { testServerValid2.Close() }()
	testServerValidEmpty := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
		_, _ = res.Write(jsonAppValidEmpty)
	}))
	defer func() { testServerValidEmpty.Close() }()
	testServerInvalid := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
		_, _ = res.Write(jsonAppInvalid)
	}))
	defer func() { testServerInvalid.Close() }()
	testServerErr := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusNotFound)
		_, _ = res.Write([]byte(`{"message":"failed"}`))
	}))
	defer func() { testServerInvalid.Close() }()

	httpClient := http.DefaultClient
	urlTestServerValid1, _ := url.Parse(testServerValid1.URL)
	urlTestServerValid2, _ := url.Parse(testServerValid2.URL)
	urlTestServerValidEmpty, _ := url.Parse(testServerValidEmpty.URL)
	urlTestServerInvalid, _ := url.Parse(testServerInvalid.URL)
	urlTestServerErr, _ := url.Parse(testServerErr.URL)
	baseApiValid1 := &BaseApi{HttpClient: httpClient, Url: urlTestServerValid1}
	baseApiValid2 := &BaseApi{HttpClient: httpClient, Url: urlTestServerValid2}
	baseApiValidEmpty := &BaseApi{HttpClient: httpClient, Url: urlTestServerValidEmpty}
	baseApiInvalid := &BaseApi{HttpClient: httpClient, Url: urlTestServerInvalid}
	baseApiErr := &BaseApi{HttpClient: httpClient, Url: urlTestServerErr}

	appsExpected1 := []*Application{
		{Id: &idUUID1, Name: &appName1},
	}
	appsExpected2 := []*Application{
		{Id: &idUUID1, Name: &appName1},
		{Id: &idUUID2, Name: &appName2},
	}

	tests := []struct {
		name    string
		fields  fields
		want    []*Application
		wantErr bool
	}{
		{
			name:    "pass 1",
			fields:  fields{BaseApi: baseApiValid1, User: &User{}},
			want:    appsExpected1,
			wantErr: false,
		},
		{
			name:    "pass 2",
			fields:  fields{BaseApi: baseApiValid2, User: &User{}},
			want:    appsExpected2,
			wantErr: false,
		},
		{
			name:    "pass empty",
			fields:  fields{BaseApi: baseApiValidEmpty, User: &User{}},
			want:    []*Application{},
			wantErr: false,
		},
		{
			name:    "fail invalid",
			fields:  fields{BaseApi: baseApiInvalid, User: &User{}},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "fail err",
			fields:  fields{BaseApi: baseApiErr, User: &User{}},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aa := &ApplicationApi{
				BaseApi: tt.fields.BaseApi,
				User:    tt.fields.User,
			}
			got, err := aa.GetApplications()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetApplications() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetApplications() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestApplicationApi_createApplicationError(t *testing.T) {
	type fields struct {
		BaseApi *BaseApi
		User    *User
	}
	type args struct {
		createAppReq CreateApplicationRequest
		err          error
	}
	appName := "test"
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name:   "pass",
			fields: fields{BaseApi: nil, User: nil},
			args:   args{createAppReq: CreateApplicationRequest{Name: appName}, err: nil},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aa := &ApplicationApi{
				BaseApi: tt.fields.BaseApi,
				User:    tt.fields.User,
			}
			err := aa.createApplicationError(tt.args.createAppReq, tt.args.err)
			if !strings.Contains(err.Error(), appName) {
				t.Errorf("createApplicationError() must contain %v", appName)
			}
		})
	}
}

func TestApplicationApi_getApplicationsError(t *testing.T) {
	type fields struct {
		BaseApi *BaseApi
		User    *User
	}
	type args struct {
		err error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name:   "pass",
			fields: fields{BaseApi: nil, User: nil},
			args:   args{err: fmt.Errorf("error")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aa := &ApplicationApi{
				BaseApi: tt.fields.BaseApi,
				User:    tt.fields.User,
			}
			if err := aa.getApplicationsError(tt.args.err); err == nil {
				t.Errorf("getApplicationsError() error must not be nil")
			}
		})
	}
}

func TestApplicationApi_unmarshalApplication(t *testing.T) {
	type fields struct {
		BaseApi *BaseApi
		User    *User
	}
	type args struct {
		body io.ReadCloser
	}

	idUUID, _ := uuid.NewRandom()
	appName := "nebukadnezar"
	jsonApp := fmt.Sprintf(`{"applicationId":"%v","applicationName":"%v"}`, idUUID.String(), appName)
	readerPass := ioutil.NopCloser(strings.NewReader(jsonApp))
	expectedApp := &Application{Id: &idUUID, Name: &appName}

	readerFail := ioutil.NopCloser(strings.NewReader(`{"invalidKey":"invalid"}"`))

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Application
		wantErr bool
	}{
		{
			name:    "pass",
			fields:  fields{BaseApi: nil, User: nil},
			args:    args{body: readerPass},
			want:    expectedApp,
			wantErr: false,
		},
		{
			name:    "fail",
			fields:  fields{BaseApi: nil, User: nil},
			args:    args{body: readerFail},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aa := &ApplicationApi{
				BaseApi: tt.fields.BaseApi,
				User:    tt.fields.User,
			}
			got, err := aa.unmarshalApplication(tt.args.body)
			if (err != nil) != tt.wantErr {
				t.Errorf("unmarshalApplication() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("unmarshalApplication() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestApplicationApi_unmarshalApplications(t *testing.T) {
	type fields struct {
		BaseApi *BaseApi
		User    *User
	}
	type args struct {
		body io.ReadCloser
	}

	idUUID1, _ := uuid.NewRandom()
	appName1 := "nebukadnezar"
	jsonApp1 := fmt.Sprintf(`{"applications":[{"applicationId":"%v","name":"%v"}]}`, idUUID1.String(), appName1)
	readerPass1 := ioutil.NopCloser(strings.NewReader(jsonApp1))
	expectedApp1 := []*Application{{Id: &idUUID1, Name: &appName1}}

	idUUID2, _ := uuid.NewRandom()
	appName2 := "nebukadnezar"
	jsonApp2 := fmt.Sprintf(
		`{"applications":[{"applicationId":"%v","name":"%v"},{"applicationId":"%v","name":"%v"}]}`,
		idUUID1.String(), appName1, idUUID2.String(), appName2)
	readerPass2 := ioutil.NopCloser(strings.NewReader(jsonApp2))
	expectedApp2 := []*Application{{Id: &idUUID1, Name: &appName1}, {Id: &idUUID2, Name: &appName2}}

	jsonAppEmpty := fmt.Sprintf(`{"applications":[]}`)
	readerPassEmpty := ioutil.NopCloser(strings.NewReader(jsonAppEmpty))

	readerFail := ioutil.NopCloser(strings.NewReader(`{"invalidKey":"invalid"}"`))

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*Application
		wantErr bool
	}{
		{
			name:    "pass one",
			fields:  fields{BaseApi: nil, User: nil},
			args:    args{body: readerPass1},
			want:    expectedApp1,
			wantErr: false,
		},
		{
			name:    "pass two",
			fields:  fields{BaseApi: nil, User: nil},
			args:    args{body: readerPass2},
			want:    expectedApp2,
			wantErr: false,
		},
		{
			name:    "pass empty",
			fields:  fields{BaseApi: nil, User: nil},
			args:    args{body: readerPassEmpty},
			want:    []*Application{},
			wantErr: false,
		},
		{
			name:    "fail",
			fields:  fields{BaseApi: nil, User: nil},
			args:    args{body: readerFail},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aa := &ApplicationApi{
				BaseApi: tt.fields.BaseApi,
				User:    tt.fields.User,
			}
			got, err := aa.unmarshalApplicationsResponse(tt.args.body)
			if (err != nil) != tt.wantErr {
				t.Errorf("unmarshalApplicationsResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("unmarshalApplicationsResponse() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEscapeSpecialCharsForValidname(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"valid app name app", args{name: "app"}, "app"},
		{"valid app name 123", args{name: "123"}, "123"},
		{"valid app name app_12", args{name: "app_12"}, "app_12"},
		{"valid app name empty", args{name: ""}, DefaultApplicationName},
		{"valid app name +++", args{name: "+++"}, DefaultApplicationName},
		{"valid app name +a+a+", args{name: "+a+a+"}, "aa"},
		{"valid app name b+a+a+b", args{name: "b+a+a+b"}, "baab"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EscapeSpecialCharsForValidApplicationName(tt.args.name); got != tt.want {
				t.Errorf("EscapeSpecialCharsForValidApplicationName() = %v, want %v", got, tt.want)
			}
		})
	}
}
