package plugin

import (
	"fmt"
	"github.com/elastic/beats/v7/libbeat/common"
)

// LogMapper does the mapping between filebeat's common.MapStr objects and Log objects.
type LogMapper struct {
	timestampMapper StringMapper
	messageMapper   StringMapper
	levelMapper     StringMapper
	metadataMapper  StringMapper
}

func (lm *LogMapper) toLog(mapSource common.MapStr) (*Log, error) {
	timestamp, err := lm.timestampMapper.doStringMap(mapSource)
	if err != nil {
		return nil, err
	}
	message, err := lm.messageMapper.doStringMap(mapSource)
	if err != nil {
		return nil, err
	}
	level, err := lm.levelMapper.doStringMap(mapSource)
	if err != nil {
		return nil, err
	}
	metadata, err := lm.metadataMapper.doStringMap(mapSource)
	if err != nil {
		return nil, err
	}
	return &Log{
		Timestamp: timestamp,
		Message:   message,
		Level:     level,
		Metadata:  metadata,
	}, nil
}

// LogBatchMapper does the mapping between filebeat's common.MapStr objects and LogBatch objects.
type LogBatchMapper struct {
	applicationNameMapper StringMapper
	tagMapper             StringMapper
	logMapper             LogMapper
}

func (lbm *LogBatchMapper) toLogBatch(mapSources []common.MapStr) ([]*LogBatch, error) {
	logBatchMap := make(map[string]*LogBatch)
	for _, mapSource := range mapSources {
		applicationName, err := lbm.applicationNameMapper.doStringMap(mapSource)
		if err != nil {
			return nil, err
		}
		tag, err := lbm.tagMapper.doStringMap(mapSource)
		if err != nil {
			return nil, err
		}
		log, err := lbm.logMapper.toLog(mapSource)
		if err != nil {
			return nil, err
		}
		key := fmt.Sprintf("%v%v", applicationName, tag)
		if val, ok := logBatchMap[key]; ok {
			val.Logs = append(val.Logs, log)
		} else {
			logBatchMap[key] = &LogBatch{
				ApplicationName: applicationName,
				Tag:             tag,
				Logs:            []*Log{log},
			}
		}
	}

	logBatchList := make([]*LogBatch, 0, len(logBatchMap))
	for _, logBatch := range logBatchMap {
		logBatchList = append(logBatchList, logBatch)
	}

	return logBatchList, nil
}
