// Copyright (c) 2016 Ventu.io, Oleg Sklyar, contributors
// The use of this source code is governed by a MIT style license found in the LICENSE file

package slog_test

import (
	"errors"
	"fmt"
	"github.com/ventu-io/slf"
	"github.com/ventu-io/slog"
	"path"
	"strings"
	"testing"
	"time"
)

func TestLogger_genericLog_success(t *testing.T) {
	th := &testhandler{done: make(chan bool)}
	f := slog.New()
	f.AddEntryHandler(th)
	logger := f.WithContext("ctx")
	logger.Log(slf.LevelDebug, "ignored")
	logger.Log(slf.LevelError, "output")
	<-th.done
	if len(th.entries) != 1 {
		t.Errorf("expected only error output, %v", th.entries)
	}
}

func TestLogger_withField_copies_success(t *testing.T) {
	th := &testhandler{done: make(chan bool)}
	f := slog.New()
	f.AddEntryHandler(th)
	logger0 := f.WithContext("ctx")
	logger1 := logger0.WithField("key", 256)
	logger0.Info("logger0")
	<-th.done
	logger1.Info("logger1")
	<-th.done
	if th.entries[0].Fields()[slog.ContextField] != "ctx" || th.entries[1].Fields()[slog.ContextField] != "ctx" {
		t.Error("expected ctx context in both cases")
	}
	if _, ok := th.entries[0].Fields()["key"]; ok {
		t.Error("unexpected key in logger0")
	}
	if val, ok := th.entries[1].Fields()["key"]; !ok || val != 256 {
		t.Error("expected key and correct value in logger1")
	}
}

func TestLogger_withFields_copies_success(t *testing.T) {
	th := &testhandler{done: make(chan bool)}
	f := slog.New()
	f.AddEntryHandler(th)
	logger0 := f.WithContext("ctx")
	logger1 := logger0.WithFields(slf.Fields{
		"key": 256,
	})
	logger0.Warn("logger0")
	<-th.done
	logger1.Warn("logger1")
	<-th.done
	if th.entries[0].Fields()[slog.ContextField] != "ctx" || th.entries[1].Fields()[slog.ContextField] != "ctx" {
		t.Error("expected ctx context in both cases")
	}
	if _, ok := th.entries[0].Fields()["key"]; ok {
		t.Error("unexpected key in logger0")
	}
	if val, ok := th.entries[1].Fields()["key"]; !ok || val != 256 {
		t.Error("expected key and correct value in logger1")
	}
}

func TestLogger_withError_copies_success(t *testing.T) {
	th := &testhandler{done: make(chan bool)}
	f := slog.New()
	f.AddEntryHandler(th)
	logger := f.WithContext("ctx")
	logger.Error("test0")
	<-th.done
	if th.entries[0].Error() != nil {
		t.Error("no error expected")
	}
	if th.entries[0].Message() != "test0" {
		t.Error("expected warn2")
	}
	th.entries = nil
	logger.WithField("key", 256).WithError(errors.New("error2")).Warn("warn2")
	<-th.done
	if th.entries[0].Error() == nil {
		t.Error("error expected")
	}
	if th.entries[0].Fields()["key"] != 256 {
		t.Error("expected 256")
	}
	if th.entries[0].Message() != "warn2" {
		t.Error("expected warn2")
	}
}

func TestLogger_setLevel_affectsDerivedLogger_success(t *testing.T) {
	th := &testhandler{done: make(chan bool)}
	f := slog.New()
	f.AddEntryHandler(th)
	logger0 := f.WithContext("ctx")
	logger1 := logger0.WithField("key", 256)
	f.SetLevel(slf.LevelDebug)
	logger0.Debug("debug0")
	<-th.done
	logger1.Debug("debug1")
	<-th.done
	if len(th.entries) != 2 {
		t.Error("expected 2 records")
	}
}

func TestLogger_formatting_success(t *testing.T) {
	th := &testhandler{done: make(chan bool)}
	f := slog.New()
	f.AddEntryHandler(th)
	logger := f.WithContext("test")
	logger.Debugf("debug%v", 0)
	logger.WithField("key", 256).Infof("info%v", 1)
	<-th.done
	logger.WithFields(slf.Fields{"key": 256}).Warnf("warn%v", 2)
	<-th.done
	logger.WithError(errors.New("error")).Errorf("error%v", 3)
	<-th.done
	if len(th.entries) != 3 {
		t.Error("expected info, warn and error")
	}
	if th.entries[0].Message() != "info1" || th.entries[1].Message() != "warn2" || th.entries[2].Message() != "error3" {
		t.Error("unexpected messages")
	}
}

func TestLogger_panic_success(t *testing.T) {
	defer func() {
		if r := recover(); r == nil || fmt.Sprintf("%v", r) != "cause panic" {
			t.Error("expected panic")
		}
	}()
	th := &testhandler{done: make(chan bool)}
	f := slog.New()
	f.AddEntryHandler(th)
	f.WithContext("test").Panic("cause panic")
	<-th.done
	if th.entries[0].Message() != "cause panic" {
		t.Error("mesage not preserved")
	}
}

func TestLogger_panicf_success(t *testing.T) {
	defer func() {
		if r := recover(); r == nil || fmt.Sprintf("%v", r) != "cause panic" {
			t.Error("expected panic")
		}
	}()
	th := &testhandler{done: make(chan bool)}
	f := slog.New()
	f.AddEntryHandler(th)
	f.WithContext("test").Panicf("cause %v", "panic")
	<-th.done
	if th.entries[0].Message() != "cause panic" {
		t.Error("mesage not preserved")
	}
}

