package api

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io"
)

var (
	loginConf = map[string]string{"method": "POST", "path": "/api/v1/auth/Login"}
)

type LoginError struct {
	err   error
	email string
}

func (le LoginError) Error() string {
	msg := fmt.Sprintf("login failed for email %v", le.email)
	if le.err != nil {
		return fmt.Sprintf("%v; %v", le.err, msg)
	}
	return msg
}

func (le LoginError) Unwrap() error {
	return le.err
}

type UserDTO struct {
	Id    uuid.UUID `json:"userId"`
	Email uuid.UUID `json:"email"`
}

type LoginResponse struct {
	Token uuid.UUID `json:"token"`
	User  UserDTO   `json:"User"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginApi struct {
	*BaseApi
}

func (la *LoginApi) Login(loginReq LoginRequest) (*LoginResponse, error) {
	method := loginConf["method"]
	// Make a copy to prevent side effects
	urlLogin := la.Url
	urlLogin.Path = loginConf["path"]

	req, err := la.BuildRequest(method, urlLogin.String(), loginReq)
	if err != nil {
		return nil, LoginError{err: err, email: loginReq.Email}
	}

	resp, err := la.HttpClient.Do(req)
	if err != nil {
		return nil, LoginError{err: err, email: loginReq.Email}
	}
	defer la.closing(resp.Body)

	if err := la.CheckStatusOrErr(resp, 200); err != nil {
		return nil, LoginError{err: err, email: loginReq.Email}
	}

	if loginResp, err := la.unmarshal(resp.Body); err != nil {
		return nil, LoginError{err: err, email: loginReq.Email}
	} else {
		return loginResp, nil
	}
}

func (la *LoginApi) unmarshal(body io.ReadCloser) (*LoginResponse, error) {
	bodyBytes, err := la.toBytes(body)
	if err != nil {
		return nil, err
	}

	var loginResp LoginResponse
	if err := json.Unmarshal(bodyBytes, &loginResp); err != nil {
		return nil, err
	}

	return &loginResp, nil
}
