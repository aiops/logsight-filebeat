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
	"time"
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
	logger := logp.NewLogger(logSelector)

	config := defaultLogsightConfig
	if err := cfg.Unpack(&config); err != nil {
		return outputs.Fail(err)
	}
	logger.Debugf("unpacked logsight config: %v", config.String())

	if config.Application.Name == "" {
		if host, err := os.Hostname(); err != nil {
			config.Application.Name = host
		} else {
			config.Application.Name = api.DefaultApplicationName
		}
	}

	proxyURL, err := parseProxyURL(config.ProxyURL)
	if err != nil {
		logger.Errorf("invalid url format for proxy: %v, Error: %v", proxyURL, err)
		return outputs.Fail(err)
	}

	host := config.Url
	logger.Infof("Creating client for host: %v", host)
	hostURL, err := url.Parse(host)
	if err != nil {
		logger.Errorf("invalid url format for host: %v, Error: %v", host, err)
		return outputs.Fail(err)
	}

	tlsConfig, err := tlscommon.LoadTLSConfig(config.TLS)
	if err != nil {
		logger.Errorf("failed to load tls config %v, Error: %v", config.TLS, err)
		return outputs.Fail(err)
	}
	logger.Debugf("TLS config: %v", tlsConfig)

	var client outputs.NetworkClient
	client, err = NewClient(config, hostURL, proxyURL, tlsConfig, observer, logger)
	if err != nil {
		logger.Errorf("failed to create client from host: %v, Error: %v", host, err)
		return outputs.Fail(err)
	}
	client = outputs.WithBackoff(client, 10*time.Second, 60*time.Minute)
	logger.Infof("created client %v", client)

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
