package mapper

import (
	"fmt"
	"github.com/aiops/logsight-filebeat/plugin/api"
	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/publisher"
)

type MappedLogBatch struct {
	LogBatch *api.LogBatch
	Events   []publisher.Event
}

type FailedEvents struct {
	Events []*publisher.Event
	Errs   []error
}

func (fl *FailedEvents) Append(event *publisher.Event, err error) {
	fl.Events = append(fl.Events, event)
	fl.Errs = append(fl.Errs, err)
}

func (fl *FailedEvents) Empty() bool {
	return len(fl.Events) == 0
}

func (fl *FailedEvents) Len() int {
	return len(fl.Events)
}

func (fl *FailedEvents) ErrorsAsStrings() []string {
	errStrings := make([]string, len(fl.Errs))
	for i, err := range fl.Errs {
		errStrings[i] = err.Error()
	}
	return errStrings
}

// LogBatchMapper does the mapping between filebeat's common.MapStr objects and LogBatch objects.
type LogBatchMapper struct {
	ApplicationNameMapper *StringMapper
	TagMapper             *StringMapper
	LogMapper             *LogMapper
}

func (lbm *LogBatchMapper) ToLogBatch(events []publisher.Event) ([]*MappedLogBatch, *FailedEvents) {
	mapSources := make([]common.MapStr, len(events))
	for i := range events {
		mapSources[i] = events[i].Content.Fields
	}

	mappedLogBatchMap := make(map[string]*MappedLogBatch)
	failedEvents := &FailedEvents{Events: []*publisher.Event{}, Errs: []error{}}
	for _, event := range events {
		eventObj := event.Content
		applicationName, err := lbm.ApplicationNameMapper.doStringMap(eventObj)
		if err != nil {
			failedEvents.Append(&event, err)
			continue
		}
		tag, err := lbm.TagMapper.doStringMap(eventObj)
		if err != nil {
			failedEvents.Append(&event, err)
			continue
		}
		log, err := lbm.LogMapper.ToLog(event.Content)
		if err != nil {
			failedEvents.Append(&event, err)
			continue
		}
		key := fmt.Sprintf("%v%v", applicationName, tag)
		if val, ok := mappedLogBatchMap[key]; ok {
			val.LogBatch.Logs = append(val.LogBatch.Logs, log)
			val.Events = append(val.Events, event)
		} else {
			mappedLogBatchMap[key] = &MappedLogBatch{
				LogBatch: &api.LogBatch{
					ApplicationName: applicationName,
					Tag:             tag,
					Logs:            []*api.Log{log},
				},
				Events: []publisher.Event{event},
			}
		}
	}

	result := make([]*MappedLogBatch, len(mappedLogBatchMap))
	i := 0
	for _, logBatch := range mappedLogBatchMap {
		result[i] = logBatch
		i++
	}

	if failedEvents.Empty() {
		return result, nil
	} else {
		return result, failedEvents
	}
}
