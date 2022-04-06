package api

import (
	"fmt"
	"github.com/google/uuid"
)

// LogBatch data structure
type LogBatch struct {
	ApplicationName string
	Tag             string
	Logs            []*Log
}

func (lb *LogBatch) ToLogBatchRequest(applicationId uuid.UUID) *LogBatchRequest {
	return &LogBatchRequest{
		ApplicationId: applicationId,
		Tag:           lb.Tag,
		Logs:          lb.Logs,
	}
}

type ApplicationNotFoundError struct {
	applicationName string
}

func (e ApplicationNotFoundError) Error() string {
	return fmt.Sprintf("application %v not found", e.applicationName)
}

func getApplicationByName(appName string, api ApplicationApiInterface) (*Application, error) {
	application, err := api.GetApplicationByName(appName)
	if err != nil {
		return nil, err
	}
	return application, nil
}

type MissingApplicationHandler interface {
	getApplicationByName(string) (*Application, error)
	handleMissingApplication(string) (*Application, error)
}

type ErrorOnMissingApplication struct {
	MissingApplicationHandler
	ApplicationApi ApplicationApiInterface
}

func (ea ErrorOnMissingApplication) getApplicationByName(appName string) (*Application, error) {
	return getApplicationByName(appName, ea.ApplicationApi)
}

func (ea ErrorOnMissingApplication) handleMissingApplication(appName string) (*Application, error) {
	return nil, ApplicationNotFoundError{applicationName: appName}
}

type AutoCreateMissingApplication struct {
	MissingApplicationHandler
	ApplicationApi ApplicationApiInterface
}

func (aca AutoCreateMissingApplication) getApplicationByName(appName string) (*Application, error) {
	return getApplicationByName(appName, aca.ApplicationApi)
}

func (aca AutoCreateMissingApplication) handleMissingApplication(appName string) (*Application, error) {
	validName := EscapeSpecialCharsForValidApplicationName(appName)
	application, err := aca.ApplicationApi.CreateApplication(CreateApplicationRequest{Name: validName})
	if err != nil {
		return nil, err
	}
	return application, nil
}

type LogSender struct {
	LogApi            *LogApi
	MissingAppHandler MissingApplicationHandler
}

func (as LogSender) Send(logBatch *LogBatch) error {
	application, err := as.MissingAppHandler.getApplicationByName(logBatch.ApplicationName)
	if err != nil {
		return err
	}
	if application == nil {
		application, err = as.MissingAppHandler.handleMissingApplication(logBatch.ApplicationName)
		if err != nil {
			return err
		}
	}

	logBatchReq := logBatch.ToLogBatchRequest(*application.Id)
	if _, err := as.LogApi.SendLogBatch(logBatchReq); err != nil {
		return err
	}

	return nil
}

func (as LogSender) Close() {
	as.LogApi.HttpClient.CloseIdleConnections()
}
