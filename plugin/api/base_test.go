package api

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func TestBaseApi_BuildRequest(t *testing.T) {
	type Test struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	testObject := Test{Name: "Hari Seldon", Age: 79}

	type fields struct {
		HttpClient *http.Client
		Url        *url.URL
	}
	httpClient := http.DefaultClient
	parsedUrl, _ := url.Parse("https://test.org:8080")

	key := "Content-Type"
	value := "application/json; charset=UTF-8"
	reqNilBody, _ := http.NewRequest(http.MethodGet, parsedUrl.String(), nil)
	reqNilBody.Header.Add(key, value)

	bodyEnc := bytes.NewBuffer(nil)
	enc := json.NewEncoder(bodyEnc)
	_ = enc.Encode(testObject)
	reqBody, _ := http.NewRequest(http.MethodPost, parsedUrl.String(), bodyEnc)
	reqBody.Header.Add(key, value)

	type args struct {
		method string
		url    string
		body   interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *http.Request
		wantErr bool
	}{
		{
			name:    "pass nil body",
			fields:  fields{HttpClient: httpClient, Url: parsedUrl},
			args:    args{method: http.MethodGet, url: parsedUrl.String(), body: nil},
			want:    reqNilBody,
			wantErr: false,
		},
		{
			name:    "pass body",
			fields:  fields{HttpClient: httpClient, Url: parsedUrl},
			args:    args{method: http.MethodPost, url: parsedUrl.String(), body: testObject},
			want:    reqBody,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ba := &BaseApi{
				HttpClient: tt.fields.HttpClient,
				Url:        tt.fields.Url,
			}
			got, err := ba.BuildRequest(tt.args.method, tt.args.url, tt.args.body)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.Body, tt.want.Body) {
				t.Errorf("BuildRequest() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBaseApi_CheckStatusOrErr(t *testing.T) {
	type fields struct {
		HttpClient *http.Client
		Url        *url.URL
	}
	httpClient := http.DefaultClient
	parsedUrl, _ := url.Parse("https://test.org:8080")
	resp := http.Response{
		Status:           "",
		StatusCode:       200,
		Proto:            "",
		ProtoMajor:       0,
		ProtoMinor:       0,
		Header:           nil,
		Body:             nil,
		ContentLength:    0,
		TransferEncoding: nil,
		Close:            false,
		Uncompressed:     false,
		Trailer:          nil,
		Request:          nil,
		TLS:              nil,
	}

	type args struct {
		resp           *http.Response
		expectedStatus int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "pass",
			fields:  fields{HttpClient: httpClient, Url: parsedUrl},
			args:    args{resp: &resp, expectedStatus: 200},
			wantErr: false,
		},
		{
			name:    "fail",
			fields:  fields{HttpClient: httpClient, Url: parsedUrl},
			args:    args{resp: &resp, expectedStatus: 404},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ba := &BaseApi{
				HttpClient: tt.fields.HttpClient,
				Url:        tt.fields.Url,
			}
			if err := ba.CheckStatusOrErr(tt.args.resp, tt.args.expectedStatus); (err != nil) != tt.wantErr {
				t.Errorf("CheckStatusOrErr() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBaseApi_encode(t *testing.T) {
	type Test struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	testObject := Test{Name: "Hari Seldon", Age: 79}
	testObjectMarshalled := []byte(`{"name":"Hari Seldon","age":79}`)
	testObjectMarshalled = append(testObjectMarshalled, 10)

	type fields struct {
		HttpClient *http.Client
		Url        *url.URL
	}
	httpClient := http.DefaultClient
	parsedUrl, _ := url.Parse("https://test.org:8080")

	type args struct {
		body interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name:    "pass",
			fields:  fields{httpClient, parsedUrl},
			args:    args{testObject},
			want:    testObjectMarshalled,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ba := &BaseApi{
				HttpClient: tt.fields.HttpClient,
				Url:        tt.fields.Url,
			}
			got, err := ba.encode(tt.args.body)
			gotBytes, err := ioutil.ReadAll(got)
			if (err != nil) != tt.wantErr {
				t.Errorf("encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotBytes, tt.want) {
				t.Errorf("encode() got = %v, want %v", gotBytes, tt.want)
			}
		})
	}
}

func TestBaseApi_toBytes(t *testing.T) {
	httpClient := http.DefaultClient
	parsedUrl, _ := url.Parse("https://test.org:8080")
	reader := ioutil.NopCloser(strings.NewReader(`{"name":"Hari Seldon","age":79}`))
	expected := []byte(`{"name":"Hari Seldon","age":79}`)

	type fields struct {
		HttpClient *http.Client
		Url        *url.URL
	}
	type args struct {
		respBody io.ReadCloser
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name:    "pass",
			fields:  fields{httpClient, parsedUrl},
			args:    args{reader},
			want:    expected,
			wantErr: false,
		},
		{
			name:    "pass",
			fields:  fields{httpClient, parsedUrl},
			args:    args{nil},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ba := &BaseApi{
				HttpClient: tt.fields.HttpClient,
				Url:        tt.fields.Url,
			}
			got, err := ba.toBytes(tt.args.respBody)
			if (err != nil) != tt.wantErr {
				t.Errorf("toBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toBytes() got = %v, want %v", got, tt.want)
			}
		})
	}
}
