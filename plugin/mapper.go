package plugin

import (
	"fmt"
	"github.com/elastic/beats/v7/libbeat/common"
	"regexp"
)

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

// ConstantStringMapper ignores the common.MapStr source and returns the constantString as the mapping result
type ConstantStringMapper struct {
	constantString string
}

func (cm *ConstantStringMapper) doMap(common.MapStr) (interface{}, error) {
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
