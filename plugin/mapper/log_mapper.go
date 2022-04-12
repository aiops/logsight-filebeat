package mapper

import (
	"github.com/aiops/logsight-filebeat/plugin/api"
	"github.com/elastic/beats/v7/libbeat/beat"
	"strings"
)

// LogMapper does the mapping between filebeat's common.MapStr objects and Log objects.
type LogMapper struct {
	TimestampMapper *StringMapper
	MessageMapper   *StringMapper
	LevelMapper     *StringMapper
	MetadataMapper  *StringMapper
}

func (lm *LogMapper) ToLog(event beat.Event) (*api.Log, error) {
	timestamp, err := lm.TimestampMapper.doStringMap(event)
	if err != nil {
		return nil, err
	}
	message, err := lm.MessageMapper.doStringMap(event)
	if err != nil {
		return nil, err
	}
	level, err := lm.LevelMapper.doStringMap(event)
	if err != nil {
		return nil, err
	}
	metadata, err := lm.MetadataMapper.doStringMap(event)
	if err != nil {
		return nil, err
	}
	log := &api.Log{
		Timestamp: timestamp,
		Message:   message,
		Level:     strings.ToUpper(level),
		Metadata:  metadata,
	}
	err = log.ValidateLog()
	if err != nil {
		return nil, err
	}
	return log, nil
}
