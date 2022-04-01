package plugin

import (
	"fmt"
	"github.com/aiops/logsight-filebeat/plugin/mapper"
	"github.com/elastic/beats/v7/libbeat/common/transport/tlscommon"
)

const DefaultLevel = "INFO"

type logsightConfig struct {
	Host        string            `config:"host" validate:"required"`
	Email       string            `config:"email" validate:"required"`
	Password    string            `config:"password" validate:"required"`
	Application application       `config:"application"`
	Tag         tag               `config:"tag"`
	Message     string            `config:"message" validate:"required"`
	Timestamp   string            `config:"timestamp"`
	Level       string            `config:"level"`
	TLS         *tlscommon.Config `config:"tls"`
	ProxyURL    string            `config:"proxy_url"`
	BatchSize   int               `config:"batch_size"`
	MaxRetries  int               `config:"max_retries"`
}

type application struct {
	Name         string `config:"name"`
	Map          string `config:"name_map"`
	RegexMatcher string `config:"name_regex_matcher"`
	AutoCreate   bool
}

func (a *application) toMapper() (mapper.Mapper, error) {
	if a.Map != "" && a.RegexMatcher != "" {
		keyMapper := mapper.KeyMapper{Key: a.Map}
		return &mapper.KeyRegexMapper{Mapper: mapper.StringMapper{Mapper: keyMapper}}, nil
	} else if a.Map != "" {
		return &mapper.KeyMapper{Key: a.Map}, nil
	} else if a.Name != "" {
		return &mapper.ConstantStringMapper{ConstantString: a.Name}, nil
	} else {
		return nil, fmt.Errorf("invalid application config %v. either name or name_map must be set")
	}
}

type tag struct {
	Name         string `config:"name"`
	Map          string `config:"name_map"`
	RegexMatcher string `config:"name_regex_matcher"`
}

func (t *tag) toMapper() (mapper.Mapper, error) {
	if t.Map != "" && t.RegexMatcher != "" {
		keyMapper := mapper.KeyMapper{Key: t.Map}
		return &mapper.KeyRegexMapper{Mapper: mapper.StringMapper{Mapper: keyMapper}}, nil
	} else if t.Map != "" {
		return &mapper.KeyMapper{Key: t.Map}, nil
	} else if t.Name != "" {
		return &mapper.ConstantStringMapper{ConstantString: t.Name}, nil
	} else {
		return nil, fmt.Errorf("invalid application config %v. either name or name_map must be set")
	}
}

var (
	defaultLogsightConfig = logsightConfig{
		ProxyURL: "https://api.ai",
		Email:    "",
		Password: "",
		Application: application{
			Name:         "",
			Map:          "",
			RegexMatcher: "",
			AutoCreate:   true,
		},
		Tag: tag{
			Name:         "default",
			Map:          "",
			RegexMatcher: "",
		},
		Timestamp:  "",
		Level:      "",
		BatchSize:  50,
		MaxRetries: 20,
	}
)
