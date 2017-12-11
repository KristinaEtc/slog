// Copyright (c) 2016 Ventu.io, Oleg Sklyar, contributors
// The use of this source code is governed by a MIT style license found in the LICENSE file

package slog_test

import (
	"errors"
	"github.com/KristinaEtc/slf"
	"github.com/KristinaEtc/slog"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"
)

type testhandler struct {
	entries []slog.Entry
	err     error
}

func (th *testhandler) Handle(entry slog.Entry) error {
	th.entries = append(th.entries, entry)
	return th.err
}

func TestVoidHandler_concurrent_performance_1e6Under4s(t *testing.T) {
	testperformance(t, true)
}

func TestVoidHandler_nonconcurrent_performance_1e6Under4s(t *testing.T) {
	testperformance(t, false)
}

func testperformance(t *testing.T, concurrent bool) {
	start := time.Now()
	h := &perfhandler{done: make(chan bool)}
	lf := slog.New()
	lf.AddEntryHandler(h)
	lf.SetConcurrent(concurrent)

	logger0 := lf.WithContext("ctx1").WithField("A", 25)
	logger1 := logger0.WithField("B", 26)
	logger2 := logger0.WithError(errors.New("err"))
	logger3 := lf.WithContext("ctx2").WithCaller(slf.CallerShort)
	log250k := func(logger slf.Logger) {
		for i := 0; i < 250000; i++ {
			logger.Infof("i=%v", i)
		}
	}
	go log250k(logger0)
	go log250k(logger1)
	go log250k(logger2)
	go log250k(logger3)
	<-h.done
	if time.Now().Sub(start) >= time.Second*4 {
		t.Error("logging 1mil records into void handler took more than 4s")
	}
}

func TestVoidHandler_concurrent_contextAqcuisition_1e6Under4s(t *testing.T) {
	start := time.Now()
	h := &perfhandler{done: make(chan bool)}
	lf := slog.New()
	lf.AddEntryHandler(h)
	log50k := func() {
		for i := 0; i < 50000; i++ {
			lf.WithContext(strconv.Itoa(rand.Intn(10000))).Infof("i=%v", i)
		}
	}
	for i := 0; i < 20; i++ {
		go log50k()
	}
	<-h.done
	if time.Now().Sub(start) >= time.Second*4 {
		t.Error("logging 1mil records into void handler took more than 4s")
	}
}

type perfhandler struct {
	sync.Mutex
	count int
	done  chan bool
}

func (h *perfhandler) Handle(e slog.Entry) error {
	h.Lock()
	defer h.Unlock()
	h.count++
	if h.count >= 1000000 {
		h.done <- true
	}
	return nil
}
