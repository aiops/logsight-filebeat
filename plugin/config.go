package plugin

import (
	"encoding/json"
	"fmt"
	"github.com/aiops/logsight-filebeat/plugin/mapper"
	"github.com/elastic/beats/v7/libbeat/common/transport/tlscommon"
	"regexp"
	"time"
)

const DefaultLevel = "INFO"

type logsightConfig struct {
	Url          string            `config:"url" validate:"required"`
	Email        string            `config:"email" validate:"required"`
	Password     string            `config:"password" validate:"required"`
	Application  applicationConf   `config:"application"`
	Tag          tagConf           `config:"tag"`
	MessageKey   string            `config:"message_key"`
	TimestampKey string            `config:"timestamp_key"`
	LevelKey     string            `config:"level_key"`
	TLS          *tlscommon.Config `config:"tls"`
	ProxyURL     string            `config:"proxy_url"`
	BatchSize    int               `config:"batch_size"`
	MaxRetries   int               `config:"max_retries"`
	Timeout      time.Duration     `config:"timeout"`
}

func (lc *logsightConfig) String() string {
	strResult, _ := json.Marshal(lc)
	return string(strResult)
}

type applicationConf struct {
	Name         string `config:"name"`
	Map          string `config:"name_key"`
	RegexMatcher string `config:"name_regex_matcher"`
	AutoCreate   bool   `config:"auto_create"`
}

func (ac *applicationConf) toMapper() (mapper.Mapper, error) {
	mc := mapperConf{
		Name:         ac.Name,
		Key:          ac.Map,
		RegexMatcher: ac.RegexMatcher,
	}
	return mc.toMapper()
}

type tagConf struct {
	Name         string `config:"name"`
	Key          string `config:"name_key"`
	RegexMatcher string `config:"name_regex_matcher"`
}

func (tc *tagConf) toMapper() (mapper.Mapper, error) {
	mc := mapperConf{
		Name:         tc.Name,
		Key:          tc.Key,
		RegexMatcher: tc.RegexMatcher,
	}
	return mc.toMapper()
}

type mapperConf struct {
	Name         string
	Key          string
	RegexMatcher string
}

func (mc *mapperConf) toMapper() (mapper.Mapper, error) {
	if mc.Key != "" && mc.RegexMatcher != "" {
		expr, err := regexp.Compile(mc.RegexMatcher)
		if err != nil {
			return nil, fmt.Errorf("%w; invalid regex expression %v", err, mc.RegexMatcher)
		}
		keyMapper := mapper.KeyMapper{Key: mc.Key}
		return &mapper.KeyRegexMapper{Mapper: mapper.StringMapper{Mapper: keyMapper}, Expr: expr}, nil
	} else if mc.Key != "" {
		return &mapper.KeyMapper{Key: mc.Key}, nil
	} else if mc.Name != "" {
		return &mapper.ConstantStringMapper{ConstantString: mc.Name}, nil
	} else {
		return nil, fmt.Errorf("invalid application config %v. either name or name_map must be set", mc)
	}
}

var (
	defaultLogsightConfig = logsightConfig{
		Url:      "",
		Email:    "",
		Password: "",
		Application: applicationConf{
			Name:         "",
			Map:          "",
			RegexMatcher: "",
			AutoCreate:   true,
		},
		Tag: tagConf{
			Name:         "default",
			Key:          "",
			RegexMatcher: "",
		},
		MessageKey:   "message",
		TimestampKey: "",
		LevelKey:     "",
		BatchSize:    100,
		MaxRetries:   20,
		Timeout:      120,
	}
)
