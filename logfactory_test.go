// Copyright (c) 2016 Ventu.io, Oleg Sklyar, contributors
// The use of this source code is governed by a MIT style license found in the LICENSE file

package slog_test

import (
	"github.com/KristinaEtc/slf"
	"github.com/KristinaEtc/slog"
	"testing"
)

func TestLogger_initialized_success(t *testing.T) {
	lf := slog.New()
	if len(lf.Contexts()) != 0 {
		t.Error("expected no contexts")
	}
}

func TestLogger_withContext_success(t *testing.T) {
	lf := slog.New()
	logger0 := lf.WithContext("test0")
	logger1 := lf.WithContext("test0")
	logger2 := lf.WithContext("test2")
	if logger0 != lf.Contexts()["test0"] {
		t.Error("expected logger for test0")
	}
	if logger1 != lf.Contexts()["test0"] {
		t.Error("expected logger for test0")
	}
	if logger2 != lf.Contexts()["test2"] {
		t.Error("expected logger for test2")
	}
	if len(lf.Contexts()) != 2 {
		t.Errorf("expected 2 contexts, found %v", len(lf.Contexts()))
	}
}

func TestLogger_setLevelWithoutContext_resetsExistingAndRoot_success(t *testing.T) {
	th := &testhandler{}
	lf := slog.New()
	lf.AddEntryHandler(th)
	lf.SetConcurrent(false)

	logger := lf.WithContext("test0")
	logger.Debug("debug1")
	logger.Info("info1")
	if len(th.entries) != 1 || th.entries[0].Message() != "info1" {
		t.Errorf("incorrect log entries found, %v", th.entries)
	}
	th.entries = nil
	// resets all, of the above logger and for the newly created ones
	lf.SetLevel(slf.LevelDebug)
	// this logger is affected
	logger.Debug("debug1")
	logger.Info("info1")
	if len(th.entries) != 2 || th.entries[0].Message() != "debug1" || th.entries[1].Message() != "info1" {
		t.Errorf("incorrect log entries found, %v", th.entries)
	}
	th.entries = nil
	// this logger is also affected
	logger = lf.WithContext("test1")
	logger.Debug("debug1")
	logger.Info("info1")
	if len(th.entries) != 2 || th.entries[0].Message() != "debug1" || th.entries[1].Message() != "info1" {
		t.Errorf("incorrect log entries found, %v", th.entries)
	}
}

func TestLogger_setLevelWithContexts_success(t *testing.T) {
	th := &testhandler{}
	lf := slog.New()
	lf.AddEntryHandler(th)
	lf.SetConcurrent(false)

	logger1 := lf.WithContext("test1")
	logger2 := lf.WithContext("test2")
	lf.SetLevel(slf.LevelDebug, "test1", "root")
	logger1.Debug("debug1")
	logger1.Info("info1")
	logger2.Debug("debug2")
	logger2.Info("info2")
	logger3 := lf.WithContext("test3")
	logger3.Debug("debug3")
	logger3.Info("info3")
	if len(th.entries) != 5 || th.entries[0].Message() != "debug1" ||
		th.entries[1].Message() != "info1" || th.entries[2].Message() != "info2" ||
		th.entries[3].Message() != "debug3" || th.entries[4].Message() != "info3" {
		t.Errorf("incorrect log entries found, %v", th.entries)
	}
}

func TestLogger_addEntryHandlers_success(t *testing.T) {
	th := &testhandler{}
	lf := slog.New()
	lf.AddEntryHandler(th)
	lf.SetEntryHandlers([]slog.EntryHandler{th, th}...)
	lf.AddEntryHandler(th)
	lf.SetConcurrent(false)

	lf.WithContext("test0").Info("test")
	if len(th.entries) != 3 {
		t.Error("3 entries expected")
	}
}

func TestLogger_setCallerInfoWithoutContext_resetsExistingAndRoot_success(t *testing.T) {
	th := &testhandler{}
	lf := slog.New()
	lf.AddEntryHandler(th)
	lf.SetConcurrent(false)

	logger := lf.WithContext("test0")
	logger.Info("message")
	if th.entries[0].Fields()[slog.CallerField] != nil {
		t.Error("unexpected caller")
	}

	th.entries = nil
	// resets all, of the above logger and for the newly created ones
	lf.SetCallerInfo(slf.CallerShort)
	// this logger is affected
	logger.Info("info")
	if th.entries[0].Fields()[slog.CallerField] == nil {
		t.Error("caller expected")
	}

	th.entries = nil
	// this new logger is also affected
	logger = lf.WithContext("test1")
	logger.Info("info")
	if th.entries[0].Fields()[slog.CallerField] == nil {
		t.Error("caller expected")
	}
}

func TestLogger_setCallerInfoWithContexts_success(t *testing.T) {
	th := &testhandler{}
	lf := slog.New()
	lf.AddEntryHandler(th)
	lf.SetConcurrent(false)

	logger1 := lf.WithContext("test1")
	logger2 := lf.WithContext("test2")
	lf.SetCallerInfo(slf.CallerShort, "test1", "root")
	logger1.Info("info1")
	logger2.Info("info2")
	logger3 := lf.WithContext("test3")
	logger3.Info("info3")
	if th.entries[0].Fields()[slog.CallerField] == nil {
		t.Error("caller expected for logger1")
	}
	if th.entries[1].Fields()[slog.CallerField] != nil {
		t.Error("unexpected caller for logger2")
	}
	if th.entries[2].Fields()[slog.CallerField] == nil {
		t.Error("caller expected for logger3")
	}

}
