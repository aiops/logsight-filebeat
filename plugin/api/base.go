package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

type BaseApi struct {
	HttpClient *http.Client
	Url        *url.URL
}

func (ba *BaseApi) BuildRequest(method string, url string, body interface{}) (*http.Request, error) {
	reqEncoded, err := ba.buildBody(body)
	if err != nil {
		return nil, err
	}

	// reqEncoded can be nil for requests without a body
	reqResult, err := http.NewRequest(method, url, reqEncoded)
	if err != nil {
		return nil, err
	}
	ba.addJsonContentTypeHeader(reqResult)

	return reqResult, nil
}

func (ba *BaseApi) BuildRequestWithBasicAuth(method string, url string, body interface{}, username string, password string) (*http.Request, error) {
	req, err := ba.BuildRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(username, password)
	return req, nil
}

func (ba *BaseApi) addJsonContentTypeHeader(req *http.Request) {
	key := "Content-Type"
	value := "application/json; charset=UTF-8"
	req.Header.Add(key, value)
}

func (ba *BaseApi) closing(c io.Closer) {
	if c != nil {
		_ = c.Close()
	}
}

func (ba *BaseApi) toBytes(respBody io.ReadCloser) ([]byte, error) {
	if respBody == nil {
		return nil, nil
	}
	bodyBytes, err := ioutil.ReadAll(respBody)
	if err != nil {
		return nil, fmt.Errorf("%w; error while reading response body", err)
	}
	return bodyBytes, nil
}

func (ba *BaseApi) encode(body interface{}) (io.Reader, error) {
	bodyEnc := bytes.NewBuffer(nil)
	enc := json.NewEncoder(bodyEnc)
	if err := enc.Encode(body); err != nil {
		return nil, err
	}
	return bodyEnc, nil
}

func (ba *BaseApi) buildBody(body interface{}) (io.Reader, error) {
	if body == nil {
		return nil, nil
	}
	bodyEnc, err := ba.encode(body)
	if err != nil {
		return nil, err
	}
	return bodyEnc, nil
}

func (ba *BaseApi) CheckStatusOrErr(resp *http.Response, expectedStatus int) error {
	if resp.StatusCode != expectedStatus {
		if respBytes, err := ba.toBytes(resp.Body); err != nil {
			return fmt.Errorf("unexpected return code %v. %v was expected", resp.StatusCode, expectedStatus)
		} else {
			return fmt.Errorf("unexpected return code %v. %v was expected. error body: %v",
				resp.StatusCode, expectedStatus, string(respBytes))
		}
	}
	return nil
}
