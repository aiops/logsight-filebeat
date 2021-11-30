package logsight

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/beats/v7/libbeat/common/transport"
	"github.com/elastic/beats/v7/libbeat/common/transport/tlscommon"
	"github.com/elastic/beats/v7/libbeat/logp"
	"github.com/elastic/beats/v7/libbeat/outputs"
	"github.com/elastic/beats/v7/libbeat/outputs/outil"
	"github.com/elastic/beats/v7/libbeat/publisher"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"time"
)

// ClientSettings struct
type ClientSettings struct {
	URL              string
	Proxy            *url.URL
	TLS              *tlscommon.TLSConfig
	Username         string
	Password         string
	Parameters       map[string]string
	Index            outil.Selector
	Pipeline         *outil.Selector
	Timeout          time.Duration
	CompressionLevel int
	Observer         outputs.Observer
	BatchPublish     bool
	Headers          map[string]string
	ContentType      string
	Format           string
}

// Client struct
type Client struct {
	settings    ClientSettings
	params      map[string]string
	log         *logp.Logger
	httpClient  *http.Client
	encoder     bodyEncoder
	ContentType string
	connected   bool
}

type eventRaw map[string]json.RawMessage

// NewClient instantiate a client.
func NewClient(s ClientSettings) (*Client, error) {
	proxy := http.ProxyFromEnvironment
	if s.Proxy != nil {
		proxy = http.ProxyURL(s.Proxy)
	}
	// logger.Info("HTTP URL: %s", s.URL)
	var dialer, tlsDialer transport.Dialer
	var err error

	dialer = transport.NetDialer(s.Timeout)
	tlsDialer = transport.TLSDialer(dialer, s.TLS, s.Timeout)

	if st := s.Observer; st != nil {
		dialer = transport.StatsDialer(dialer, st)
		tlsDialer = transport.StatsDialer(tlsDialer, st)
	}
	params := s.Parameters
	var encoder bodyEncoder
	compression := s.CompressionLevel
	if compression == 0 {
		switch s.Format {
		case "json":
			encoder = newJSONEncoder(nil)
		case "json_lines":
			encoder = newJSONLinesEncoder(nil)
		}
	} else {
		switch s.Format {
		case "json":
			encoder, err = newGzipEncoder(compression, nil)
		case "json_lines":
			encoder, err = newGzipLinesEncoder(compression, nil)
		}
		if err != nil {
			return nil, err
		}
	}

	client := &Client{
		settings: s,
		params:   params,
		log:      logp.NewLogger("logsight"),
		httpClient: &http.Client{
			Transport: &http.Transport{
				Dial:    dialer.Dial,
				DialTLS: tlsDialer.Dial,
				Proxy:   proxy,
			},
			Timeout: s.Timeout,
		},
		encoder:     encoder,
		ContentType: s.ContentType,
		connected:   false,
	}

	return client, nil
}

// Connect establishes a connection to the clients sink.
func (c *Client) Connect() error {
	log.Info(fmt.Sprintf("Connected"))
	c.connected = true
	return nil
}

// Close closes a connection.
func (c *Client) Close() error {
	c.connected = false
	return nil
}

func (c *Client) String() string {
	return c.settings.URL
}

// Publish sends events to the clients sink.
func (c *Client) Publish(_ context.Context, batch publisher.Batch) error {
	log.Info("Publish")
	st := c.settings.Observer
	events := batch.Events()
	st.NewBatch(len(events))

	processed := 0
	for _ = range events {
		processed++
	}
	batch.ACK()

	log.Info(fmt.Sprintf("Processes %v messages\n", processed))

	return nil
}
