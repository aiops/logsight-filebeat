package logsight_filebeat

import (
	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/common/transport/tlscommon"
	"github.com/elastic/beats/v7/libbeat/outputs"

	log "github.com/sirupsen/logrus"
	"net/url"
	"strings"
)

func init() {
	outputs.RegisterType("logsight", MakeLogsightAPI)
}

func MakeLogsightAPI(
	_ outputs.IndexManager,
	_ beat.Info,
	observer outputs.Observer,
	cfg *common.Config,
) (outputs.Group, error) {
	config := defaultConfig
	if err := cfg.Unpack(&config); err != nil {
		return outputs.Fail(err)
	}
	tlsConfig, err := tlscommon.LoadTLSConfig(config.TLS)
	if err != nil {
		return outputs.Fail(err)
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
		log.Info("Using proxy URL: %v", proxyURL)
	}
	params := config.Params
	if len(params) == 0 {
		params = nil
	}
	clients := make([]outputs.NetworkClient, len(hosts))
	for i, host := range hosts {
		log.Info("Making client for host: %v", host)
		hostURL, err := common.MakeURL(config.Protocol, config.Path, host, 80)
		if err != nil {
			log.Error("Invalid host param set: %v, Error: %v", host, err)
			return outputs.Fail(err)
		}
		var client outputs.NetworkClient
		client, err = NewClient(ClientSettings{
			URL:              hostURL,
			Proxy:            proxyURL,
			TLS:              tlsConfig,
			Username:         config.Username,
			Password:         config.Password,
			Parameters:       params,
			Timeout:          config.Timeout,
			CompressionLevel: config.CompressionLevel,
			Observer:         observer,
			BatchPublish:     config.BatchPublish,
			Headers:          config.Headers,
			ContentType:      config.ContentType,
			Format:           config.Format,
		})
		// client.log.Info("Final host URL: " + hostURL)

		if err != nil {
			return outputs.Fail(err)
		}
		client = outputs.WithBackoff(client, config.Backoff.Init, config.Backoff.Max)
		clients[i] = client
	}
	return outputs.SuccessNet(config.LoadBalance, config.BatchSize, config.MaxRetries, clients)
}

func parseProxyURL(raw string) (*url.URL, error) {
	if raw == "" {
		return nil, nil
	}
	parsedUrl, err := url.Parse(raw)
	if err == nil && strings.HasPrefix(parsedUrl.Scheme, "http") {
		return parsedUrl, err
	}
	// Proxy was bogus. Try prepending "http://" to it and
	// see if that parses correctly.
	return url.Parse("http://" + raw)
}
