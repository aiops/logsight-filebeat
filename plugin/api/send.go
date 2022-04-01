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

type SenderInterface interface {
	Close()
	Send(*LogBatch) error
	handleMissingApplication(string) error
}

type Sender struct {
	LogApi         *LogApi
	ApplicationApi ApplicationApiInterface
}

func (as Sender) Send(logBatch *LogBatch) error {
	application, err := as.ApplicationApi.GetApplicationByName(logBatch.ApplicationName)
	if err != nil {
		return err
	}
	if application == nil {
		if err := as.handleMissingApplication(logBatch.ApplicationName); err != nil {
			return err
		}
	}

	logBatchReq := logBatch.ToLogBatchRequest(application.Id)
	if _, err := as.LogApi.SendLogBatch(logBatchReq); err != nil {
		return err
	}

	return nil
}

func (as Sender) Close() {
	as.LogApi.HttpClient.CloseIdleConnections()
}

func (as Sender) handleMissingApplication(name string) error {
	return &ApplicationNotFoundError{applicationName: name}
}

type AutoCreateSender struct {
	Sender
}

func (acs AutoCreateSender) handleMissingApplication(name string) error {
	validName := EscapeSpecialCharsForValidApplicationName(name)
	if _, err := acs.ApplicationApi.CreateApplication(CreateApplicationRequest{Name: validName}); err != nil {
		return err
	}
	return nil
}
