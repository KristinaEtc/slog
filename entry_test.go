// Copyright (c) 2016 Ventu.io, Oleg Sklyar, contributors
// The use of this source code is governed by a MIT style license found in the LICENSE file

package slog_test

import (
	"errors"
	"github.com/KristinaEtc/slf"
	"github.com/KristinaEtc/slog"
	"testing"
	"time"
)

func TestEntry_containsAllData_success(t *testing.T) {
	th := &testhandler{}
	lf := slog.New()
	lf.AddEntryHandler(th)
	lf.SetConcurrent(false)

	lf.WithContext("test").WithField("key", 256).WithError(errors.New("error1")).Error("error2")
	if len(th.entries) != 1 {
		t.Error("unexpected entries")
	}
	rec := th.entries[0]
	if time.Now().Sub(rec.Time()) > time.Millisecond*50 {
		t.Error("unexpected time")
	}
	if rec.Level() != slf.LevelError {
		t.Errorf("unexpected level, %v", rec.Level())
	}
	if rec.Message() != "error2" {
		t.Errorf("unexpected message, %v", rec.Message())
	}
	if rec.Error().Error() != "error1" {
		t.Errorf("unexpected error, %v", rec.Error())
	}
	flds := rec.Fields()
	if ctx, ok := flds[slog.ContextField]; !ok || ctx != "test" {
		t.Errorf("unexpected context, %v", ctx)
	}
	if val, ok := flds["key"]; !ok || val != 256 {
		t.Errorf("unexpected value, %v", val)
	}
	if len(flds) != 2 {
		t.Error("unexpected field length")
	}
}
