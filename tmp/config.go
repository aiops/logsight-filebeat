package tmp

import (
	"time"
)

type httpConfig struct {
	URL              string        `config:"url"`
	Email            string        `config:"email"`
	Password         string        `config:"password"`
	BatchPublish     bool          `config:"batch_publish"`
	BatchSize        int           `config:"batch_size"`
	CompressionLevel int           `config:"compression_level" validate:"min=0, max=9"`
	MaxRetries       int           `config:"max_retries"`
	Timeout time.Duration `config:"timeout"`
	Backoff backoff       `config:"backoff"`
}

type backoff struct {
	Init time.Duration
	Max  time.Duration
}

var (
	defaultConfig = httpConfig{
		URL:              "https://logsight.ai:443",
		Email:            "",
		Password:         "",
		BatchPublish:     false,
		BatchSize:        2048,
		Timeout:          90 * time.Second,
		CompressionLevel: 0,
		MaxRetries:       3,
		Backoff: backoff{
			Init: 1 * time.Second,
			Max:  60 * time.Second,
		},
	}
)
