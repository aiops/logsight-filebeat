package logsight

import (
	"fmt"
	"github.com/elastic/beats/v7/libbeat/logp"
	"net/url"
	"strconv"
)

type Logsight struct {
	baseURL  *url.URL
	email    string
	password string
	request  Request
	logger   *logp.Logger

	token   string
	userId  int
	userKey string
}

func (l *Logsight) Init() error {
	err := l.login()
	if err != nil {
		return err
	}
	userInfoResp, err := l.getUserInfo()
	if err != nil {
		return err
	}
	err = l.readUserId(userInfoResp)
	if err != nil {
		return err
	}
	err = l.readUserKey(userInfoResp)
	if err != nil {
		return err
	}
	return nil
}

func (l *Logsight) login() error {
	urlLogin := *l.baseURL
	urlLogin.Path = "/api/auth/login"
	body := map[string]string{"email": l.email, "password": l.password}
	status, resp, err := l.request.request("POST", &urlLogin, body, nil)

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
	if respJson, err := resp.AsJson(); err != nil {
		return err
	} else {
		if userKey, err := l.readStringFieldFromJson("token", respJson); err != nil {
			return err
		} else {
			l.token = userKey
		}
	}
	return nil
}

func (l *Logsight) getUserInfo() (*ResponseBody, error) {
	urlUser := *l.baseURL
	urlUser.Path = "/api/users"

	resp, _, err := l.authRequest("GET", &urlUser, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("get user information request failed: %w", err)
	}

	return resp, nil
}

func (l *Logsight) readUserId(resp *ResponseBody) error {
	if respJson, err := resp.AsJson(); err != nil {
		return err
	} else {
		if userIdStr, err := l.readStringFieldFromJson("id", respJson); err != nil {
			return err
		} else {
			userId, err := strconv.Atoi(userIdStr)
			if err != nil {
				return fmt.Errorf("user id %v is not an integer: %v", userIdStr, err)
			} else {
				l.userId = userId
			}
		}
	}
	return nil
}

func (l *Logsight) readUserKey(resp *ResponseBody) error {
	if respJson, err := resp.AsJson(); err != nil {
		return err
	} else {
		if userKey, err := l.readStringFieldFromJson("key", respJson); err != nil {
			return err
		} else {
			l.userKey = userKey
		}
	}
	return nil
}

func (l *Logsight) SendLogs(logs []string, app string) error {
	urlLogs := *l.baseURL
	urlLogs.Path = fmt.Sprintf("/api/logs/%v/%v/send_logs", l.userKey, app)
	_, _, err := l.authRequest("POST", &urlLogs, logs, map[string]string{"Content-Type": "text/plain"})
	if err != nil {
		return err
	}

	return nil
}

func (l *Logsight) CreateApp(name string) error {
	if appExists, err := l.AppExists(name); err != nil || appExists == false {
		urlLogs := *l.baseURL
		urlLogs.Path = "/api/applications/create"
		body := map[string]string{"name": name, "key": l.userKey}
		_, _, err := l.authRequest("POST", &urlLogs, body, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

func (l *Logsight) GetAppList() ([]string, error) {
	urlApp := *l.baseURL
	urlApp.Path = "/api/applications"
	resp, _, err := l.authRequest("GET", &urlApp, nil, nil)
	if err != nil {
		return nil, err
	}

	respJsonArray, err := resp.AsJsonArray()
	if err != nil {
		return nil, fmt.Errorf("failed to decode response body to JSON %v: %w", resp.AsString(), err)
	}
	var apps []string
	for i := range respJsonArray {
		if app, err := l.readStringFieldFromJson("name", respJsonArray[i]); err == nil {
			apps = append(apps, app)
		}
	}

	return apps, nil
}

func (l *Logsight) AppExists(name string) (bool, error) {
	urlApp := *l.baseURL
	urlApp.Path = fmt.Sprintf("/api/applications/%v", name)
	_, status, err := l.authRequest("GET", &urlApp, nil, nil)
	if err != nil {
		if status == 404 {
			return false, nil
		} else {
			return false, err
		}
	}
	return true, nil
}

func (l *Logsight) authRequest(
	method string,
	url *url.URL,
	body interface{},
	headers map[string]string,
) (*ResponseBody, int, error) {
	// auth requests require a bearer token which is acquired by calling Login()
	if l.token == "" {
		l.logger.Warnf("%v requests to %v require authorization. trying to login...", method, url)
		if err := l.login(); err != nil {
			return nil, -1, fmt.Errorf("authorization via login failed: %w", err)
		}
	}

	// prepare for auth header
	if headers == nil {
		headers = make(map[string]string)
	}
	// create a Bearer string by appending string access token
	bearer := fmt.Sprintf("Bearer %v", l.token)
	// add authorization header
	headers["Authorization"] = bearer

	// try to execute request with auth header, retry if httpClient status is unauthorized (401) or forbidden (403)
	status, resp, err := l.request.request(method, url, body, headers)
	if err != nil {
		return nil, -1, err
	}
	if status == 401 || status == 403 {
		l.logger.Warnf("unauthorized / forbidden return code from logsight with current bearer token. re-trying login...", method, url)
		if err := l.login(); err != nil {
			return nil, status, fmt.Errorf("authorization via login failed: %w", err)
		}
		l.logger.Infof("login successful. re-trying authorized %v request to %v", method, url)
		status, resp, err := l.request.request(method, url, body, headers)
		if err != nil {
			return nil, status, err
		}
		if status != 200 {
			return nil, status, fmt.Errorf("%v request to %v failed due to unexpected httpClient status %v", method, url, status)
		}
		return resp, status, nil
	} else if status != 200 {
		return nil, status, fmt.Errorf("%v request to %v failed due to unexpected httpClient status %v", method, url, status)
	}

	return resp, status, nil
}

func (l *Logsight) readStringFieldFromJson(key string, respJson map[string]interface{}) (string, error) {
	var result string
	if value, ok := respJson[key]; ok {
		result = fmt.Sprintf("%v", value)
	} else {
		return "", fmt.Errorf("response body does not contain '%v' field: %v", key, respJson)
	}
	return result, nil
}

func (l *Logsight) GetUserInfos() map[string]string {
	return map[string]string{"key": l.userKey, "id": fmt.Sprintf("%v", l.userId)}
}
