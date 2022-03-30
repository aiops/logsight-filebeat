package logsight

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type bulkWriter interface {
	Add(meta, obj interface{}) error
	AddRaw(raw interface{}) error
}

type encoder interface {
	bulkBodyEncoder
	Reader() io.Reader
	Marshal(doc interface{}) error
}

type bulkBodyEncoder interface {
	bulkWriter

	AddHeader(*http.Header, string)
	Reset()
}

type jsonEncoder struct {
	Buf *bytes.Buffer
}

func NewJSONEncoder(buf *bytes.Buffer) *jsonEncoder {
	if buf == nil {
		buf = bytes.NewBuffer(nil)
	}
	return &jsonEncoder{buf}
}

func (b *jsonEncoder) Reset() {
	b.Buf.Reset()
}

func (b *jsonEncoder) AddHeader(header *http.Header, contentType string) {
	if contentType == "" {
		header.Add("Content-Type", "application/json; charset=UTF-8")
	} else {
		header.Add("Content-Type", contentType)
	}
}

func (b *jsonEncoder) Reader() io.Reader {
	return b.Buf
}

func (b *jsonEncoder) Marshal(obj interface{}) error {
	b.Reset()
	enc := json.NewEncoder(b.Buf)
	return enc.Encode(obj)
}

func (b *jsonEncoder) AddRaw(raw interface{}) error {
	enc := json.NewEncoder(b.Buf)
	return enc.Encode(raw)
}

func (b *jsonEncoder) Add(meta, obj interface{}) error {
	enc := json.NewEncoder(b.Buf)
	pos := b.Buf.Len()

	if err := enc.Encode(meta); err != nil {
		b.Buf.Truncate(pos)
		return err
	}
	if err := enc.Encode(obj); err != nil {
		b.Buf.Truncate(pos)
		return err
	}
	return nil
}
