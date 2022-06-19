package api

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io"
	"net/http"
	"regexp"
)

const levelRegex = "^INFO$|^WARNING$|^WARN$|^FINER$|^FINE$|^DEBUG$|^ERROR$|^ERR$|^EXCEPTION$|^SEVERE$"
const iso8601Regex = "^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}(\\.\\d+)?(([+-]\\d{2}:\\d{2})|Z)?$"

var (
	postLogBatchConf = map[string]string{"method": "POST", "path": "/api/v1/logs"}
)

// Log data structure used in LogBatch. It must comply with the
// request body of the /api/v1/logs POST interface
type Log struct {
	Timestamp string            `json:"timestamp" validate:"required"`
	Message   string            `json:"message" validate:"required"`
	Level     string            `json:"level" validate:"required"`
	Tags      map[string]string `json:"tags" validate:"required"`
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
		return fmt.Errorf("invalid log level. must be one of %v", levelRegex)
	}
}

func (l *Log) validateTimestamp() error {
	reg := regexp.MustCompile(iso8601Regex)
	if match := reg.MatchString(l.Timestamp); match {
		return nil
	} else {
		return fmt.Errorf("timestamp must be in ISO 8601 format (must match %v)", iso8601Regex)
	}
}

// LogReceipt is returned upon sending a LogBatchRequest to the API.
type LogReceipt struct {
	ReceiptId uuid.UUID `json:"receiptId"`
	LogsCount int       `json:"logsCount"`
	BatchId   uuid.UUID `json:"batchId"`
	Status    int       `json:"status"`
}

type LogApi struct {
	*BaseApi

	User *User
}

func (la *LogApi) SendLogs(logs []*Log) (*LogReceipt, error) {
	method := postLogBatchConf["method"]
	// Make a copy to prevent side effects
	urlLogin := la.Url
	urlLogin.Path = postLogBatchConf["path"]

	req, err := la.BuildRequestWithBasicAuth(method, urlLogin.String(), logs, la.User.Email, la.User.Password)
	if err != nil {
		return nil, la.sendLogBatchError(logs, err)
	}

	resp, err := la.HttpClient.Do(req)
	if err != nil {
		return nil, la.sendLogBatchError(logs, err)
	}
	defer la.closing(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, la.GetUnexpectedStatusError(resp, http.StatusOK)
	}
	return la.unmarshalLogReceipt(resp.Body), nil
}

func (la *LogApi) unmarshalLogReceipt(body io.ReadCloser) *LogReceipt {
	bodyBytes, err := la.toBytes(body)
	if err != nil {
		return nil
	}
	var logReceipt LogReceipt
	if err := json.Unmarshal(bodyBytes, &logReceipt); err != nil {
		return nil
	}
	return &logReceipt
}

func (la *LogApi) sendLogBatchError(logs []*Log, err error) error {
	return fmt.Errorf("%w; log sending with request %v failed", err, logs)
}
