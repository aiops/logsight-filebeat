package tmp

import (
	"fmt"
	"github.com/elastic/beats/v7/libbeat/logp"
	"net/url"
)

type Logsight struct {
	BaseURL  *url.URL
	Email    string
	Password string

	Request Request

	Token string
	key   string
	log    *logp.Logger
}

func (l *Logsight) Login() error {
	urlLogin := *l.BaseURL
	urlLogin.Path =  "/api/auth/login"
	body := map[string]string{"email": l.Email, "password": l.Password}
	status, resp, err  := l.Request.request("POST", &urlLogin, body, nil)

	if err != nil {
		return fmt.Errorf("failed to execute login request: %w", err)
	}

	if status != 200 {
		var errMsg = fmt.Sprintf("server responded with unexpected return code %v", status)
		respStr := resp.AsString()
		if err == nil {
			errMsg += ", response: " + respStr
		}
		return fmt.Errorf(errMsg)
	}
	if resp != nil {
		if err := l.readToken(resp); err != nil {
			return fmt.Errorf("error while reading the bearer token: %w", err)
		}
	} else {
		return fmt.Errorf("response body is nil but must contain the bearer token")
	}

	return nil
}

func (l *Logsight) readToken(resp *ResponseBody) error {
	if respJson, err := resp.AsJson(); err == nil {
		if value, ok := respJson["token"]; ok {
			l.Token = fmt.Sprintf("%v", value)
		} else {
			return fmt.Errorf("response body does not contain 'token' field: %v", respJson)
		}
	} else {
		respStr := resp.AsString()
		return fmt.Errorf("failed to decode response body as JSON %v: %w", respStr, err)
	}
	return nil
}



