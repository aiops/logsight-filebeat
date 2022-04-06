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
	logBatchMapper *mapper.LogBatchMapper
	logSender      api.LogSender
	observer       *outputs.Observer
	logger         *logp.Logger
}

// NewClient instantiates a client.
func NewClient(config logsightConfig, hostURL *url.URL, proxyURL *url.URL, tlsConfig *tlscommon.TLSConfig, observer outputs.Observer, logger *logp.Logger) (*Client, error) {
	proxy := http.ProxyFromEnvironment
	if proxyURL != nil {
		proxy = http.ProxyURL(proxyURL)
	}
	var dialer, tlsDialer transport.Dialer

	dialer = transport.NetDialer(60 * time.Second)
	tlsDialer = transport.TLSDialer(dialer, tlsConfig, 60*time.Second)

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
		Timeout: 60 * time.Second,
	}

	baseApi := &api.BaseApi{HttpClient: httpClient, Url: hostURL}
	userApi := &api.UserApi{LoginApi: &api.LoginApi{BaseApi: baseApi}}
	user, err := userApi.Login(config.Email, config.Password)
	if err != nil {
		return nil, err
	}
	applicationApi := &api.ApplicationApi{BaseApi: baseApi, User: user}
	applicationApiCacheProxy := api.NewApplicationApiCacheProxy(applicationApi)
	logApi := &api.LogApi{BaseApi: baseApi, User: user}

	var missingAppHandler api.MissingApplicationHandler
	if config.Application.autoCreate {
		logger.Infof("Using AutoCreate application log sender.")
		missingAppHandler = api.AutoCreateMissingApplication{ApplicationApi: applicationApiCacheProxy}
	} else {
		logger.Infof("Using default log sender.")
		missingAppHandler = api.ErrorOnMissingApplication{ApplicationApi: applicationApiCacheProxy}
	}
	logSender := api.LogSender{
		LogApi:            logApi,
		MissingAppHandler: missingAppHandler,
	}

	// Create mappers
	applicationMapper, err := config.Application.toMapper()
	if err != nil {
		return nil, err
	}
	applicationNameMapper := &mapper.StringMapper{Mapper: applicationMapper}
	tagMapper, err := config.Tag.toMapper()
	if err != nil {
		return nil, err
	}
	tagStringMapper := &mapper.StringMapper{Mapper: tagMapper}
	var timestampMapper *mapper.StringMapper
	if config.Timestamp == "" {
		timestampMapper = &mapper.StringMapper{Mapper: mapper.GeneratorMapper{Generator: mapper.ISO8601TimestampGenerator{}}}
	} else {
		timestampMapper = &mapper.StringMapper{Mapper: mapper.KeyMapper{Key: config.Timestamp}}
	}
	var levelMapper *mapper.StringMapper
	if config.Level == "" {
		levelMapper = &mapper.StringMapper{Mapper: mapper.ConstantStringMapper{ConstantString: DefaultLevel}}
	} else {
		levelMapper = &mapper.StringMapper{Mapper: mapper.KeyMapper{Key: config.Level}}
	}
	messageMapper := &mapper.StringMapper{Mapper: mapper.KeyMapper{Key: config.Message}}
	metaDataMapper := &mapper.StringMapper{Mapper: mapper.ConstantStringMapper{ConstantString: ""}} // currently, not used

	logMapper := &mapper.LogMapper{
		TimestampMapper: timestampMapper,
		MessageMapper:   messageMapper,
		LevelMapper:     levelMapper,
		MetadataMapper:  metaDataMapper,
	}

	logBatchMapper := &mapper.LogBatchMapper{
		ApplicationNameMapper: applicationNameMapper,
		TagMapper:             tagStringMapper,
		LogMapper:             logMapper,
	}

	client := &Client{
		logBatchMapper: logBatchMapper,
		logSender:      logSender,
		observer:       &observer,
		logger:         logger,
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
	mappedLogBatches, err := c.eventsToMappedLogBatches(events)
	if err != nil {
		c.logger.Debugf("%v", err)
	}
	if mappedLogBatches != nil {
		resend, err := c.publish(mappedLogBatches)
		if len(resend) == 0 {
			batch.ACK()
			return nil
		} else {
			batch.RetryEvents(resend)
			return err
		}
	}
	return nil
}

func (c *Client) eventsToMappedLogBatches(events []publisher.Event) ([]*mapper.MappedLogBatch, error) {
	mappedLogBatches, failedEvents := c.logBatchMapper.ToLogBatch(events)
	if failedEvents != nil {
		if failedEvents.Len() == len(events) {
			return nil, fmt.Errorf("mapping failed for all %v logs. errors: %v",
				len(events), strings.Join(failedEvents.ErrorsAsStrings(), "\n"))
		} else {
			return mappedLogBatches, fmt.Errorf("mapping failed for %v out of %v logs.  %v",
				failedEvents.Len(), len(events), strings.Join(failedEvents.ErrorsAsStrings(), "\n"))
		}
	}
	return mappedLogBatches, nil
}

func (c Client) publish(logBatches []*mapper.MappedLogBatch) ([]publisher.Event, error) {
	var resend []publisher.Event
	var allErr error
	for _, mappedLogBatch := range logBatches {
		if err := c.logSender.Send(mappedLogBatch.LogBatch); err != nil {
			c.logger.Infof("%v", err)
			if c.isRetryError(err) {
				resend = append(resend, mappedLogBatch.Events...)
				allErr = fmt.Errorf("%w; %v", allErr, err)
			}
		}
	}
	if len(resend) != 0 {
		return resend, allErr
	} else {
		return nil, nil
	}
}

//TODO error handling
func (c Client) isRetryError(err error) bool {
	return false
}
