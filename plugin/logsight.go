package plugin

import (
	"github.com/aiops/logsight-filebeat/plugin/api"
	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/common/transport/tlscommon"
	"github.com/elastic/beats/v7/libbeat/logp"
	"github.com/elastic/beats/v7/libbeat/outputs"
	"net/url"
	"os"
)

func init() {
	outputs.RegisterType("api", makeLogsight)
}

const logSelector = "api"

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

	if config.Application.Name == "" {
		if host, err := os.Hostname(); err != nil {
			config.Application.Name = host
		} else {
			config.Application.Name = api.DefaultApplicationName
		}
	}

	proxyURL, err := parseProxyURL(config.ProxyURL)
	if err != nil {
		log.Errorf("invalid url format for proxy: %v, Error: %v", proxyURL, err)
		return outputs.Fail(err)
	}

	host := config.Host
	log.Infof("Creating client for host: %v", host)
	hostURL, err := url.Parse(host)
	if err != nil {
		log.Errorf("invalid url format for host: %v, Error: %v", host, err)
		return outputs.Fail(err)
	}

	tlsConfig, err := tlscommon.LoadTLSConfig(config.TLS)
	if err != nil {
		log.Errorf("failed to load tls config %v, Error: %v", config.TLS, err)
		return outputs.Fail(err)
	}

	var client outputs.NetworkClient
	client, err = NewClient(config, hostURL, proxyURL, tlsConfig, observer, log)
	if err != nil {
		log.Errorf("failed to create client from host: %v, Error: %v", host, err)
		return outputs.Fail(err)
	}
	client = outputs.WithBackoff(client, 1, 60)
	log.Infof("created client %v", client)

	return outputs.SuccessNet(false, config.BatchSize, config.MaxRetries, []outputs.NetworkClient{client})
}

func parseProxyURL(raw string) (*url.URL, error) {
	if raw == "" {
		return nil, nil
	}
	parsedUrl, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	return parsedUrl, nil
}
