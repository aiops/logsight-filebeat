package api

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io"
	"net/http"
)

var (
	loginConf = map[string]string{"method": "POST", "path": "/api/v1/auth/login"}
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
	Id    *uuid.UUID `json:"userId" validate:"required"`
	Email *string    `json:"email" validate:"required"`
}

type LoginResponse struct {
	Token *string  `json:"token" validate:"required"`
	User  *UserDTO `json:"user" validate:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
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

	if resp.StatusCode != http.StatusOK {
		return nil, LoginError{err: la.GetUnexpectedStatusError(resp, http.StatusOK), email: loginReq.Email}
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
		return nil, fmt.Errorf("%w; error while unmarshalling the Login object %v", err, string(bodyBytes))
	}

	errMsg := fmt.Sprintf("unmarshalling login response from %v failed", bodyBytes)
	if loginResp.Token == nil {
		return nil, fmt.Errorf("%v; token is nil", errMsg)
	}
	if loginResp.User == nil {
		return nil, fmt.Errorf("%v; user is nil", errMsg)
	}
	if loginResp.User.Id == nil {
		return nil, fmt.Errorf("%v; user id is nil", errMsg)
	}
	if loginResp.User.Email == nil {
		return nil, fmt.Errorf("%v; user email is nil", errMsg)
	}

	return &loginResp, nil
}
