package api

import (
	"bytes"
	"encoding/base64"
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

func (ba *BaseApi) BuildRequest(method string, url string, req interface{}) (*http.Request, error) {
	reqEncoded, err := ba.buildBody(req)
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

func (ba *BaseApi) buildBasicAuth(username string, password string) string {
	auth := fmt.Sprintf("%v:%v", username, password)
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func (ba *BaseApi) closing(c io.Closer) {
	if c != nil {
		_ = c.Close()
	}
}

func (ba *BaseApi) toBytes(respBody io.ReadCloser) ([]byte, error) {
	bodyBytes, err := ioutil.ReadAll(respBody)
	if err != nil {
		return nil, fmt.Errorf("%w; error while reading response body", err)
	}
	return bodyBytes, nil
}

func (ba *BaseApi) marshal(obj interface{}) ([]byte, error) {
	marshaledReqBody, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	return marshaledReqBody, nil
}

func (ba *BaseApi) encode(marshalled []byte) (io.Reader, error) {
	body := bytes.NewBuffer(nil)
	enc := json.NewEncoder(body)
	if err := enc.Encode(marshalled); err != nil {
		return nil, err
	}
	return body, nil
}

func (ba *BaseApi) buildBody(req interface{}) (io.Reader, error) {
	if req == nil {
		return nil, nil
	}

	reqBytes, err := ba.marshal(req)
	if err != nil {
		return nil, err
	}

	reqEncoded, err := ba.encode(reqBytes)
	if err != nil {
		return nil, err
	}

	return reqEncoded, nil
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
