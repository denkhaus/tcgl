// Tideland Common Go Library - Cells - Input/Output
//
// Copyright (C) 2010-2012 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed 
// by the new BSD license.

package cells

//--------------------
// IMPORTS
//--------------------

import (
	"time"
)

//--------------------
// INPUT
//--------------------

// Input is a simple input for events and to connect
// handlers like cells to.
type Input struct {
	subscriptions   subscriptionMap
	eventChan       EventChannel
	subscribeChan   chan *subscription
	unsubscribeChan chan *subscription
	stopChan        chan bool
}

// NewInput creates an input.
func NewInput(ecLen int) *Input {
	i := &Input{
		subscriptions:   newSubscriptionMap(),
		eventChan:       make(EventChannel, ecLen),
		subscribeChan:   make(chan *subscription),
		unsubscribeChan: make(chan *subscription),
		stopChan:        make(chan bool),
	}
	go i.backend()
	return i
}

// HandleEvent handles an event.
func (i *Input) HandleEvent(e Event) {
	i.eventChan <- e
}

// Subscribe a handler to this input.
func (i *Input) Subscribe(id string, h Handler) {
	i.subscribeChan <- &subscription{id, h}
}

// Unsubscribe a handler from this input.
func (i *Input) Unsubscribe(id string) {
	i.unsubscribeChan <- &subscription{id, nil}
}

// Stop the input.
func (i *Input) Stop() {
	i.stopChan <- true
}

// backend goroutine of the input.
func (i *Input) backend() {
	// Add an error handler. Here simply do a restart.
	defer func() {
		if err := recover(); err != nil {
			go i.backend()
		}
	}()
	// Main event loop.
	for {
		select {
		case e := <-i.eventChan:
			// Distribute an event.
			i.subscriptions.handleEvent(e)
		case s := <-i.subscribeChan:
			// Connect a new handler.
			i.subscriptions.subscribe(s)
		case s := <-i.unsubscribeChan:
			// Disconnect a handler.
			i.subscriptions.unsubscribe(s)
		case <-i.stopChan:
			// Received stop signal.
			return
		}
	}
}

//--------------------
// TICKER INPUT
//--------------------

// TickerEvent signals a tick to ticker subscribers.
type TickerEvent struct {
	id   string
	time time.Time
}

// Topic returns the topic of the event, here "ticker([id])".
func (t TickerEvent) Topic() string {
	return "ticker(" + t.id + ")"
}

// Targets returns the targets of the events, here nil for all
// subscribers.
func (t TickerEvent) Targets() []string {
	return nil
}

// Payload returns the payload of the event, here the time in
// nanoseconds.
func (t TickerEvent) Payload() interface{} {
	return t.time
}

// Ticker delivers TickerEvents at defined intervals to
// a given input.
type Ticker struct {
	id       string
	input    *Input
	ticker   *time.Ticker
	stopChan chan bool
}

// NewTicker creates a new ticker.
func NewTicker(id string, d time.Duration, i *Input) *Ticker {
	t := &Ticker{
		id:       id,
		input:    i,
		ticker:   time.NewTicker(d),
		stopChan: make(chan bool),
	}
	go t.backend()
	return t
}

// Stop the input.
func (t *Ticker) Stop() {
	t.stopChan <- true
}

// Backend of the ticker.
func (t *Ticker) backend() {
	// Add an error handler.
	defer func() {
		if err := recover(); err != nil {
			// Just restart.
			logger.Warningf("restarting ticker after error: %v", err)
			go t.backend()
		}
	}()
	// Main event loop.
	for {
		select {
		case tick := <-t.ticker.C:
			// Received ticker signal, send event to input.
			t.input.HandleEvent(&TickerEvent{t.id, tick})
		case <-t.stopChan:
			// Received stop signal.
			t.ticker.Stop()
			return
		}
	}
}

//--------------------
// OUTPUT TYPES
//--------------------

// HandlerFunc is a simple one-way event handler function.
type HandlerFunc func(e Event)

// HandleEvent fulfills the handler interface for the handler func.
func (hf HandlerFunc) HandleEvent(e Event) {
	hf(e)
}

// FunctionOutput takes simple handler functions
// for output processing.
type FunctionOutput []HandlerFunc

// NewFunctionOutput creates a new function output.
func NewFunctionOutput() FunctionOutput {
	return make(FunctionOutput, 0)
}

// Add a handler function.
func (fo FunctionOutput) Add(hf HandlerFunc) {
	fo = append(fo, hf)
}

// HandleEvent takes an event and let the handlers handle it.
func (fo FunctionOutput) HandleEvent(e Event) {
	for _, h := range fo {
		fo.secureHandleEvent(h, e)
	}
}

// secureHandleEvent provides a secure wrapper of a handler call
// with logging in case of an error.
func (fo FunctionOutput) secureHandleEvent(h Handler, e Event) {
	defer func() {
		if err := recover(); err != nil {
			logger.Errorf("cannot handle event '%s': %v", e.Topic(), err)
		}
	}()
	h.HandleEvent(e)
}

// NewLoggingFunctionOutput creates a function output with logging as first handler.
func NewLoggingFunctionOutput(id string) FunctionOutput {
	fo := NewFunctionOutput()
	fo.Add(func(e Event) { 
		logger.Infof("[%s] topic: '%s' / payload: '%v'", id, e.Topic(), e.Payload())
	})
	return fo
}

// EOF
