package mapper

import (
	"fmt"
	"github.com/elastic/beats/v7/libbeat/common"
	"regexp"
	"time"
)

type Mapper interface {
	DoMap(common.MapStr) (interface{}, error)
}

// StringMapper is a wrapper around Mapper to ensure that the mapping result is a string
type StringMapper struct {
	Mapper Mapper
}

func (sm *StringMapper) doStringMap(mapSource common.MapStr) (string, error) {
	v, err := sm.Mapper.DoMap(mapSource)
	if err != nil {
		return "", err
	}
	switch ty := v.(type) {
	case string:
		return fmt.Sprintf("%v", v), nil
	default:
		return "", fmt.Errorf("result of applying Mapper %v on string %v is not a string but %v",
			sm.Mapper, v, ty)
	}
}

// ConstantStringMapper ignores the common.MapStr source and returns the ConstantString as the mapping result
type ConstantStringMapper struct {
	ConstantString string
}

func (cm ConstantStringMapper) DoMap(common.MapStr) (interface{}, error) {
	return cm.ConstantString, nil
}

type generator interface {
	generate() (string, error)
}

type ISO8601TimestampGenerator struct {
	generator
}

func (tg ISO8601TimestampGenerator) generate() (string, error) {
	timeStr := fmt.Sprintf("%v", time.Now().Format(time.RFC3339))
	return timeStr, nil
}

// GeneratorMapper expects a generator to generate values when DoMap is called.
type GeneratorMapper struct {
	Generator generator
}

func (fm GeneratorMapper) DoMap(common.MapStr) (interface{}, error) {
	timeStr, err := fm.Generator.generate()
	if err != nil {
		return nil, err
	}
	return timeStr, nil
}

// KeyMapper searches for the Key in a common.MapStr object and returns the
type KeyMapper struct {
	Key string
}

func (km KeyMapper) DoMap(mapSource common.MapStr) (interface{}, error) {
	if v, err := mapSource.GetValue(km.Key); err == nil {
		return v, nil
	} else {
		return "", fmt.Errorf("Key %v not found in logp %v", km.Key, mapSource)
	}
}

type KeyRegexMapper struct {
	Mapper StringMapper
	Expr   *regexp.Regexp
}

func (krm KeyRegexMapper) DoMap(mapSource common.MapStr) (interface{}, error) {
	value, err := krm.Mapper.doStringMap(mapSource)
	if err != nil {
		return "", err
	}
	return krm.applyRegex(value)
}

func (krm *KeyRegexMapper) applyRegex(value string) (string, error) {
	if matches := krm.Expr.FindStringSubmatch(value); matches == nil {
		return "", fmt.Errorf("no matches found in string %v with regular expression %v", value, krm.Expr)
	} else {
		// The regular expression matching should be done with submatches, i.e. the regex must contain expressions in
		// brackets which should result in a string submatch. The first of such submatches is returned.
		// It should fail if no submatches were found, even if the whole string matches.
		if len(matches) < 2 {
			return "", fmt.Errorf("no string submatches found in string %v with regular expression %v",
				value, krm.Expr)
		}
		result := matches[1]
		if result == "" {
			return "", fmt.Errorf("regular expression %v results in an Empty string for source string %v",
				value, krm.Expr)
		}
		return result, nil
	}
}
