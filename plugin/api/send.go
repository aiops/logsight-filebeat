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

type MissingApplicationHandler interface {
	handleApplication(string) (*Application, error)
}

type ErrorOnMissingApplication struct {
	MissingApplicationHandler
	ApplicationApi ApplicationApiInterface
}

func (ea ErrorOnMissingApplication) handleApplication(appName string) (*Application, error) {
	application, err := ea.ApplicationApi.GetApplicationByName(appName)
	if err != nil {
		return nil, fmt.Errorf("%w; while handling application %v", err, appName)
	}
	if application == nil {
		return nil, ApplicationNotFoundError{applicationName: appName}
	}
	return application, nil
}

type AutoCreateMissingApplication struct {
	MissingApplicationHandler
	ApplicationApi ApplicationApiInterface
}

func (aca AutoCreateMissingApplication) handleApplication(appName string) (*Application, error) {
	application, err := aca.ApplicationApi.GetApplicationByName(appName)
	if err != nil {
		return nil, fmt.Errorf("%w; while handling application %v", err, appName)
	}
	if application == nil {
		application, err = aca.createMissingApplication(appName)
		if err != nil {
			return nil, fmt.Errorf("%w; error while auto-creating application %v", err, appName)
		}
		return nil, ApplicationNotFoundError{applicationName: appName}
	}
	return application, nil

}

func (aca AutoCreateMissingApplication) createMissingApplication(appName string) (*Application, error) {
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
	application, err := as.MissingAppHandler.handleApplication(logBatch.ApplicationName)
	if err != nil {
		return err
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
