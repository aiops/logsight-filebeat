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

func TestNewApplicationApiCacheProxy(t *testing.T) {
	type args struct {
		applicationAPI *ApplicationApi
	}
	urlLocalhost, _ := url.Parse("http://localhost")
	baseApi := &BaseApi{HttpClient: http.DefaultClient, Url: urlLocalhost}
	applicationApi := &ApplicationApi{BaseApi: baseApi, User: &User{}}
	want := &applicationApiCacheProxy{
		applicationAPI:   applicationApi,
		applicationCache: NewApplicationCache(),
	}

	tests := []struct {
		name string
		args args
		want *applicationApiCacheProxy
	}{
		{
			name: "pass",
			args: args{applicationAPI: applicationApi},
			want: want,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewApplicationApiCacheProxy(tt.args.applicationAPI); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewApplicationApiCacheProxy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_applicationApiCacheProxy_ClearCache(t *testing.T) {
	type fields struct {
		applicationAPI   *ApplicationApi
		applicationCache *applicationCache
	}

	urlLocalhost, _ := url.Parse("http://localhost")
	baseApi := &BaseApi{HttpClient: http.DefaultClient, Url: urlLocalhost}
	applicationApi := &ApplicationApi{BaseApi: baseApi, User: &User{}}
	emptyCache := NewApplicationCache()
	nonEmptyCache := NewApplicationCache()
	appName := "test"
	nonEmptyCache.add(&Application{Name: &appName})

	tests := []struct {
		name   string
		fields fields
	}{
		{
			name:   "pass empty",
			fields: fields{applicationAPI: applicationApi, applicationCache: emptyCache},
		},
		{
			name:   "pass non-empty",
			fields: fields{applicationAPI: applicationApi, applicationCache: nonEmptyCache},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capt := &applicationApiCacheProxy{
				applicationAPI:   tt.fields.applicationAPI,
				applicationCache: tt.fields.applicationCache,
			}
			capt.ClearCache()
			if !capt.applicationCache.isEmpty() {
				t.Errorf("ClearCache() must result in an empty cache")
				return
			}
		})
	}
}

func Test_applicationApiCacheProxy_CreateApplication(t *testing.T) {
	type fields struct {
		applicationAPI   *ApplicationApi
		applicationCache *applicationCache
	}
	type args struct {
		req CreateApplicationRequest
	}
	idUUID, _ := uuid.NewRandom()
	idStr := idUUID.String()
	appName := "heighliner"
	jsonApps := []byte(fmt.Sprintf(
		`{"applications":[{"applicationId":"%v","name":"%v"}]}`, idStr, appName))
	jsonAppsEmpty := []byte(fmt.Sprintf(`{"applications":[]}`))
	jsonApp := []byte(fmt.Sprintf(
		`{"applicationId":"%v","applicationName":"%v"}`, idStr, appName))

	testServerAppsValid := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
		_, _ = res.Write(jsonApps)
	}))
	defer func() { testServerAppsValid.Close() }()
	testServerAppsEmptyApp := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
		if req.Method == http.MethodGet {
			_, _ = res.Write(jsonAppsEmpty)
		} else {
			_, _ = res.Write(jsonApp)
		}
	}))
	defer func() { testServerAppsEmptyApp.Close() }()
	testServerErr := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusInternalServerError)
		_, _ = res.Write([]byte(`{"message":"failed"}`))
	}))
	defer func() { testServerErr.Close() }()

	httpClient := http.DefaultClient
	urlTestServerAppsValid, _ := url.Parse(testServerAppsValid.URL)
	urlTestServerAppsEmptyApp, _ := url.Parse(testServerAppsEmptyApp.URL)
	urlTestServerErr, _ := url.Parse(testServerErr.URL)
	baseApiAppValid := &BaseApi{HttpClient: httpClient, Url: urlTestServerAppsValid}
	baseApiAppsEmptyApp := &BaseApi{HttpClient: httpClient, Url: urlTestServerAppsEmptyApp}
	applicationAPIAppValid := &ApplicationApi{BaseApi: baseApiAppValid, User: &User{}}
	applicationAPIAppsEmptyApp := &ApplicationApi{BaseApi: baseApiAppsEmptyApp, User: &User{}}
	baseApiErr := &BaseApi{HttpClient: httpClient, Url: urlTestServerErr}
	applicationAPIErr := &ApplicationApi{BaseApi: baseApiErr, User: &User{}}

	appExpected := &Application{Id: &idUUID, Name: &appName}

	appCacheFilled := NewApplicationCache()
	appCacheFilled.add(appExpected)

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Application
		wantErr bool
	}{
		{
			name:    "pass created",
			fields:  fields{applicationAPI: applicationAPIAppValid, applicationCache: NewApplicationCache()},
			args:    args{req: CreateApplicationRequest{Name: appName}},
			want:    appExpected,
			wantErr: false,
		},
		{
			name:    "pass create",
			fields:  fields{applicationAPI: applicationAPIAppsEmptyApp, applicationCache: NewApplicationCache()},
			args:    args{req: CreateApplicationRequest{Name: appName}},
			want:    appExpected,
			wantErr: false,
		},
		{
			name:    "pass already created cached",
			fields:  fields{applicationAPI: nil, applicationCache: appCacheFilled}, // no http request should be needed
			args:    args{req: CreateApplicationRequest{Name: appName}},
			want:    appExpected,
			wantErr: false,
		},
		{
			name:    "fail",
			fields:  fields{applicationAPI: applicationAPIErr, applicationCache: NewApplicationCache()},
			args:    args{req: CreateApplicationRequest{Name: appName}},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capt := &applicationApiCacheProxy{
				applicationAPI:   tt.fields.applicationAPI,
				applicationCache: tt.fields.applicationCache,
			}
			got, err := capt.CreateApplication(tt.args.req)
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

func Test_applicationApiCacheProxy_GetApplicationByName(t *testing.T) {
	type fields struct {
		applicationAPI   *ApplicationApi
		applicationCache *applicationCache
	}
	type args struct {
		name string
	}
	idUUID, _ := uuid.NewRandom()
	idStr := idUUID.String()
	appName := "heighliner"
	jsonApp := []byte(fmt.Sprintf(
		`{"applications":[{"applicationId":"%v","name":"%v"}]}`, idStr, appName))

	testServerValid := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
		_, _ = res.Write(jsonApp)
	}))
	defer func() { testServerValid.Close() }()
	testServerErr := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusInternalServerError)
		_, _ = res.Write([]byte(`{"message":"failed"}`))
	}))
	defer func() { testServerErr.Close() }()

	httpClient := http.DefaultClient
	urlTestServerValid, _ := url.Parse(testServerValid.URL)
	urlTestServerErr, _ := url.Parse(testServerErr.URL)
	baseApiValid := &BaseApi{HttpClient: httpClient, Url: urlTestServerValid}
	applicationAPIValid := &ApplicationApi{BaseApi: baseApiValid, User: &User{}}
	baseApiErr := &BaseApi{HttpClient: httpClient, Url: urlTestServerErr}
	applicationAPIErr := &ApplicationApi{BaseApi: baseApiErr, User: &User{}}

	appExpected := &Application{Id: &idUUID, Name: &appName}

	appCacheFilled := NewApplicationCache()
	appCacheFilled.add(appExpected)

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Application
		wantErr bool
	}{
		{
			name:    "pass not cached",
			fields:  fields{applicationAPI: applicationAPIValid, applicationCache: NewApplicationCache()},
			args:    args{name: appName},
			want:    appExpected,
			wantErr: false,
		},
		{
			name:    "pass cached",
			fields:  fields{applicationAPI: applicationAPIValid, applicationCache: appCacheFilled},
			args:    args{name: appName},
			want:    appExpected,
			wantErr: false,
		},
		{
			name:    "fail",
			fields:  fields{applicationAPI: applicationAPIErr, applicationCache: NewApplicationCache()},
			args:    args{name: appName},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capt := &applicationApiCacheProxy{
				applicationAPI:   tt.fields.applicationAPI,
				applicationCache: tt.fields.applicationCache,
			}
			got, err := capt.GetApplicationByName(tt.args.name)
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

func Test_applicationApiCacheProxy_GetApplications(t *testing.T) {
	type fields struct {
		applicationAPI   *ApplicationApi
		applicationCache *applicationCache
	}

	idUUID1, _ := uuid.NewRandom()
	idStr1 := idUUID1.String()
	appName1 := "heighliner1"
	jsonApp := []byte(fmt.Sprintf(
		`{"applications":[{"applicationId":"%v","name":"%v"}]}`, idStr1, appName1))
	idUUID2 := idUUID1
	appName2 := "heighliner2"

	testServerValid := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
		_, _ = res.Write(jsonApp)
	}))
	defer func() { testServerValid.Close() }()
	testServerErr := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusInternalServerError)
		_, _ = res.Write([]byte(`{"message":"failed"}`))
	}))
	defer func() { testServerErr.Close() }()

	httpClient := http.DefaultClient
	urlTestServerValid, _ := url.Parse(testServerValid.URL)
	urlTestServerErr, _ := url.Parse(testServerErr.URL)
	baseApiValid := &BaseApi{HttpClient: httpClient, Url: urlTestServerValid}
	applicationAPIValid := &ApplicationApi{BaseApi: baseApiValid, User: &User{}}
	baseApiErr := &BaseApi{HttpClient: httpClient, Url: urlTestServerErr}
	applicationAPIErr := &ApplicationApi{BaseApi: baseApiErr, User: &User{}}

	app1 := &Application{Id: &idUUID1, Name: &appName1}
	app2 := &Application{Id: &idUUID2, Name: &appName2}

	appCacheFilled1 := NewApplicationCache()
	appCacheFilled1.add(app1)

	appCacheFilled2 := NewApplicationCache()
	appCacheFilled2.add(app1)
	appCacheFilled2.add(app2)

	tests := []struct {
		name    string
		fields  fields
		want    []*Application
		wantErr bool
	}{
		{
			name:    "pass cached 1",
			fields:  fields{applicationAPI: applicationAPIValid, applicationCache: appCacheFilled1},
			want:    []*Application{app1},
			wantErr: false,
		},
		{
			name:    "pass cached 2",
			fields:  fields{applicationAPI: applicationAPIValid, applicationCache: appCacheFilled2},
			want:    []*Application{app1, app2},
			wantErr: false,
		},
		{
			name:    "pass not cached",
			fields:  fields{applicationAPI: applicationAPIValid, applicationCache: NewApplicationCache()},
			want:    []*Application{app1},
			wantErr: false,
		},
		{
			name:    "fail",
			fields:  fields{applicationAPI: applicationAPIErr, applicationCache: NewApplicationCache()},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capt := &applicationApiCacheProxy{
				applicationAPI:   tt.fields.applicationAPI,
				applicationCache: tt.fields.applicationCache,
			}
			got, err := capt.GetApplications()
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
