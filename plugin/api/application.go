package api

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io"
	"regexp"
)

const DefaultApplicationName = "filebeat_source"

var (
	getApplicationConf  = map[string]string{"method": "GET", "path": "/api/v1/users/%v/applications"}
	postApplicationConf = map[string]string{"method": "POST", "path": "/api/v1/users/%v/applications"}
)

type Application struct {
	Id   uuid.UUID `json:"applicationId"`
	Name string    `json:"name"`
}

type CreateApplicationRequest struct {
	Name string `json:"applicationName"`
}

type ApplicationApiInterface interface {
	GetApplications() ([]*Application, error)
	GetApplicationByName(string) (*Application, error)
	CreateApplication(CreateApplicationRequest) (*Application, error)
}

type ApplicationApi struct {
	*BaseApi
	ApplicationApiInterface

	User *User
}

func (aa *ApplicationApi) GetApplications() ([]*Application, error) {
	method := getApplicationConf["method"]
	// Make a copy to prevent side effects
	urlLogin := aa.Url
	urlLogin.Path = fmt.Sprintf(getApplicationConf["path"], aa.User.Id.String())

	req, err := aa.BuildRequestWithBasicAuth(method, urlLogin.String(), nil, aa.User.Email, aa.User.Password)
	if err != nil {
		return nil, aa.getApplicationsError(err)
	}

	resp, err := aa.HttpClient.Do(req)
	if err != nil {
		return nil, aa.getApplicationsError(err)
	}
	defer aa.closing(resp.Body)

	if err := aa.CheckStatusOrErr(resp, 200); err != nil {
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
	return fmt.Errorf("%w; get request to get all applications for User %v failed", err, aa.User)
}

// GetApplicationByName retrieves all applications and searches for the given name. If not found, nil is returned.
func (aa *ApplicationApi) GetApplicationByName(name string) (*Application, error) {
	applications, err := aa.GetApplications()
	if err != nil {
		return nil, err
	}

	applicationMap := map[string]*Application{}
	for _, app := range applications {
		if app != nil {
			applicationMap[app.Name] = app
		}
	}

	application, _ := applicationMap[name]
	return application, nil
}

func (aa *ApplicationApi) CreateApplication(createAppReq CreateApplicationRequest) (*Application, error) {
	method := postApplicationConf["method"]
	// Make a copy to prevent side effects
	urlLogin := aa.Url
	urlLogin.Path = fmt.Sprintf(postApplicationConf["path"], aa.User.Id.String())

	req, err := aa.BuildRequestWithBasicAuth(method, urlLogin.String(), nil, aa.User.Email, aa.User.Password)
	if err != nil {
		return nil, aa.createApplicationError(createAppReq, err)
	}

	resp, err := aa.HttpClient.Do(req)
	if err != nil {
		return nil, aa.createApplicationError(createAppReq, err)
	}
	defer aa.closing(resp.Body)

	if err := aa.CheckStatusOrErr(resp, 201); err != nil {
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
	return fmt.Errorf("%w; create application %v failed for User %v failed", err, createAppReq.Name, aa.User)
}

func EscapeSpecialCharsForValidApplicationName(name string) string {
	// Make a Regex to say we only want letters and numbers
	reg := regexp.MustCompile("[^a-z0-9_]+")
	result := reg.ReplaceAllString(name, "")
	if result == "" {
		return DefaultApplicationName
	}
	return result
}
