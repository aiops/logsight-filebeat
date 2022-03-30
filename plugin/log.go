package plugin

import "github.com/google/uuid"

// Log data structure used in LogBatch. It must comply with the
// request body of the /api/v1/logs POST interface
type Log struct {
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
	Level     string `json:"level"`
	Metadata  string `json:"metadata"`
}

// LogBatch data structure
type LogBatch struct {
	ApplicationName string
	Tag             string
	Logs            []*Log
}

// LogBatchRequest data structure used for sending requests to logsight. It must comply with the
// request body of the /api/v1/logs POST interface
type LogBatchRequest struct {
	ApplicationId uuid.UUID `json:"applicationId"`
	Tag           string    `json:"tag"`
	Logs          []*Log    `json:"logs"`
}

type LogApi struct {
}
