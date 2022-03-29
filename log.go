package logsight

import (
	"fmt"
	"github.com/elastic/beats/v7/libbeat/common"
	"regexp"
)

// Log data structure used in LogBatch. It must comply with the
// request body of the /api/v1/logs POST interface
type Log struct {
	timestamp string
	message   string
	level     string
	metadata  string
}

// LogBatch data structure used for sending requests to logsight. It must comply with the
// request body of the /api/v1/logs POST interface
type LogBatch struct {
	applicationName string
	tag             string
	logs            []*Log
}

type Mapper interface {
	doMap(common.MapStr) (interface{}, error)
}

// StringMapper is a wrapper around Mapper to ensure that the mapping result is a string
type StringMapper struct {
	mapper Mapper
}

func (sm *StringMapper) doStringMap(mapSource common.MapStr) (string, error) {
	v, err := sm.mapper.doMap(mapSource)
	if err != nil {
		return "", err
	}
	switch ty := v.(type) {
	case string:
		return fmt.Sprintf("%v", v), nil
	default:
		return "", fmt.Errorf("result of applying mapper %v on string %v is not a string but %v",
			sm.mapper, v, ty)
	}
}

// LogMapper does the mapping between filebeat's common.MapStr objects and Log objects.
type LogMapper struct {
	timestampMapper StringMapper
	messageMapper   StringMapper
	levelMapper     StringMapper
	metadataMapper  StringMapper
}

func (lm *LogMapper) doMap(mapSource common.MapStr) (*Log, error) {
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
		timestamp: timestamp,
		message:   message,
		level:     level,
		metadata:  metadata,
	}, nil
}

// LogBatchMapper does the mapping between filebeat's common.MapStr objects and LogBatch objects.
type LogBatchMapper struct {
	applicationNameMapper StringMapper
	tagMapper             StringMapper
	logMapper             LogMapper
}

func (lbm *LogBatchMapper) doMap(mapSources []common.MapStr) ([]*LogBatch, error) {
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
		log, err := lbm.logMapper.doMap(mapSource)
		if err != nil {
			return nil, err
		}
		key := fmt.Sprintf("%v%v", applicationName, tag)
		if val, ok := logBatchMap[key]; ok {
			val.logs = append(val.logs, log)
		} else {
			logBatchMap[key] = &LogBatch{
				applicationName: applicationName,
				tag:             tag,
				logs:            []*Log{log},
			}
		}
	}

	logBatchList := make([]*LogBatch, 0, len(logBatchMap))
	for _, logBatch := range logBatchMap {
		logBatchList = append(logBatchList, logBatch)
	}

	return logBatchList, nil
}

// ConstantStringMapper ignores the common.MapStr source and returns the constantString as the mapping result
type ConstantStringMapper struct {
	constantString string
}

func (cm *ConstantStringMapper) doMap(ignored common.MapStr) (interface{}, error) {
	return cm.constantString, nil
}

// KeyMapper searches for the key in a common.MapStr object and returns the
type KeyMapper struct {
	key string
}

func (km *KeyMapper) doMap(mapSource common.MapStr) (interface{}, error) {
	if v, err := mapSource.GetValue(km.key); err == nil {
		return v, nil
	} else {
		return "", fmt.Errorf("key %v not found in logp %v", km.key, mapSource)
	}
}

type KeyRegexMapper struct {
	mapper StringMapper
	expr   *regexp.Regexp
}

func (krm *KeyRegexMapper) doMap(mapSource common.MapStr) (interface{}, error) {
	value, err := krm.mapper.doStringMap(mapSource)
	if err != nil {
		return "", err
	}
	return krm.applyRegex(value)
}

func (krm *KeyRegexMapper) applyRegex(value string) (string, error) {
	if matches := krm.expr.FindStringSubmatch(value); matches == nil {
		return "", fmt.Errorf("no matches found in string %v with regular expression %v", value, krm.expr)
	} else {
		// The regular expression matching should be done with submatches, i.e. the regex must contain expressions in
		// brackets which should result in a string submatch. The first of such submatches is returned.
		// It should fail if no submatches were found, even if the whole string matches.
		if len(matches) < 2 {
			return "", fmt.Errorf("no string submatches found in string %v with regular expression %v",
				value, krm.expr)
		}
		result := matches[1]
		if result == "" {
			return "", fmt.Errorf("regular expression %v results in an empty string for source string %v",
				value, krm.expr)
		}
		return result, nil
	}
}
