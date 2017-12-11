// Copyright (c) 2016 Ventu.io, Oleg Sklyar, contributors
// The use of this source code is governed by a MIT style license found in the LICENSE file

package json_test

import (
	"bytes"
	"fmt"
	"github.com/KristinaEtc/slf"
	"github.com/KristinaEtc/slog"
	"github.com/KristinaEtc/slog/json"
	"strings"
	"testing"
	"time"
)

type stringwriter struct {
	res      string
	cutshort bool
	err      error
}

func (sw *stringwriter) Write(p []byte) (n int, err error) {
	sw.res = sw.res + string(p)
	if sw.cutshort {
		return len(p) - 1, sw.err
	} else {
		return len(p), sw.err
	}
}

type interceptor struct {
	entry chan slog.Entry
}

func (i *interceptor) Handle(e slog.Entry) error {
	i.entry <- e
	return nil
}

func TestJSON_standardLogging_success(t *testing.T) {
	lf := slog.New()
	i := &interceptor{entry: make(chan slog.Entry, 1)}
	lf.AddEntryHandler(i)
	lf.SetConcurrent(false)

	f := make(map[string]interface{})
	f["A"] = 25

	lf.WithContext("json").WithField("F", f).WithError(fmt.Errorf("error")).Infof("info=%v", 26)

	sw := &stringwriter{}
	h := json.New(sw)
	if err := h.Handle(<-i.entry); err != nil {
		t.Error(err)
	}
	if !strings.Contains(sw.res, `{"timestamp":"`) {
		t.Errorf("unexpected json, %v", sw.res)
	}
	if !strings.Contains(sw.res, `,"level":"INFO","message":"info=26","error":"error","fields":{"F":{"A":25},"context":"json"}}`) {
		t.Errorf("unexpected json, %v", sw.res)
	}
}

func TestJSON_withoutError_success(t *testing.T) {
	lf := slog.New()
	i := &interceptor{entry: make(chan slog.Entry, 1)}
	lf.AddEntryHandler(i)
	lf.SetConcurrent(false)

	f := make(map[string]interface{})
	f["A"] = 25

	lf.WithContext("json").WithField("F", f).Infof("info=%v", 26)

	sw := &stringwriter{}
	h := json.New(sw)
	if err := h.Handle(<-i.entry); err != nil {
		t.Error(err)
	}
	if !strings.Contains(sw.res, `{"timestamp":"`) {
		t.Errorf("unexpected json, %v", sw.res)
	}
	if !strings.Contains(sw.res, `,"level":"INFO","message":"info=26","fields":{"F":{"A":25},"context":"json"}}`) {
		t.Errorf("unexpected json, %v", sw.res)
	}
}

func TestJSON_withoutFields_success(t *testing.T) {
	lf := slog.New()
	i := &interceptor{entry: make(chan slog.Entry, 1)}
	lf.AddEntryHandler(i)
	lf.SetConcurrent(false)

	lf.WithContext("json").WithError(fmt.Errorf("error")).Infof("info=%v", 26)

	sw := &stringwriter{}
	h := json.New(sw)
	if err := h.Handle(<-i.entry); err != nil {
		t.Error(err)
	}
	if !strings.Contains(sw.res, `{"timestamp":"`) {
		t.Errorf("unexpected json, %v", sw.res)
	}
	if !strings.Contains(sw.res, `,"level":"INFO","message":"info=26","error":"error","fields":{"context":"json"}}`) {
		t.Errorf("unexpected json, %v", sw.res)
	}
}

func TestJSON_withCaller_success(t *testing.T) {
	lf := slog.New()
	i := &interceptor{entry: make(chan slog.Entry, 1)}
	lf.AddEntryHandler(i)
	lf.SetConcurrent(false)

	// beware: line number 116 is expected in the check below
	lf.WithContext("json").WithCaller(slf.CallerShort).Infof("info=%v", 26)

	sw := &stringwriter{}
	h := json.New(sw)
	if err := h.Handle(<-i.entry); err != nil {
		t.Error(err)
	}
	if !strings.Contains(sw.res, `{"timestamp":"`) {
		t.Errorf("unexpected json, %v", sw.res)
	}
	if !strings.Contains(sw.res, `","level":"INFO","message":"info=26","fields":{"caller":"handler_test.go:117","context":"json"}}`) {
		t.Errorf("unexpected json, %v", sw.res)
	}
}

