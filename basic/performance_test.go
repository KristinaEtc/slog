// Copyright (c) 2016 Ventu.io, Oleg Sklyar, contributors
// The use of this source code is governed by a MIT style license found in the LICENSE file

package basic_test

import (
	"errors"
	"github.com/KristinaEtc/slf"
	"github.com/KristinaEtc/slog"
	"github.com/KristinaEtc/slog/basic"
	"testing"
	"time"
)

type voidwriter struct {
	counter int
	max     int
	done    chan bool
}

func (vw *voidwriter) Write(p []byte) (int, error) {
	vw.counter++
	if vw.counter >= vw.max {
		vw.done <- true
	}
	return len(p), nil
}

func TestBasic_perf_withVoidWriter_1e6Under5s(t *testing.T) {
	lf := slog.New()
	h := basic.New()
	wr := &voidwriter{0, 1000000, make(chan bool)}
	h.SetWriter(wr)
	lf.AddEntryHandler(h)
	slf.Set(lf)
	logger0 := slf.WithContext("ctx1").WithField("A", 25)
	logger1 := logger0.WithField("B", 26)
	logger2 := logger0.WithError(errors.New("err"))
	logger3 := slf.WithContext("ctx2").WithCaller(slf.CallerShort)
	log250k := func(logger slf.Logger) {
		for i := 0; i < 250000; i++ {
			logger.Infof("%v", i)
		}
	}
	start := time.Now()
	go log250k(logger0)
	go log250k(logger1)
	go log250k(logger2)
	go log250k(logger3)
	<-wr.done
	if time.Now().Sub(start) > time.Second*5 {
		t.Error("log handler too slow")
	}
}