func TestLogger_trace_success(t *testing.T) {
	th := &testhandler{done: make(chan bool)}
	f := slog.New()
	f.AddEntryHandler(th)
	logger := f.WithContext("test")
	tracer := logger.Info("test1")
	<-th.done
	logger.Error("test2")
	<-th.done
	time.Sleep(time.Millisecond * 100)
	tracer.Trace(nil)
	<-th.done
	if len(th.entries) != 3 {
		t.Error("expected 3 entries")
	}
	dur, ok := th.entries[2].Fields()[slog.TraceField].(time.Duration)
	if !ok {
		t.Error("duration expected")
	}
	if dur < time.Millisecond*100 || dur > time.Second {
		t.Error("unexpected duration")
	}
	if th.entries[2].Level() != slf.LevelError {
		t.Error("unexpected level, must be last logged")
	}
	if th.entries[2].Message() != "trace" {
		t.Error("unexpected message")
	}
}

func TestLogger_trace_withError_overwrites(t *testing.T) {
	err := errors.New("error0")
	th := &testhandler{done: make(chan bool)}
	f := slog.New()
	f.AddEntryHandler(th)
	logger := f.WithContext("test").WithError(errors.New("error1"))
	tracer := logger.Info("test1")
	<-th.done
	time.Sleep(time.Millisecond * 100)
	tracer.Trace(&err)
	<-th.done
	if len(th.entries) != 2 {
		t.Error("expected 2 entries")
	}
	dur, ok := th.entries[1].Fields()[slog.TraceField].(time.Duration)
	if !ok {
		t.Error("duration expected")
	}
	if dur < time.Millisecond*100 || dur > time.Second {
		t.Error("unexpected duration")
	}
	if th.entries[1].Error().Error() != "error0" {
		t.Error("unexpected error")
	}
}

func TestLogger_trace_toLowLelel_ignored(t *testing.T) {
	th := &testhandler{done: make(chan bool)}
	f := slog.New()
	f.AddEntryHandler(th)
	logger := f.WithContext("test").WithError(errors.New("error1"))
	tracer := logger.Info("test1")
	<-th.done
	logger.Debug("void")
	time.Sleep(time.Millisecond * 500)
	tracer.Trace(nil)
	if len(th.entries) != 1 {
		t.Error("expecting neither debug nor trace")
	}
}

func TestLogger_withCallerLong_success(t *testing.T) {
	th := &testhandler{done: make(chan bool)}
	f := slog.New()
	f.AddEntryHandler(th)
	logger := f.WithContext("test").WithCaller(slf.CallerLong)
	info := logger.Info("test1")
	<-th.done
	info.Trace(nil)
	<-th.done
	logger.Infof("test%v", 2)
	<-th.done
	logger.Log(slf.LevelInfo, "test3")
	<-th.done
	callers := make(map[string]struct{})
	for _, e := range th.entries {
		caller := strings.Split(fmt.Sprint(e.Fields()[slog.CallerField]), ":")[0]
		if !strings.Contains(caller, "logger_test.go") {
			t.Errorf("expected to contain logger_test.go, %v", caller)
		}
		callers[caller] = struct{}{}
	}
	if len(callers) != 1 {
		t.Error("different callers detected")
	}
}

func TestLogger_withCallerShort_success(t *testing.T) {
	th := &testhandler{done: make(chan bool)}
	f := slog.New()
	f.AddEntryHandler(th)
	logger := f.WithContext("test").WithCaller(slf.CallerShort)
	info := logger.Info("test1")
	<-th.done
	info.Trace(nil)
	<-th.done
	logger.Infof("test%v", 2)
	<-th.done
	logger.Log(slf.LevelInfo, "test3")
	<-th.done
	callers := make(map[string]struct{})
	for _, e := range th.entries {
		caller := strings.Split(fmt.Sprint(e.Fields()[slog.CallerField]), ":")[0]
		if caller != "logger_test.go" {
			t.Errorf("expected to contain logger_test.go, %v", caller)
		}
		callers[caller] = struct{}{}
	}
	if len(callers) != 1 {
		t.Error("different callers detected")
	}
}

func TestLogger_withCaller_successiveCallOverrides_success(t *testing.T) {
	th := &testhandler{done: make(chan bool)}
	f := slog.New()
	f.AddEntryHandler(th)
	logger0 := f.WithContext("test").WithCaller(slf.CallerLong)
	logger1 := logger0.WithCaller(slf.CallerShort)
	logger2 := logger1.WithCaller(slf.CallerNone)
	logger0.Info("test0")
	<-th.done
	logger1.Info("test1")
	<-th.done
	logger2.Info("test2")
	<-th.done
	caller := strings.Split(fmt.Sprint(th.entries[0].Fields()[slog.CallerField]), ":")[0]
	if path.Base(caller) == caller || !strings.Contains(caller, "logger_test.go") {
		t.Errorf("expected long caller, %v", caller)
	}
	caller = strings.Split(fmt.Sprint(th.entries[1].Fields()[slog.CallerField]), ":")[0]
	if caller != "logger_test.go" {
		t.Errorf("expected short caller, %v", caller)
	}
	if _, ok := th.entries[2].Fields()[slog.CallerField]; ok {
		t.Error("expected no caller")
	}
}