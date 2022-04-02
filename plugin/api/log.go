package api

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io"
	"regexp"
)

const levelRegex = "^INFO$|^WARNING$|^WARN$|^FINER$|^FINE$|^DEBUG$|^ERROR$|^ERR$|^EXCEPTION$|^SEVERE$"

var (
	InvalidLevelError = fmt.Errorf("level must match one of %v", levelRegex)

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

func (l *Log) ValidateLog() error {
	if err := l.validateLevel(); err != nil {
		return err
	}
	if err := l.validateTimestamp(); err != nil {
		return err
	}
	return nil
}

func (l *Log) validateLevel() error {
	reg := regexp.MustCompile(levelRegex)
	if match := reg.MatchString(l.Level); match {
		return nil
	} else {
		return InvalidLevelError
	}
}

func (l *Log) validateTimestamp() error {
	iso8601Regex := "^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}(\\.\\d+)?(([+-]\\d{2}:\\d{2})|Z)?$"
	reg := regexp.MustCompile(iso8601Regex)
	if match := reg.MatchString(l.Timestamp); match {
		return nil
	} else {
		return fmt.Errorf("timestamp must be in ISO 8601 format (match %v)", iso8601Regex)
	}
}

// LogBatchRequest data structure used for sending requests to api. It must comply with the
// request body of the /api/v1/logs POST interface
type LogBatchRequest struct {
	ApplicationId uuid.UUID `json:"applicationId"`
	Tag           string    `json:"tag"`
	Logs          []*Log    `json:"logs"`
}

// LogReceipt is returned uppon sending a LogBatchRequest to the api API.
type LogReceipt struct {
	ReceiptId     uuid.UUID `json:"receiptId"`
	LogsCount     int       `json:"logsCount"`
	Source        string    `json:"source"`
	ApplicationId uuid.UUID `json:"applicationId"`
}

type LogApi struct {
	*BaseApi

	User *User
}

func (la *LogApi) SendLogBatch(logBatchReq *LogBatchRequest) (*LogReceipt, error) {
	method := postLogBatchConf["method"]
	// Make a copy to prevent side effects
	urlLogin := la.Url
	urlLogin.Path = postLogBatchConf["path"]

	req, err := la.BuildRequestWithBasicAuth(method, urlLogin.String(), nil, la.User.Email, la.User.Password)
	if err != nil {
		return nil, la.sendLogBatchError(logBatchReq, err)
	}

	resp, err := la.HttpClient.Do(req)
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
