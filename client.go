package logsight

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/common/transport"
	"github.com/elastic/beats/v7/libbeat/common/transport/tlscommon"
	"github.com/elastic/beats/v7/libbeat/logp"
	"github.com/elastic/beats/v7/libbeat/outputs"
	"github.com/elastic/beats/v7/libbeat/publisher"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// Client struct
type Client struct {
	logBatchMapper LogBatchMapper
	observer       outputs.Observer
}

type eventRaw map[string]json.RawMessage

// NewClient instantiate a client.
func NewClient(clientConfig ClientConfig, hostURL *url.URL, observer outputs.Observer, logger *logp.Logger) (*Client, error) {
	proxy := http.ProxyFromEnvironment
	if clientConfig.Proxy != "" {
		proxyURL, err := parseProxyURL(clientConfig.Proxy)
		if err != nil {
			return nil, err
		}
		if proxyURL != nil {
			logger.Infof("using proxy URL: %v", proxyURL)
		}
		proxy = http.ProxyURL(proxyURL)
	}
	var dialer, tlsDialer transport.Dialer

	dialer = transport.NetDialer(clientConfig.Timeout)
	tlsConfig, err := tlscommon.LoadTLSConfig(nil)
	if err != nil {
		return nil, err
	}
	tlsDialer = transport.TLSDialer(dialer, tlsConfig, clientConfig.Timeout)

	if st := observer; st != nil {
		dialer = transport.StatsDialer(dialer, st)
		tlsDialer = transport.StatsDialer(tlsDialer, st)
	}
	logsightAPI := &Logsight{
		baseURL:  hostURL,
		email:    clientConfig.Email,
		password: clientConfig.Password,

		request: Request{
			encoder: NewJSONEncoder(nil),
			httpClient: &http.Client{
				Transport: &http.Transport{
					Dial:    dialer.Dial,
					DialTLS: tlsDialer.Dial,
					Proxy:   proxy,
				},
				Timeout: 60 * time.Second,
			},
		},
		logger: logger,
	}
	if err := logsightAPI.Init(); err != nil {
		return nil, err
	}

	// Only accept valid regex expressions
	_, err = regexp.Compile(clientConfig.App.KeyRegexMatcher)
	if err != nil {
		return nil, err
	}

	var publisher AppPublisher
	if clientConfig.App.Key != "" {
		publisher = &MultiAppPublisher{
			SingleAppPublisher: SingleAppPublisher{
				LogsightAPI: logsightAPI,
			},
			autoCreateApp: clientConfig.App.AutoCreate,
			existentApps:  make(map[string]struct{}),
		}
	} else {
		publisher = &SingleAppPublisher{
			LogsightAPI: logsightAPI,
		}
		if clientConfig.App.AutoCreate {
			if err := logsightAPI.CreateApp(clientConfig.App.Name); err != nil {
				logger.Warnf("failed to create application ")
			}
		}
	}

	client := &Client{
		Publisher: publisher,
		logp:      logger,
		appName:   clientConfig.App.Name,
		observer:  observer,
	}

	return client, nil
}

// Connect establishes a connection to the clients sink.
func (c *Client) Connect() error {
	return nil
}

// Close closes a connection.
func (c *Client) Close() error {
	return nil
}

func (c *Client) String() string {
	return fmt.Sprintf("%v", "LogsightClient")
}

// Publish sends events to the clients sink.
func (c *Client) Publish(_ context.Context, batch publisher.Batch) error {
	st := c.observer
	events := batch.Events()
	st.NewBatch(len(events))

	if len(events) == 0 {
		batch.ACK()
		return nil
	}

	logs := make([]common.MapStr, len(events))
	for i := range events {
		logEvent := events[i].Content.Fields
		logEvent["@timestamp"] = fmt.Sprintf(time.Now().Format(time.RFC3339)) //ISO 8601
		logs[i] = logEvent
	}

	var failedCnt int
	appLogs := c.Publisher.groupByApp(logs)
	for app, logs := range appLogs {
		if err := c.Publisher.send(logs, app); err != nil {
			c.logp.Warnf("%v", err)
			failedCnt += len(logs)
			c.logp.Infof("failed to transmit %v messages to app %v", len(logs), app)
		} else {
			c.logp.Infof("successfully transmitted %v messages to app %v", len(logs), app)
		}
	}

	st.Acked(len(events) - failedCnt)
	st.Failed(failedCnt)
	batch.ACK()

	return nil
}

type LogsightPublisher interface {
	groupByApp(logs []common.MapStr) map[string][]string
	send(logs []string, app string) error
}

type SingleAppPublisher struct {
	LogsightAPI           *Logsight
	applicationNameMapper *Mapper
}

func (sap *SingleAppPublisher) groupByApp(logs []common.MapStr) map[string][]string {
	appLogs := make(map[string][]string)
	for i := range logs {
		// HIER
		appLogs["app"][i] = fmt.Sprintf("%v", logs[i])
	}

	return appLogs
}

func (sap *SingleAppPublisher) send(logs []string, app string) error {
	err := sap.LogsightAPI.SendLogs(logs, app)
	if err != nil {
		return err
	}
	return nil
}

type MultiAppPublisher struct {
	SingleAppPublisher
	autoCreateApp bool
	existentApps  map[string]struct{}
}

func (mac *MultiAppPublisher) groupByApp(logs []common.MapStr) map[string][]string {
	// the string representations of all logs are grouped by app names
	appLogs := make(map[string][]string)
	for range appLogs {
		logStr := fmt.Sprintf("%v", logs[1])
		fmt.Sprintf("%v", logStr)
		//if element, err := logs[i].GetValue(mac.key); err != nil {
		//	appLogs["defaultApp"] = append(appLogs[defaultApp], logStr)
		//} else {
		//	if app, ok := element.(string); ok {
		//		appLogs[app] = append(appLogs[app], logStr)
		//	} else {
		//		appLogs["defaultApp"] = append(appLogs[defaultApp], logStr)
		//	}
		//}
	}

	return appLogs
}

func (mac *MultiAppPublisher) send(logs []string, app string) error {
	app = prepareAppName(strings.ToLower(app))
	if _, ok := mac.existentApps[app]; !ok {
		if err := mac.LogsightAPI.CreateApp(app); err != nil {
			return fmt.Errorf("failed to send logs to app %v: %v", app, err)
		} else {
			mac.existentApps[app] = struct{}{}
		}
	}
	err := mac.LogsightAPI.SendLogs(logs, app)
	if err != nil {
		return err
	}
	return nil
}

func prepareAppName(appName string) string {
	// Make a Regex to say we only want letters and numbers
	reg, _ := regexp.Compile("[^a-z0-9_]+")
	return reg.ReplaceAllString(appName, "")

}