func TestJSON_withTrace_success(t *testing.T) {
	lf := slog.New()
	i := &interceptor{entry: make(chan slog.Entry, 2)}
	lf.AddEntryHandler(i)
	lf.SetConcurrent(false)

	err := fmt.Errorf("error")
	lf.WithContext("json").Infof("info=%v", 26).Trace(&err)

	sw := &stringwriter{}
	h := json.New(sw)
	if err := h.Handle(<-i.entry); err != nil {
		t.Error(err)
	}
	if err := h.Handle(<-i.entry); err != nil {
		t.Error(err)
	}
	if !strings.Contains(sw.res, `","level":"INFO","message":"info=26","fields":{"context":"json"}}{"timestamp":"`) {
		t.Errorf("unexpected json, %v", sw.res)
	}
	if !strings.Contains(sw.res, `","level":"INFO","message":"trace","error":"error","fields":{"context":"json","trace":`) {
		t.Errorf("unexpected json, %v", sw.res)
	}
}

type custom struct{}

func (c custom) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("custom error")
}

func TestJSON_onMarshallError_error(t *testing.T) {
	lf := slog.New()
	i := &interceptor{entry: make(chan slog.Entry, 1)}
	lf.AddEntryHandler(i)
	lf.SetConcurrent(false)

	lf.WithContext("json").WithField("A", custom{}).Infof("info=%v", 26)

	sw := &stringwriter{}
	h := json.New(sw)
	if err := h.Handle(<-i.entry); err == nil || err.Error() != "json: error calling MarshalJSON for type json_test.custom: custom error" {
		t.Errorf("expecting different error, %v", err)
	}
}

func TestJSON_onStreamError_error(t *testing.T) {
	lf := slog.New()
	i := &interceptor{entry: make(chan slog.Entry, 1)}
	lf.AddEntryHandler(i)
	lf.SetConcurrent(false)

	lf.WithContext("json").Infof("info=%v", 26)

	sw := &stringwriter{err: fmt.Errorf("stream error")}
	h := json.New(sw)
	if err := h.Handle(<-i.entry); err == nil || err.Error() != "stream error" {
		t.Errorf("expecting different error, %v", err)
	}
}

func TestJSON_onPartlyWritten_error(t *testing.T) {
	lf := slog.New()
	i := &interceptor{entry: make(chan slog.Entry, 1)}
	lf.AddEntryHandler(i)
	lf.SetConcurrent(false)

	lf.WithContext("json").Infof("info=%v", 26)

	sw := &stringwriter{cutshort: true}
	h := json.New(sw)
	if err := h.Handle(<-i.entry); err == nil || err.Error() != "json.Handler: Wrote only 102 bytes out of 103" {
		t.Errorf("expecting different error, %v", err)
	}
}

func TestHandler_timeformat_success(t *testing.T) {
	lf := slog.New()
	i := &interceptor{entry: make(chan slog.Entry, 1)}
	lf.AddEntryHandler(i)
	lf.SetConcurrent(false)

	lf.WithContext("json").Infof("info=%v", 26)

	sw := &stringwriter{}
	h := json.New(sw)
	h.SetTimeFormat("2006-01-02")
	if err := h.Handle(<-i.entry); err != nil {
		t.Error(err)
	}

	if !strings.Contains(sw.res, time.Now().Format("2006-01-02")) {
		t.Errorf("expected to find today's date, %v", sw.res)
	}
}

func TestHandler_EOL_success(t *testing.T) {
	lf := slog.New()
	i := &interceptor{entry: make(chan slog.Entry, 1)}
	lf.AddEntryHandler(i)
	lf.SetConcurrent(false)

	f := make(map[string]interface{})
	f["A"] = 25

	lf.WithContext("json").WithField("F", f).Infof("info=%v", 26)

	sw := &stringwriter{}
	h := json.New(sw)
	h.SetAddingEOL(true)
	if err := h.Handle(<-i.entry); err != nil {
		t.Error(err)
	}
	if !strings.Contains(sw.res, `"fields":{"F":{"A":25},"context":"json"}}`) {
		t.Errorf("unexpected json, %v", sw.res)
	}
	if bytes.LastIndexByte([]byte(sw.res), byte(10)) < 0 {
		t.Errorf("no EOL, %v", sw.res)
	}
}
