package plugin

import (
	"context"
	"fmt"
	"github.com/aiops/logsight-filebeat/plugin/api"
	"github.com/aiops/logsight-filebeat/plugin/mapper"
	"github.com/elastic/beats/v7/libbeat/common/transport"
	"github.com/elastic/beats/v7/libbeat/common/transport/tlscommon"
	"github.com/elastic/beats/v7/libbeat/logp"
	"github.com/elastic/beats/v7/libbeat/outputs"
	"github.com/elastic/beats/v7/libbeat/publisher"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client struct
type Client struct {
	logMapper *mapper.LogMapper
	logSender api.LogSender
	observer  *outputs.Observer
	logger    *logp.Logger
}

// NewClient instantiates a client.
func NewClient(config logsightConfig, hostURL *url.URL, proxyURL *url.URL, tlsConfig *tlscommon.TLSConfig, observer outputs.Observer, logger *logp.Logger) (*Client, error) {
	proxy := http.ProxyFromEnvironment
	if proxyURL != nil {
		proxy = http.ProxyURL(proxyURL)
	}
	var dialer, tlsDialer transport.Dialer

	dialer = transport.NetDialer(config.Timeout * time.Second)
	tlsDialer = transport.TLSDialer(dialer, tlsConfig, config.Timeout*time.Second)

	if st := observer; st != nil {
		dialer = transport.StatsDialer(dialer, st)
		tlsDialer = transport.StatsDialer(tlsDialer, st)
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			Dial:    dialer.Dial,
			DialTLS: tlsDialer.Dial,
			Proxy:   proxy,
		},
		Timeout: config.Timeout * time.Second,
	}

	baseApi := &api.BaseApi{HttpClient: httpClient, Url: hostURL}
	userApi := &api.UserApi{LoginApi: &api.LoginApi{BaseApi: baseApi}}
	user, err := userApi.Login(config.Email, config.Password)
	if err != nil {
		return nil, err
	}
	logApi := &api.LogApi{BaseApi: baseApi, User: user}
	logSender := api.LogSender{
		LogApi: logApi,
	}

	// Create mappers
	var timestampMapper *mapper.StringMapper
	if config.TimestampKey == "" {
		timestampMapper = &mapper.StringMapper{Mapper: mapper.EventTimeMapper{}}
	} else {
		timestampMapper = &mapper.StringMapper{Mapper: mapper.KeyMapper{Key: config.TimestampKey}}
	}
	var levelMapper *mapper.StringMapper
	if config.LevelKey == "" {
		levelMapper = &mapper.StringMapper{Mapper: mapper.ConstantStringMapper{ConstantString: DefaultLevel}}
	} else {
		levelMapper = &mapper.StringMapper{Mapper: mapper.KeyMapper{Key: config.LevelKey}}
	}
	messageMapper := &mapper.StringMapper{Mapper: mapper.KeyMapper{Key: config.MessageKey}}
	tagsMapper := &mapper.MultipleKeyValueStringMapper{
		Mapper: mapper.MultipleKeyValueMapper{KeyValuePairs: config.TagsMapping},
	}

	logMapper := &mapper.LogMapper{
		TimestampMapper: timestampMapper,
		MessageMapper:   messageMapper,
		LevelMapper:     levelMapper,
		TagsMapper:      tagsMapper,
	}

	client := &Client{
		logMapper: logMapper,
		logSender: logSender,
		observer:  &observer,
		logger:    logger,
	}

	return client, nil
}

func (c *Client) Connect() error {
	return nil
}

func (c *Client) Close() error {
	c.logSender.Close()
	return nil
}

func (c *Client) String() string {
	return fmt.Sprintf("%v", "logsight client")
}

// Publish sends events to the clients sink.
func (c *Client) Publish(_ context.Context, batch publisher.Batch) error {
	events := batch.Events()
	mappedLogs, err := c.eventsToMappedLogs(events)
	if err != nil {
		c.logger.Debugf("%v", err)
	}
	if mappedLogs != nil {
		err := c.publish(mappedLogs)
		if err == nil {
			batch.ACK()
			return nil
		} else {
			batch.RetryEvents(events)
			return err
		}
	}
	return nil
}

func (c *Client) eventsToMappedLogs(events []publisher.Event) ([]*api.Log, error) {
	mappedLogs, failedEvents := c.logMapper.ToLogs(events)
	if failedEvents != nil {
		if len(failedEvents) == len(events) {
			return nil, fmt.Errorf("mapping failed for all %v logs. errors: %v",
				len(events), strings.Join(c.ErrorsAsStrings(failedEvents), "\n"))
		} else {
			return mappedLogs, fmt.Errorf("mapping failed for %v out of %v logs.  %v",
				len(failedEvents), len(events), strings.Join(c.ErrorsAsStrings(failedEvents), "\n"))
		}
	}
	return mappedLogs, nil
}

func (c *Client) ErrorsAsStrings(failedMappings []*mapper.FailedMapping) []string {
	errStrings := make([]string, len(failedMappings))
	for i, fm := range failedMappings {
		errStrings[i] = fmt.Sprintf("%v", fm.Err)
	}
	return errStrings
}

func (c Client) publish(logs []*api.Log) error {
	return c.logSender.Send(logs)
}

//TODO error handling
func (c Client) isRetryError(err error) bool {
	return true
}
