package plugin

import (
	"fmt"
)

type logsightConfig struct {
	Email       string      `config:"email"`
	Password    string      `config:"password"`
	Application application `config:"application"`
	Tag         tag         `config:"tag"`
	Timestamp   string      `config:"timestamp"`
	Level       string      `config:"level"`
	ProxyURL    string      `config:"proxy_url"`
}

type application struct {
	Name         string `config:"name"`
	Map          string `config:"name_map"`
	RegexMatcher string `config:"name_regex_matcher"`
	AutoCreate   bool
}

type tag struct {
	Name         string `config:"name"`
	Map          string `config:"name_map"`
	RegexMatcher string `config:"name_regex_matcher"`
}

var (
	defaultLogsightConfig = logsightConfig{
		ProxyURL: "",
		Email:    "",
		Password: "",
		Application: application{
			Name:         "",
			Map:          "",
			RegexMatcher: "",
			AutoCreate:   true,
		},
		Tag: tag{
			Name:         "",
			Map:          "",
			RegexMatcher: "",
		},
		Timestamp: "",
		Level:     "",
	}
)

func (lc *logsightConfig) Validate() error {
	if lc.Email == "" || lc.Password == "" {
		return fmt.Errorf("email and password must be defined")
	}
	return nil
}
