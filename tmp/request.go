package tmp

import (
	"encoding/json"
	"fmt"
	"github.com/elastic/beats/v7/libbeat/logp"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

type Request struct {
	Encoder encoder
	Http    *http.Client
	Log     *logp.Logger
}

func (r *Request) request(
	method string,
	url *url.URL,
	body interface{},
	headers map[string]string,
) (int, *ResponseBody, error) {

	if body != nil {
		if err := r.Encoder.Marshal(body); err != nil {
			return 0, nil, fmt.Errorf("failed to encode body (%v): %w", body, err)
		}
	}

	req, err := http.NewRequest(method, url.String(), r.Encoder.Reader())
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		r.Encoder.AddHeader(&req.Header, "")
	}

	return r.execHTTPRequest(req, headers)
}

func (r *Request) execHTTPRequest(req *http.Request, headers map[string]string) (int, *ResponseBody, error) {
	//req.Header.Add("Accept", "application/json")
	for key, value := range headers {
		req.Header.Add(key, value)
	}

	resp, err := r.Http.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to execute request (%v): %w", req, err)
	}
	defer r.closing(resp.Body)

	respBody, err := NewResponseBody(resp.Body)
	if err != nil {
		r.Log.Warnf("error while reading the response body: %v", err)
		return resp.StatusCode, nil, nil
	}

	return resp.StatusCode, respBody, nil
}

func (r *Request) closing(c io.Closer) {
	if c != nil {
		err := c.Close()
		if err != nil {
			r.Log.Warnf("Closing of response body failed: %w", err)
		}
	}
}

type ResponseBody struct {
	respBody []byte
}

func NewResponseBody(respBody io.ReadCloser) (*ResponseBody, error){
	body, err := ioutil.ReadAll(respBody)
	if err != nil {
		return nil, fmt.Errorf("error while reading response body: %w", err)
	}
	return &ResponseBody{body}, nil
}

func (resp *ResponseBody) AsJson() (map[string]interface{}, error) {
	if resp == nil {
		return nil, nil
	}
	var respMap map[string]interface{}
	if err := json.Unmarshal(resp.respBody ,&respMap); err != nil {
		return nil, err
	}
	return respMap, nil
}

func (resp *ResponseBody) AsString() string {
	return string(resp.respBody)
}
