package logsight

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io"
)

var (
	postLogBatchConf = map[string]string{"method": "POST", "path": "/api/v1/logs"}
)

// Log data structure used in LogBatch. It must comply with the
// request body of the /api/v1/logs POST interface
type Log struct {
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
	Level     string `json:"level"`
	Metadata  string `json:"metadata"`
}

// LogBatchRequest data structure used for sending requests to logsight. It must comply with the
// request body of the /api/v1/logs POST interface
type LogBatchRequest struct {
	ApplicationId uuid.UUID `json:"applicationId"`
	Tag           string    `json:"tag"`
	Logs          []*Log    `json:"logs"`
}

// LogReceipt is returned uppon sending a LogBatchRequest to the logsight API.
type LogReceipt struct {
	ReceiptId     uuid.UUID `json:"receiptId"`
	LogsCount     int       `json:"logsCount"`
	Source        string    `json:"source"`
	ApplicationId uuid.UUID `json:"applicationId"`
}

type LogApi struct {
	BaseApi

	user User
}

func (la *LogApi) SendLogBatch(logBatchReq *LogBatchRequest) (*LogReceipt, error) {
	method := postLogBatchConf["method"]
	// Make a copy to prevent side effects
	urlLogin := la.url
	urlLogin.Path = postLogBatchConf["path"]

	req, err := la.BuildRequestWithBasicAuth(method, urlLogin.String(), nil, la.user.Email, la.user.Password)
	if err != nil {
		return nil, la.sendLogBatchError(logBatchReq, err)
	}

	resp, err := la.httpClient.Do(req)
	if err != nil {
		return nil, la.sendLogBatchError(logBatchReq, err)
	}
	defer la.closing(resp.Body)

	if err := la.CheckStatusOrErr(resp, 200); err != nil {
		return nil, la.sendLogBatchError(logBatchReq, err)
	}

	if applications, err := la.unmarshalLogReceipt(resp.Body); err != nil {
		return nil, la.sendLogBatchError(logBatchReq, err)
	} else {
		return applications, nil
	}
}

func (la *LogApi) unmarshalLogReceipt(body io.ReadCloser) (*LogReceipt, error) {
	bodyBytes, err := la.toBytes(body)
	if err != nil {
		return nil, err
	}

	var logReceipt LogReceipt
	if err := json.Unmarshal(bodyBytes, &logReceipt); err != nil {
		return nil, err
	}

	return &logReceipt, nil
}

func (la *LogApi) sendLogBatchError(logBatchReq *LogBatchRequest, err error) error {
	return fmt.Errorf("%w; log sending with request %v failed", err, logBatchReq)
}
