package plugin

import (
	"fmt"
	"github.com/aiops/logsight-filebeat/plugin/logsight"
	"github.com/elastic/beats/v7/libbeat/common"
)

// LogMapper does the mapping between filebeat's common.MapStr objects and Log objects.
type LogMapper struct {
	timestampMapper StringMapper
	messageMapper   StringMapper
	levelMapper     StringMapper
	metadataMapper  StringMapper
}

func (lm *LogMapper) toLog(mapSource common.MapStr) (*logsight.Log, error) {
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
	return &logsight.Log{
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

func (lbm *LogBatchMapper) toLogBatch(mapSources []common.MapStr) ([]*logsight.LogBatch, error) {
	logBatchMap := make(map[string]*logsight.LogBatch)
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
			logBatchMap[key] = &logsight.LogBatch{
				ApplicationName: applicationName,
				Tag:             tag,
				Logs:            []*logsight.Log{log},
			}
		}
	}

	logBatchList := make([]*logsight.LogBatch, 0, len(logBatchMap))
	for _, logBatch := range logBatchMap {
		logBatchList = append(logBatchList, logBatch)
	}

	return logBatchList, nil
}
