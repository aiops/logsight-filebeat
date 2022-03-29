package logsight

import (
	"fmt"
	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/logp"
	"github.com/elastic/beats/v7/libbeat/outputs"
	"net/url"
	"os"
	"strings"
	"time"
)

type ClientConfig struct {
	Proxy        string        `config:"proxy"`
	TLS          string        `config:"tls"`
	Email        string        `config:"email"`
	Password     string        `config:"password"`
	App          application   `config:"application"`
	BatchPublish bool          `config:"batch_publish"`
	BatchSize    int           `config:"batch_size"`
	MaxRetries   int           `config:"max_retries"`
	Timeout      time.Duration `config:"timeout"`
	Backoff      backoff       `config:"backoff"`
}

type backoff struct {
	Init time.Duration
	Max  time.Duration
}

type application struct {
	Name            string
	Key             string
	KeyRegexMatcher string
	AutoCreate      bool
}

var (
	defaultConfig = ClientConfig{
		Proxy:    "",
		Email:    "",
		Password: "",
		App: application{
			Name:            "",
			Key:             "",
			KeyRegexMatcher: "",
			AutoCreate:      true,
		},
		BatchPublish: false,
		BatchSize:    2,
		Timeout:      90 * time.Second,
		MaxRetries:   3,
		Backoff: backoff{
			Init: 1 * time.Second,
			Max:  60 * time.Second,
		},
	}
)

func init() {
	outputs.RegisterType("logsight", MakeLogsightAPI)
}

const logSelector = "logsight"

func MakeLogsightAPI(
	_ outputs.IndexManager,
	_ beat.Info,
	observer outputs.Observer,
	cfg *common.Config,
) (outputs.Group, error) {
	log := logp.NewLogger(logSelector)

	config := defaultConfig
	if err := cfg.Unpack(&config); err != nil {
		return outputs.Fail(err)
	}

	if config.Email == "" {
		return outputs.Fail(fmt.Errorf("email parameter needs to be set for logsight output"))
	}
	if config.Password == "" {
		return outputs.Fail(fmt.Errorf("password parameter needs to be set for logsight output"))
	}

	if config.App.Name == "" {
		if host, err := os.Hostname(); err != nil {
			config.App.Name = host
		} else {
			config.App.Name = "filebeat_source"
		}
	}

	hosts, err := outputs.ReadHostList(cfg)
	if err != nil {
		return outputs.Fail(err)
	}
	proxyURL, err := parseProxyURL("")
	if err != nil {
		return outputs.Fail(err)
	}
	if proxyURL != nil {
		log.Infof("Using proxy URL: %v", proxyURL)
	}
	clients := make([]outputs.NetworkClient, len(hosts))
	for i, host := range hosts {
		log.Infof("Creating client for host: %v", host)
		hostURL, err := url.Parse(host)
		if err != nil {
			log.Errorf("invalid url format: %v, Error: %v", host, err)
			return outputs.Fail(err)
		}
		var client outputs.NetworkClient
		client, err = NewClient(config, hostURL, observer, log)
		if err != nil {
			return outputs.Fail(err)
		}
		log.Infof("created client %v", client)

		client = outputs.WithBackoff(client, config.Backoff.Init, config.Backoff.Max)
		clients[i] = client
	}
	log.Infof("created %v clients", len(clients))
	return outputs.SuccessNet(false, config.BatchSize, config.MaxRetries, clients)
}

func parseProxyURL(raw string) (*url.URL, error) {
	if raw == "" {
		return nil, nil
	}
	parsedUrl, err := url.Parse(raw)
	if err == nil && strings.HasPrefix(parsedUrl.Scheme, "httpClient") {
		return parsedUrl, err
	}
	// Proxy was bogus. Try prepending "httpClient://" to it and
	// see if that parses correctly.
	return url.Parse("httpClient://" + raw)
}
