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
	MessageKey   string            `config:"message_key"`
	TimestampKey string            `config:"timestamp_key"`
	LevelKey     string            `config:"level_key"`
	TagsMapping  map[string]string `config:"tags_mapping"`
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
		Url:          "",
		Email:        "",
		Password:     "",
		MessageKey:   "message",
		TimestampKey: "",
		LevelKey:     "",
		TagsMapping:  map[string]string{"service": "host"},
		BatchSize:    100,
		MaxRetries:   20,
		Timeout:      120,
	}
)
