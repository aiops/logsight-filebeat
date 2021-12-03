package logsight

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
)

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

type bulkWriter interface {
	Add(meta, obj interface{}) error
	AddRaw(raw interface{}) error
}

type RawEncoder struct {
	Buf *bytes.Buffer
}

type JsonEncoder struct {
	Buf *bytes.Buffer
}

type JsonLinesEncoder struct {
	Buf *bytes.Buffer
}

type gzipEncoder struct {
	buf  *bytes.Buffer
	gzip *gzip.Writer
}

type gzipLinesEncoder struct {
	buf  *bytes.Buffer
	gzip *gzip.Writer
}

func NewRawEncoder(buf *bytes.Buffer) *RawEncoder {
	if buf == nil {
		buf = bytes.NewBuffer(nil)
	}
	return &RawEncoder{buf}
}

func (b *RawEncoder) Reset() {
	b.Buf.Reset()
}

func (b *RawEncoder) AddHeader(header *http.Header, contentType string) {
	if contentType == "" {
		header.Add("Content-Type", "application/json; charset=UTF-8")
	} else {
		header.Add("Content-Type", contentType)
	}
}

func (b *RawEncoder) Reader() io.Reader {
	return b.Buf
}

func (b *RawEncoder) Marshal(obj interface{}) error {
	b.Reset()
	enc := json.NewEncoder(b.Buf)
	return enc.Encode(obj)
}

func (b *RawEncoder) AddRaw(raw interface{}) error {
	enc := json.NewEncoder(b.Buf)
	return enc.Encode(raw)
}

func (b *RawEncoder) Add(meta, obj interface{}) error {
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

func NewJSONEncoder(buf *bytes.Buffer) *JsonEncoder {
	if buf == nil {
		buf = bytes.NewBuffer(nil)
	}
	return &JsonEncoder{buf}
}

func (b *JsonEncoder) Reset() {
	b.Buf.Reset()
}

func (b *JsonEncoder) AddHeader(header *http.Header, contentType string) {
	if contentType == "" {
		header.Add("Content-Type", "application/json; charset=UTF-8")
	} else {
		header.Add("Content-Type", contentType)
	}
}

func (b *JsonEncoder) Reader() io.Reader {
	return b.Buf
}

func (b *JsonEncoder) Marshal(obj interface{}) error {
	b.Reset()
	enc := json.NewEncoder(b.Buf)
	return enc.Encode(obj)
}

func (b *JsonEncoder) AddRaw(raw interface{}) error {
	enc := json.NewEncoder(b.Buf)
	return enc.Encode(raw)
}

func (b *JsonEncoder) Add(meta, obj interface{}) error {
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

func newJSONLinesEncoder(buf *bytes.Buffer) *JsonLinesEncoder {
	if buf == nil {
		buf = bytes.NewBuffer(nil)
	}
	return &JsonLinesEncoder{buf}
}

func (b *JsonLinesEncoder) Reset() {
	b.Buf.Reset()
}

func (b *JsonLinesEncoder) AddHeader(header *http.Header, contentType string) {
	if contentType == "" {
		header.Add("Content-Type", "application/x-ndjson; charset=UTF-8")
	} else {
		header.Add("Content-Type", contentType)
	}
}

func (b *JsonLinesEncoder) Reader() io.Reader {
	return b.Buf
}

func (b *JsonLinesEncoder) Marshal(obj interface{}) error {
	b.Reset()
	return b.AddRaw(obj)
}

func (b *JsonLinesEncoder) AddRaw(obj interface{}) error {
	enc := json.NewEncoder(b.Buf)

	// single event
	if reflect.TypeOf(obj).Kind() == reflect.Map {
		return enc.Encode(obj)
	}

	// batch of events
	for _, item := range obj.([]eventRaw) {
		err := enc.Encode(item)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *JsonLinesEncoder) Add(meta, obj interface{}) error {
	pos := b.Buf.Len()

	if err := b.AddRaw(meta); err != nil {
		b.Buf.Truncate(pos)
		return err
	}
	if err := b.AddRaw(obj); err != nil {
		b.Buf.Truncate(pos)
		return err
	}

	return nil
}

func newGzipEncoder(level int, buf *bytes.Buffer) (*gzipEncoder, error) {
	if buf == nil {
		buf = bytes.NewBuffer(nil)
	}
	w, err := gzip.NewWriterLevel(buf, level)
	if err != nil {
		return nil, err
	}

	return &gzipEncoder{buf, w}, nil
}

func (b *gzipEncoder) Reset() {
	b.buf.Reset()
	b.gzip.Reset(b.buf)
}

func (b *gzipEncoder) Reader() io.Reader {
	b.gzip.Close()
	return b.buf
}

func (b *gzipEncoder) AddHeader(header *http.Header, contentType string) {
	if contentType == "" {
		header.Add("Content-Type", "application/json; charset=UTF-8")
	} else {
		header.Add("Content-Type", contentType)
	}
	header.Add("Content-Encoding", "gzip")
}

func (b *gzipEncoder) Marshal(obj interface{}) error {
	b.Reset()
	enc := json.NewEncoder(b.gzip)
	err := enc.Encode(obj)
	return err
}

func (b *gzipEncoder) AddRaw(raw interface{}) error {
	enc := json.NewEncoder(b.gzip)
	return enc.Encode(raw)
}

func (b *gzipEncoder) Add(meta, obj interface{}) error {
	enc := json.NewEncoder(b.gzip)
	pos := b.buf.Len()

	if err := enc.Encode(meta); err != nil {
		b.buf.Truncate(pos)
		return err
	}
	if err := enc.Encode(obj); err != nil {
		b.buf.Truncate(pos)
		return err
	}

	b.gzip.Flush()
	return nil
}

func newGzipLinesEncoder(level int, buf *bytes.Buffer) (*gzipLinesEncoder, error) {
	if buf == nil {
		buf = bytes.NewBuffer(nil)
	}
	w, err := gzip.NewWriterLevel(buf, level)
	if err != nil {
		return nil, err
	}

	return &gzipLinesEncoder{buf, w}, nil
}

func (b *gzipLinesEncoder) Reset() {
	b.buf.Reset()
	b.gzip.Reset(b.buf)
}

func (b *gzipLinesEncoder) Reader() io.Reader {
	b.gzip.Close()
	return b.buf
}

func (b *gzipLinesEncoder) AddHeader(header *http.Header, contentType string) {
	if contentType == "" {
		header.Add("Content-Type", "application/x-ndjson; charset=UTF-8")
	} else {
		header.Add("Content-Type", contentType)
	}
	header.Add("Content-Encoding", "gzip")
}

func (b *gzipLinesEncoder) Marshal(obj interface{}) error {
	b.Reset()
	return b.AddRaw(obj)
}

func (b *gzipLinesEncoder) AddRaw(obj interface{}) error {
	enc := json.NewEncoder(b.gzip)

	// single event
	if reflect.TypeOf(obj).Kind() == reflect.Map {
		return enc.Encode(obj)
	}

	// batch of events
	for _, item := range obj.([]eventRaw) {
		err := enc.Encode(item)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *gzipLinesEncoder) Add(meta, obj interface{}) error {
	pos := b.buf.Len()

	if err := b.AddRaw(meta); err != nil {
		b.buf.Truncate(pos)
		return err
	}
	if err := b.AddRaw(obj); err != nil {
		b.buf.Truncate(pos)
		return err
	}

	b.gzip.Flush()
	return nil
}
