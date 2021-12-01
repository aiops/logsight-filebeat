package main

import (
	"fmt"
	"github.com/aiops/logsight-filebeat/tmp"
	"github.com/elastic/beats/v7/libbeat/common/transport"
	"github.com/elastic/beats/v7/libbeat/common/transport/tlscommon"
	"github.com/elastic/beats/v7/libbeat/logp"
	"net/http"
	"net/url"
	"time"
)

func main() {
	url, _ := url.Parse("http://wally113.cit.tu-berlin.de:8080")

	var dialer, tlsDialer transport.Dialer
	var tlsConfig *tlscommon.TLSConfig

	dialer = transport.NetDialer(60 * time.Second)
	tlsConfig, _ = tlscommon.LoadTLSConfig(nil)
	tlsDialer = transport.TLSDialer(dialer, tlsConfig, 60 * time.Second)

	api := tmp.Logsight{
		BaseURL:  url,
		Email:    "info@logsight.ai",
		Password: "alex123456",

		Request: tmp.Request{
			Encoder: tmp.NewJSONEncoder(nil),
			Http: &http.Client{
				Transport: &http.Transport{
					Dial:    dialer.Dial,
					DialTLS: tlsDialer.Dial,
					Proxy:   http.ProxyFromEnvironment,
				},
				Timeout: 60 * time.Second,
			},
			Log: logp.NewLogger("logsight"),
		},
	}

	l := logp.NewLogger("test")
	if err := api.Login(); err != nil {
		fmt.Printf("%v", err)
		l.Error(err)
	} else {
		fmt.Printf("Success!!! %v", api.Token)
	}
}
