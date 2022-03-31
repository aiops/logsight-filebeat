package plugin

import (
	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/logp"
	"github.com/elastic/beats/v7/libbeat/outputs"
	"net/url"
	"os"
	"strings"
)

func init() {
	outputs.RegisterType("logsight", makeLogsight)
}

const logSelector = "logsight"

func makeLogsight(
	im outputs.IndexManager,
	beat beat.Info,
	observer outputs.Observer,
	cfg *common.Config,
) (outputs.Group, error) {
	log := logp.NewLogger(logSelector)

	config := defaultLogsightConfig
	if err := cfg.Unpack(&config); err != nil {
		return outputs.Fail(err)
	}
	if err := config.Validate(); err != nil {
		return outputs.Fail(err)
	}

	if config.Application.Name == "" {
		if host, err := os.Hostname(); err != nil {
			config.Application.Name = host
		} else {
			config.Application.Name = "filebeat_source"
		}
	}

	hosts, err := outputs.ReadHostList(cfg)
	if err != nil {
		return outputs.Fail(err)
	}
	proxyURL, err := parseProxyURL(config.ProxyURL)
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
