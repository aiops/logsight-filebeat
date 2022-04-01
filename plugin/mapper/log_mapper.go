package mapper

import (
	"github.com/aiops/logsight-filebeat/plugin/api"
	"github.com/elastic/beats/v7/libbeat/common"
	"strings"
)

// LogMapper does the mapping between filebeat's common.MapStr objects and Log objects.
type LogMapper struct {
	TimestampMapper *StringMapper
	MessageMapper   *StringMapper
	LevelMapper     *StringMapper
	MetadataMapper  *StringMapper
}

func (lm *LogMapper) ToLog(mapSource common.MapStr) (*api.Log, error) {
	timestamp, err := lm.TimestampMapper.doStringMap(mapSource)
	if err != nil {
		return nil, err
	}
	message, err := lm.MessageMapper.doStringMap(mapSource)
	if err != nil {
		return nil, err
	}
	level, err := lm.LevelMapper.doStringMap(mapSource)
	if err != nil {
		return nil, err
	}
	metadata, err := lm.MetadataMapper.doStringMap(mapSource)
	if err != nil {
		return nil, err
	}
	return &api.Log{
		Timestamp: timestamp,
		Message:   message,
		Level:     strings.ToUpper(level),
		Metadata:  metadata,
	}, nil
}
