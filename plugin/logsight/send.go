package logsight

import (
	"errors"
	"fmt"
	"github.com/aiops/logsight-filebeat/plugin"
	"github.com/elastic/beats/v7/libbeat/publisher"
	"github.com/google/uuid"
)

// LogBatch data structure
type LogBatch struct {
	ApplicationName string
	Tag             string
	Logs            []*Log
	pubEventRefs    []*publisher.Event // Refs to libbeat publisher.Event to track failed sending for later retry or drop
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

type Sender interface {
	send(logBatches []*LogBatch) error
}

type AppSender struct {
	Sender

	logBatchMapper *plugin.LogBatchMapper
	logApi         *LogApi
	applicationApi ApplicationApiInterface
}

func (as *AppSender) Send(logBatches []*LogBatch) []*LogBatch {
	var failedLogBatches []*LogBatch
	for _, logBatch := range logBatches {
		err := as.sendLogBatch(logBatch)

		//TODO: Connection fail, sending failed (only for certain failures maybe)
		var applicationNotFoundError *ApplicationNotFoundError
		if !errors.As(err, &applicationNotFoundError) {
			failedLogBatches = append(failedLogBatches, logBatch)
		}
	}
	return logBatches
}

func (as *AppSender) sendLogBatch(logBatch *LogBatch) error {
	application, err := as.applicationApi.GetApplicationByName(logBatch.ApplicationName)
	if err != nil {
		return err
	}
	if application == nil {
		if err := as.handleMissingApplication(logBatch.ApplicationName); err != nil {
			return err
		}
	}

	logBatchReq := logBatch.ToLogBatchRequest(application.Id)
	if _, err := as.logApi.SendLogBatch(logBatchReq); err != nil {
		return err
	}

	return nil
}

func (as *AppSender) handleMissingApplication(name string) error {
	return &ApplicationNotFoundError{applicationName: name}
}

type AutoCreateAppSender struct {
	AppSender
}

func (acs *AutoCreateAppSender) handleMissingApplication(name string) error {
	if _, err := acs.applicationApi.CreateApplication(name); err != nil {
		return err
	}
	return nil
}


