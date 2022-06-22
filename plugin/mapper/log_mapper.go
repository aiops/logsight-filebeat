package mapper

import (
	"github.com/aiops/logsight-filebeat/plugin/api"
	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/publisher"
	"strings"
)

type FailedMapping struct {
	Event *publisher.Event
	Err   *error
}

func (fl *FailedMapping) Append(event *publisher.Event, err *error) {
	fl.Event = event
	fl.Err = err
}

// LogMapper does the mapping between filebeat's common.MapStr objects and Log objects.
type LogMapper struct {
	TimestampMapper *StringMapper
	MessageMapper   *StringMapper
	LevelMapper     *StringMapper
	TagsMapper      *MultipleKeyValueStringMapper
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
	tags, err := lm.TagsMapper.DoMultipleStringMap(event)
	if err != nil {
		return nil, err
	}
	log := &api.Log{
		Timestamp: timestamp,
		Message:   message,
		Level:     strings.ToUpper(level),
		Tags:      tags,
	}
	err = log.ValidateLog()
	if err != nil {
		return nil, err
	}
	return log, nil
}

func (lm *LogMapper) ToLogs(events []publisher.Event) ([]*api.Log, []*FailedMapping) {
	var logs []*api.Log
	var failedMappings []*FailedMapping

	for _, event := range events {
		log, err := lm.ToLog(event.Content)
		if err != nil {
			failedMappings = append(failedMappings, &FailedMapping{
				Event: &event,
				Err:   &err,
			})
		} else {
			logs = append(logs, log)
		}
	}

	if len(failedMappings) > 0 {
		return logs, failedMappings
	} else {
		return logs, nil
	}
}
