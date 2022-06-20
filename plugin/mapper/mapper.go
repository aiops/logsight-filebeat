package mapper

import (
	"fmt"
	"github.com/elastic/beats/v7/libbeat/beat"
	"regexp"
	"time"
)

type Mapper interface {
	DoMap(beat.Event) (interface{}, error)
}

// StringMapper is a wrapper around Mapper to ensure that the mapping result is a string
type StringMapper struct {
	Mapper Mapper
}

func (sm *StringMapper) doStringMap(event beat.Event) (string, error) {
	v, err := sm.Mapper.DoMap(event)
	if err != nil {
		return "", err
	}
	return sm.checkString(v)
}

func (sm *StringMapper) checkString(value interface{}) (string, error) {
	switch ty := value.(type) {
	case string:
		return fmt.Sprintf("%v", value), nil
	default:
		return "", fmt.Errorf("result of applying Mapper %v on string %v is not a string but %v",
			sm.Mapper, value, ty)
	}
}

// ConstantStringMapper ignores the common.MapStr source and returns the ConstantString as the mapping result
type ConstantStringMapper struct {
	ConstantString string
}

func (cm ConstantStringMapper) DoMap(beat.Event) (interface{}, error) {
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

func (fm GeneratorMapper) DoMap(beat.Event) (interface{}, error) {
	timeStr, err := fm.Generator.generate()
	if err != nil {
		return nil, err
	}
	return timeStr, nil
}

// EventTimeMapper expects a generator to generate values when DoMap is called.
type EventTimeMapper struct {
	Generator generator
}

func (etm EventTimeMapper) DoMap(event beat.Event) (interface{}, error) {
	return fmt.Sprintf(event.Timestamp.Format(time.RFC3339)), nil
}

// KeyMapper searches for the Key in a common.MapStr object and returns the values
type KeyMapper struct {
	Key string
}

func (km KeyMapper) DoMap(event beat.Event) (interface{}, error) {
	if v, err := event.GetValue(km.Key); err == nil {
		return v, nil
	} else {
		return "", fmt.Errorf("Key %v not found in logp %v", km.Key, event)
	}
}

// MultipleKeyValueMapper searches for all given Keys in a common.MapStr object and returns the values together with the
// configured key values
type MultipleKeyValueMapper struct {
	KeyValuePairs map[string]string
}

func (mkvm MultipleKeyValueMapper) DoMap(event beat.Event) (interface{}, error) {
	var values = make(map[string]interface{})
	for key, valueSource := range mkvm.KeyValuePairs {
		if v, err := event.GetValue(valueSource); err == nil {
			values[key] = v
		}
	}
	return values, nil
}

// MultipleKeyValueStringMapper is a wrapper around MultipleKeyValueMapper to ensure that the mapping results are strings
type MultipleKeyValueStringMapper struct {
	StringMapper
	Mapper MultipleKeyValueMapper
}

func (msm *MultipleKeyValueStringMapper) DoMultipleStringMap(event beat.Event) (map[string]string, error) {
	var result = make(map[string]string)
	values, err := msm.Mapper.DoMap(event)
	if err != nil {
		return nil, err
	}
	for k, v := range values.(map[string]interface{}) {
		// TODO better handling of failed mappings (somehow log errors later)
		checkedValue, err := msm.checkString(v)
		if err == nil {
			result[k] = checkedValue
		}
	}
	return result, nil
}

type KeyRegexMapper struct {
	Mapper StringMapper
	Expr   *regexp.Regexp
}

func (krm KeyRegexMapper) DoMap(event beat.Event) (interface{}, error) {
	value, err := krm.Mapper.doStringMap(event)
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
