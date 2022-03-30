package plugin

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io"
)

var (
	getApplicationConf  = map[string]string{"method": "GET", "path": "/api/v1/users/%v/applications"}
	postApplicationConf = map[string]string{"method": "POST", "path": "/api/v1/users/%v/applications"}
)

type Application struct {
	Id   *uuid.UUID `json:"applicationId"`
	Name string     `json:"name"`
}

type CreateApplicationRequest struct {
	Name string `json:"applicationName"`
}

type ApplicationApiInterface interface {
	GetApplications()
	CreateApplication(string)
}

type ApplicationApi struct {
	BaseApi
	ApplicationApiInterface
	user User
}

func (aa *ApplicationApi) GetApplications() ([]*Application, error) {
	method := getApplicationConf["method"]
	// Make a copy to prevent side effects
	urlLogin := aa.url
	urlLogin.Path = fmt.Sprintf(getApplicationConf["path"], aa.user.Id.String())

	req, err := aa.BuildRequest(method, urlLogin.String(), nil)
	if err != nil {
		return nil, aa.getApplicationsError(err)
	}

	resp, err := aa.httpClient.Do(req)
	if err != nil {
		return nil, aa.getApplicationsError(err)
	}
	defer aa.closing(resp.Body)

	if err := aa.checkStatusOrErr(resp, 200); err != nil {
		return nil, aa.getApplicationsError(err)
	}

	if applications, err := aa.unmarshalApplications(resp.Body); err != nil {
		return nil, aa.getApplicationsError(err)
	} else {
		return applications, nil
	}
}

func (aa *ApplicationApi) unmarshalApplications(body io.ReadCloser) ([]*Application, error) {
	bodyBytes, err := aa.toBytes(body)
	if err != nil {
		return nil, err
	}

	var applications []Application
	if err := json.Unmarshal(bodyBytes, &applications); err != nil {
		return nil, err
	}

	applicationsResult := make([]*Application, len(applications))
	for i, application := range applications {
		applicationsResult[i] = &application
	}

	return applicationsResult, nil
}

func (aa *ApplicationApi) getApplicationsError(err error) error {
	return fmt.Errorf("%w; get request to get all applications for user %v failed", err, aa.user)
}

func (aa *ApplicationApi) CreateApplication(name string) (*Application, error) {
	method := postApplicationConf["method"]
	// Make a copy to prevent side effects
	urlLogin := aa.url
	urlLogin.Path = fmt.Sprintf(postApplicationConf["path"], aa.user.Id.String())
	createAppReq := CreateApplicationRequest{Name: name}

	req, err := aa.BuildRequest(method, urlLogin.String(), nil)
	if err != nil {
		return nil, aa.createApplicationError(createAppReq, err)
	}

	resp, err := aa.httpClient.Do(req)
	if err != nil {
		return nil, aa.createApplicationError(createAppReq, err)
	}
	defer aa.closing(resp.Body)

	if err := aa.checkStatusOrErr(resp, 201); err != nil {
		return nil, aa.createApplicationError(createAppReq, err)
	}

	if application, err := aa.unmarshalApplication(resp.Body); err != nil {
		return nil, aa.createApplicationError(createAppReq, err)
	} else {
		return application, nil
	}
}

func (aa *ApplicationApi) unmarshalApplication(body io.ReadCloser) (*Application, error) {
	bodyBytes, err := aa.toBytes(body)
	if err != nil {
		return nil, err
	}

	var application Application
	if err := json.Unmarshal(bodyBytes, &application); err != nil {
		return nil, err
	}

	return &application, nil
}

func (aa *ApplicationApi) createApplicationError(createAppReq CreateApplicationRequest, err error) error {
	return fmt.Errorf("%w; create application %v failed for user %v failed", err, createAppReq.Name, aa.user)
}
