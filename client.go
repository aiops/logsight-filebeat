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
	Publisher AppPublisher
	appName   string
	log       *logp.Logger
	observer  outputs.Observer
}

type eventRaw map[string]json.RawMessage

// NewClient instantiate a client.
func NewClient(cc ClientConfig, hostURL *url.URL, observer outputs.Observer, log *logp.Logger) (*Client, error) {
	proxy := http.ProxyFromEnvironment
	if cc.Proxy != "" {
		proxyURL, err := parseProxyURL(cc.Proxy)
		if err != nil {
			return nil, err
		}
		if proxyURL != nil {
			log.Infof("using proxy URL: %v", proxyURL)
		}
		proxy = http.ProxyURL(proxyURL)
	}
	// logger.Info("HTTP URL: %s", s.URL)
	var dialer, tlsDialer transport.Dialer

	dialer = transport.NetDialer(cc.Timeout)
	tlsConfig, err := tlscommon.LoadTLSConfig(nil)
	if err != nil {
		return nil, err
	}
	tlsDialer = transport.TLSDialer(dialer, tlsConfig, cc.Timeout)

	if st := observer; st != nil {
		dialer = transport.StatsDialer(dialer, st)
		tlsDialer = transport.StatsDialer(tlsDialer, st)
	}
	logsightAPI := Logsight{
		baseURL:  hostURL,
		email:    cc.Email,
		password: cc.Password,

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
			log: log,
		},
		log: log,
	}
	if err := logsightAPI.Init(); err != nil {
		return nil, err
	}

	var publisher AppPublisher
	if cc.App.Key != "" {
		publisher = &MultiAppPublisher{
			SingleAppPublisher: SingleAppPublisher{
				LogsightAPI: logsightAPI,
				appName:     prepareAppName(cc.App.Name),
			},
			autoCreateApp: cc.App.AutoCreate,
			key:           cc.App.Key,
			existentApps:  make(map[string]struct{}),
		}
	} else {
		publisher = &SingleAppPublisher{
			LogsightAPI: logsightAPI,
			appName:     prepareAppName(cc.App.Name),
		}
		if cc.App.AutoCreate {
			if err := logsightAPI.CreateApp(cc.App.Name); err != nil {
				log.Warnf("failed to create application ")
			}
		}
	}

	client := &Client{
		Publisher: publisher,
		log:       log,
		appName:   cc.App.Name,
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
			c.log.Warnf("%v", err)
			failedCnt += len(logs)
			c.log.Infof("failed to transmit %v messages to app %v", len(logs), app)
		} else {
			c.log.Infof("successfully transmitted %v messages to app %v", len(logs), app)
		}
	}

	st.Acked(len(events) - failedCnt)
	st.Failed(failedCnt)
	batch.ACK()

	return nil
}

type AppPublisher interface {
	groupByApp(logs []common.MapStr) map[string][]string
	send(logs []string, app string) error
}

type SingleAppPublisher struct {
	LogsightAPI Logsight
	appName     string
}

func (sap *SingleAppPublisher) groupByApp(logs []common.MapStr) map[string][]string {
	app := sap.appName

	appLogs := make(map[string][]string)
	appLogs[app] = make([]string, len(logs))
	for i := range logs {
		appLogs[app][i] = fmt.Sprintf("%v", logs[i])
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
	key           string
	existentApps  map[string]struct{}
}

func (mac *MultiAppPublisher) groupByApp(logs []common.MapStr) map[string][]string {
	defaultApp := mac.appName

	// the string representations of all logs are grouped by app names
	// a default app name is used in case of problems
	appLogs := make(map[string][]string)
	for i := range logs {
		logStr := fmt.Sprintf("%v", logs[i])
		if element, err := logs[i].GetValue(mac.key); err != nil {
			appLogs[defaultApp] = append(appLogs[defaultApp], logStr)
		} else {
			if app, ok := element.(string); ok {
				appLogs[app] = append(appLogs[app], logStr)
			} else {
				appLogs[defaultApp] = append(appLogs[defaultApp], logStr)
			}
		}
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
