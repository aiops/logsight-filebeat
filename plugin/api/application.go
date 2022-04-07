package api

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io"
	"net/http"
	"regexp"
)

const DefaultApplicationName = "filebeat_source"

var (
	getApplicationConf  = map[string]string{"method": "GET", "path": "/api/v1/users/%v/applications"}
	postApplicationConf = map[string]string{"method": "POST", "path": "/api/v1/users/%v/applications"}
)

type Application struct {
	Id   *uuid.UUID `json:"applicationId"`
	Name *string    `json:"name"`
}

type applicationsResponse struct {
	Applications []*Application `json:"applications"`
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
	ApplicationApiInterface
	*BaseApi

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

	if resp.StatusCode != http.StatusOK {
		return nil, aa.getApplicationsError(aa.GetUnexpectedStatusError(resp, http.StatusOK))
	}

	if applications, err := aa.unmarshalApplicationsResponse(resp.Body); err != nil {
		return nil, aa.getApplicationsError(err)
	} else {
		return applications, nil
	}
}

func (aa *ApplicationApi) unmarshalApplicationsResponse(body io.ReadCloser) ([]*Application, error) {
	bodyBytes, err := aa.toBytes(body)
	if err != nil {
		return nil, err
	}

	var applicationResponse applicationsResponse
	if err := json.Unmarshal(bodyBytes, &applicationResponse); err != nil {
		return nil, fmt.Errorf("%w; failed to unmarshal %v", err, string(bodyBytes))
	}

	if applicationResponse.Applications == nil {
		return nil, fmt.Errorf("failed to parse applications response %v", string(bodyBytes))
	}

	applicationsResult := []*Application{}
	for _, application := range applicationResponse.Applications {
		if application != nil && application.Name != nil && application.Id != nil {
			applicationsResult = append(applicationsResult, application)
		} else {
			break
		}
	}

	if len(applicationsResult) != len(applicationResponse.Applications) {
		return nil, fmt.Errorf("failed to unmarshal applications from json %v", string(bodyBytes))
	} else {
		return applicationsResult, nil
	}
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
			applicationMap[*app.Name] = app
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

	req, err := aa.BuildRequestWithBasicAuth(method, urlLogin.String(), createAppReq, aa.User.Email, aa.User.Password)
	if err != nil {
		return nil, aa.createApplicationError(createAppReq, err)
	}

	resp, err := aa.HttpClient.Do(req)
	if err != nil {
		return nil, aa.createApplicationError(createAppReq, err)
	}
	defer aa.closing(resp.Body)

	if resp.StatusCode == http.StatusConflict {
		application, err := aa.GetApplicationByName(createAppReq.Name)
		if err != nil || application == nil {
			return nil, fmt.Errorf("%w; logsight reports that application %v already exists but failed to load it",
				aa.createApplicationError(createAppReq, err), createAppReq.Name)
		}
		return application, nil
	} else if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, aa.createApplicationError(createAppReq, aa.BaseApi.GetUnexpectedStatusError(resp, http.StatusCreated))
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

	errMsg := fmt.Sprintf("unmarshalling application from %v failed", bodyBytes)
	if application.Name == nil {
		return nil, fmt.Errorf("%v; application name is nil", errMsg)
	}
	if application.Id == nil {
		return nil, fmt.Errorf("%v; application id is nil", errMsg)
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
