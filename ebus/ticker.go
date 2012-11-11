// Tideland Common Go Library - Event Bus - Ticker
//
// Copyright (C) 2012 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed 
// by the new BSD license.

package ebus

//--------------------
// IMPORTS
//--------------------

import (
	"sync"
	"time"
)

//--------------------
// FUNCTIONS
//--------------------

// AddTicker adds a new ticker for periodical ticker events.
func AddTicker(id string, period time.Duration, topics ...string) error {
	tickers.mutex.Lock()
	defer tickers.mutex.Unlock()
	if _, ok := tickers.tickers[id]; ok {
		return &DuplicateTickerError{id}
	}
	tickers.tickers[id] = startTicker(id, period, topics...)
	return nil
}

// RemoveTicker removes a periodical ticker event.
func RemoveTicker(id string) error {
	tickers.mutex.Lock()
	defer tickers.mutex.Unlock()
	if ticker, ok := tickers.tickers[id]; ok {
		ticker.stop()
		delete(tickers.tickers, id)
		return nil
	}
	return &TickerNotFoundError{id}
}

//--------------------
// TICKER
//--------------------

// tickers stores all active tickers.
var tickers = struct {
	mutex   sync.Mutex
	tickers map[string]*ticker
}{
	tickers: make(map[string]*ticker),
}

// stopTickers tells all tickers to stop working.
func stopTickers() {
	for id := range tickers.tickers {
		RemoveTicker(id)
	}
}

type Tick struct {
	Id   string
	Time time.Time
}

// ticker emits periodic events.
type ticker struct {
	id       string
	period   time.Duration
	topics   []string
	stopChan chan bool
}

// startTicker starts a new ticker in the background.
func startTicker(id string, period time.Duration, topics ...string) *ticker {
	t := &ticker{id, period, topics, make(chan bool)}
	go t.backend()
	return t
}

// stop lets the backend goroutine stop working.
func (t *ticker) stop() {
	t.stopChan <- true
}

// backend is the goroutine running the ticker.
func (t *ticker) backend() {
	defer func() {
		t.stopChan = nil
	}()
	for {
		select {
		case <-time.After(t.period):
			tick := Tick{t.id, time.Now()}
			for _, topic := range t.topics {
				Emit(tick, topic)
			}
		case <-t.stopChan:
			return
		}
	}
}

// IsTickerEvent checks if an event is a ticker event and returns the tick.
func IsTickerEvent(event Event) (bool, Tick) {
	var tick Tick
	err := event.Payload(&tick)
	if err != nil {
		return false, tick

	}
	return true, tick
}

// EOF
