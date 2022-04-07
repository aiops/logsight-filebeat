package plugin

import (
	"fmt"
	"github.com/aiops/logsight-filebeat/plugin/mapper"
	"github.com/elastic/beats/v7/libbeat/common/transport/tlscommon"
	"regexp"
)

const DefaultLevel = "INFO"

type logsightConfig struct {
	Url         string            `config:"url" validate:"required"`
	Email       string            `config:"email" validate:"required"`
	Password    string            `config:"password" validate:"required"`
	Application applicationConf   `config:"application"`
	Tag         tagConf           `config:"tag"`
	Message     string            `config:"message"`
	Timestamp   string            `config:"timestamp"`
	Level       string            `config:"level"`
	TLS         *tlscommon.Config `config:"tls"`
	ProxyURL    string            `config:"proxy_url"`
	BatchSize   int               `config:"batch_size"`
	MaxRetries  int               `config:"max_retries"`
}

type applicationConf struct {
	Name         string `config:"name"`
	Map          string `config:"name_map"`
	RegexMatcher string `config:"name_regex_matcher"`
	AutoCreate   bool   `config:"auto_create"`
}

func (ac *applicationConf) toMapper() (mapper.Mapper, error) {
	mc := mapperConf{
		Name:         ac.Name,
		Map:          ac.Map,
		RegexMatcher: ac.RegexMatcher,
	}
	return mc.toMapper()
}

type tagConf struct {
	Name         string `config:"name"`
	Map          string `config:"name_map"`
	RegexMatcher string `config:"name_regex_matcher"`
}

func (tc *tagConf) toMapper() (mapper.Mapper, error) {
	mc := mapperConf{
		Name:         tc.Name,
		Map:          tc.Map,
		RegexMatcher: tc.RegexMatcher,
	}
	return mc.toMapper()
}

type mapperConf struct {
	Name         string
	Map          string
	RegexMatcher string
}

func (mc *mapperConf) toMapper() (mapper.Mapper, error) {
	fmt.Printf("%v, %v, %v", mc.Name, mc.Map, mc.RegexMatcher)
	if mc.Map != "" && mc.RegexMatcher != "" {
		expr, err := regexp.Compile(mc.RegexMatcher)
		if err != nil {
			return nil, fmt.Errorf("%w; invalid regex expression %v", err, mc.RegexMatcher)
		}
		keyMapper := mapper.KeyMapper{Key: mc.Map}
		return &mapper.KeyRegexMapper{Mapper: mapper.StringMapper{Mapper: keyMapper}, Expr: expr}, nil
	} else if mc.Map != "" {
		return &mapper.KeyMapper{Key: mc.Map}, nil
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
			Map:          "",
			RegexMatcher: "",
		},
		Message:    "message",
		Timestamp:  "",
		Level:      "",
		BatchSize:  50,
		MaxRetries: 20,
	}
)
