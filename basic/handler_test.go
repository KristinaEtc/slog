// Copyright (c) 2016 Ventu.io, Oleg Sklyar, contributors
// The use of this source code is governed by a MIT style license found in the LICENSE file

package basic_test

import (
	"errors"
	"github.com/KristinaEtc/slf"
	"github.com/KristinaEtc/slog"
	"github.com/KristinaEtc/slog/basic"
	"strings"
	"testing"
	"time"
)

type stringwriter struct {
	res string
	err error
}

func (sw *stringwriter) Write(p []byte) (n int, err error) {
	sw.res = sw.res + string(p)
	return len(p), sw.err
}

func TestHandler_basicOperation_success(t *testing.T) {
	lf := slog.New()
	h := basic.New()
	h.SetTemplate(basic.StandardTextTemplate)
	wr := &stringwriter{}
	h.SetWriter(wr)
	lf.AddEntryHandler(h)
	lf.SetConcurrent(false)

	lf.WithContext("test").WithField("A", 24).Info("done1")
	// XXX the check below assumes next line is No 37
	lf.WithContext("test").WithFields(slf.Fields{"B": 26, "A": 25}).WithCaller(slf.CallerShort).WithError(errors.New("some error")).Warn("done2")
	if !strings.Contains(wr.res, " [INFO] test: done1 A=24\n") || !strings.Contains(wr.res, " [WARN] test (handler_test.go:37): done2 (error: some error) A=25; B=26\n") {
		t.Errorf("no match, %v", wr.res)
	}
}

func TestHandler_basicOperation_termTemplate_success(t *testing.T) {
	lf := slog.New()
	h := basic.New()
	wr := &stringwriter{}
	h.SetWriter(wr)
	lf.AddEntryHandler(h)
	lf.SetConcurrent(false)

	lf.WithContext("test").WithField("A", 24).Info("done1")
	// XXX the check below assumes next line is No 53
	lf.WithContext("test").WithFields(slf.Fields{"B": 26, "A": 25}).WithCaller(slf.CallerShort).WithError(errors.New("some error")).Warn("done2")
	if !strings.Contains(wr.res, " [\033[32mINFO\033[0m] test: done1 A=24\n") {
		t.Errorf("no match, %v", wr.res)
	}
	if !strings.Contains(wr.res, " [\033[33mWARN\033[0m] test (handler_test.go:53): done2 (\033[31merror: some error\033[0m) A=25; B=26\n") {
		t.Errorf("no match, %v", wr.res)
	}
}

func TestHandler_nonCompilableTemplate_error(t *testing.T) {
	h := basic.New()
	if err := h.SetTemplate("Some really buggy template }} {{index . \"bug"); err == nil {
		t.Error("expected an error")
	}
}

func TestHandler_timeformat_success(t *testing.T) {
	lf := slog.New()
	h := basic.New()
	h.SetTemplate(basic.StandardTextTemplate)
	h.SetTimeFormat("2006-01-02")
	wr := &stringwriter{}
	h.SetWriter(wr)
	lf.AddEntryHandler(h)
	lf.SetConcurrent(false)

	lf.WithContext("test").WithField("A", 24).Info("done1")
	if !strings.Contains(wr.res, time.Now().Format("2006-01-02")) {
		t.Errorf("expected to find today's date, %v", wr.res)
	}
}

func TestHandler_setColors_success(t *testing.T) {
	lf := slog.New()
	h := basic.New()
	cols := make(map[slf.Level]int)
	cols[slf.LevelDebug] = 32
	h.SetColors(cols)
	wr := &stringwriter{}
	h.SetWriter(wr)
	lf.AddEntryHandler(h)
	lf.SetConcurrent(false)

	lf.WithContext("test").WithField("A", 24).Info("done1")
	if !strings.Contains(wr.res, " [\033[37mINFO\033[0m] test: done1 A=24\n") {
		t.Errorf("no match, %v", wr.res)
	}
}

func TestHandler_errorNotInTemplate_success(t *testing.T) {
	lf := slog.New()
	h := basic.New()
	if err := h.SetTemplate("[{{.Level}}] {{.Context}}{{if .Caller}} ({{.Caller}}){{end}}: {{.Message}} {{.Fields}}"); err != nil {
		t.Error(err)
	}
	wr := &stringwriter{}
	h.SetWriter(wr)
	lf.AddEntryHandler(h)
	lf.SetConcurrent(false)

	lf.WithContext("test").WithField("A", 24).WithError(errors.New("some error")).Info("done1")
	if wr.res != "[INFO] test: done1 A=24; error=some error\n" {
		t.Errorf("no match, %v", wr.res)
	}
}

type interceptor struct {
	entry chan slog.Entry
}

func (i *interceptor) Handle(e slog.Entry) error {
	i.entry <- e
	return nil
}

func TestHandler_onFormattingError_error(t *testing.T) {
	i := &interceptor{entry: make(chan slog.Entry, 1)}
	lf := slog.New()
	lf.AddEntryHandler(i)
	lf.SetConcurrent(false)
	lf.WithContext("test").Info("done1")

	h := basic.New()
	if err := h.SetTemplate("{{.Foo}}"); err != nil {
		t.Error(err)
	}
	if err := h.Handle(<-i.entry); err == nil {
		t.Error("error expected")
	}
}

type panicwriter struct{}

func (sw *panicwriter) Write(p []byte) (n int, err error) {
	panic("I am panicing")
}

func TestHandler_onPanic_error(t *testing.T) {
	i := &interceptor{entry: make(chan slog.Entry, 1)}
	lf := slog.New()
	lf.AddEntryHandler(i)
	lf.SetConcurrent(false)
	lf.WithContext("test").Info("done1")

	h := basic.New()
	h.SetWriter(&panicwriter{})
	if err := h.Handle(<-i.entry); err == nil || err.Error() != "I am panicing" {
		t.Errorf("error expected, %v", err)
	}
}
