// Tideland Common Go Library - Cells - Behaviors
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
	"code.google.com/p/tcgl/applog"
)

//--------------------
// DEBUG BEHAVIOR
//--------------------

// LogBehavior can be subscribed to cells which emitted event
// will be logged with info level.
type LogBehavior struct {
	id string
}

// NewLogBehavior creates a debug behavior.
func NewLogBehavior() *LogBehavior {
	return &LogBehavior{}
}

// Init the behavior.
func (db *LogBehavior) Init(env *Environment, id string) error {
	db.id = id
	return nil
}

// ProcessEvent processes an event.
func (db *LogBehavior) ProcessEvent(e Event, emitChan EventChannel) {
	applog.Infof("behavior: '%s' event topic: '%s' payload: '%v'", db.id, e.Topic(), e.Payload())
}

// Recover from an error. Can't even log, it's a logging problem.
func (db *LogBehavior) Recover(err interface{}, e Event) {}

// Stop the behavior.
func (db *LogBehavior) Stop() error {
	return nil
}

//--------------------
// BROADCAST BEHAVIOR
//--------------------

// BroadcastBehavior just emits a received event. It's intended to work
// as an entry point vor events, which shall be immediately processed by
// several subscribers.
type BroadcastBehavior struct{}

// NewBroadcastBehavior creates a broadcast behavior.
func NewBroadcastBehavior() *BroadcastBehavior {
	return &BroadcastBehavior{}
}

// Init the behavior.
func (db *BroadcastBehavior) Init(env *Environment, id string) error {
	return nil
}

// ProcessEvent processes an event.
func (db *BroadcastBehavior) ProcessEvent(e Event, emitChan EventChannel) {
	emitChan <- e
}

// Recover from an error.
func (db *BroadcastBehavior) Recover(err interface{}, e Event) {}

// Stop the behavior.
func (db *BroadcastBehavior) Stop() error {
	return nil
}

//--------------------
// POOL BEHAVIOR
//--------------------

type PoolBehavior struct {
	cellPool chan *cell
	factory  func() Behavior
	ps       int
}

func NewPoolBehavior(f func() Behavior, ps int) *PoolBehavior {
	return &PoolBehavior{make(chan *cell, ps), f, ps}
}

// Init the behavior.
func (pb *PoolBehavior) Init(env *Environment, id string) error {
	for i := 0; i < pb.ps; i++ {
		c, err := newCell(env, id, pb.factory(), pb.ps)
		if err != nil {
			return err
		}
		pb.cellPool <- c
	}
	return nil
}

// ProcessEvent processes an event.
func (pb *PoolBehavior) ProcessEvent(e Event, emitChan EventChannel) {
	c := <-pb.cellPool

	c.processEvent(e)

	pb.cellPool <- c
}

// Recover from an error.
func (pb *PoolBehavior) Recover(err interface{}, e Event) {}

// Stop the behavior.
func (pb *PoolBehavior) Stop() error {
	for i := 0; i < pb.ps; i++ {
		c := <-pb.cellPool
		if err := c.stop(); err != nil {
			return err
		}
	}
	close(pb.cellPool)
	return nil
}

//--------------------
// SIMPLE ACTION BEHAVIOR
//--------------------

// SimpleActionFunc is a function type for simple event handling. 
// To use any function with this signature as cell just do 
// AddCell("...", SimpleActionFunc(myFunc), ...).
type SimpleActionFunc func(e Event, emitChan EventChannel)

// Init the behavior.
func (saf SimpleActionFunc) Init(env *Environment, id string) error {
	return nil
}

// ProcessEvent fulfills the behavior interface for the simple
// action.
func (saf SimpleActionFunc) ProcessEvent(e Event, emitChan EventChannel) {
	saf(e, emitChan)
}

// Recover from an error.
func (saf SimpleActionFunc) Recover(err interface{}, e Event) {
	applog.Infof("cells", "cannot perform simple action func: '%v'", err)
}

// Stop the behavior.
func (saf SimpleActionFunc) Stop() error {
	return nil
}

//--------------------
// FILTERED SIMPLE ACTION BEHAVIOR
//--------------------

// Filter is a function type checking if an event shall be handled.
type FilterFunc func(e Event) bool

// FilteredSimpleActionBehavior takes a function for
// the processing of an event.
type FilteredSimpleActionBehavior struct {
	filter FilterFunc
	action SimpleActionFunc
}

// NewFilteredSimpleActionBehavior creates a filtered simple action cell behavior.
func NewFilteredSimpleActionBehavior(f FilterFunc, a SimpleActionFunc) *FilteredSimpleActionBehavior {
	return &FilteredSimpleActionBehavior{f, a}
}

// Init the behavior.
func (fsab *FilteredSimpleActionBehavior) Init(env *Environment, id string) error {
	return nil
}

// ProcessEvent processes an event.
func (fsab *FilteredSimpleActionBehavior) ProcessEvent(e Event, emitChan EventChannel) {
	if fsab.filter(e) {
		fsab.action(e, emitChan)
	}
}

// Recover from an error.
func (fsab *FilteredSimpleActionBehavior) Recover(err interface{}, e Event) {}

// Stop the behavior.
func (fsab *FilteredSimpleActionBehavior) Stop() error {
	return nil
}

//--------------------
// COUNTER BEHAVIOR
//--------------------

// CounterFunc is the signature of a function which analyzis
// an event and returns, which counters shall be incremented.
type CounterFunc func(e Event) []string

type CounterBehavior struct {
	counterFunc CounterFunc
	counters    map[string]int
}

// NewCounterBehavior creates a counter behavior with the given
// counter function.
func NewCounterBehavior(cf CounterFunc) *CounterBehavior {
	return &CounterBehavior{cf, make(map[string]int)}
}

// Init the behavior.
func (cb *CounterBehavior) Init(env *Environment, id string) error {
	return nil
}

// ProcessEvent processes an event.
func (cb *CounterBehavior) ProcessEvent(e Event, emitChan EventChannel) {
	cids := cb.counterFunc(e)
	if cids != nil {
		for _, cid := range cids {
			v, ok := cb.counters[cid]
			if ok {
				cb.counters[cid] = v + 1
			} else {
				cb.counters[cid] = 1
			}
			emitChan <- NewSimpleEvent("counter:"+cid, cb.counters[cid])
		}
	}
}

// Recover from an error.
func (cb *CounterBehavior) Recover(err interface{}, e Event) {}

// Stop the behavior.
func (cb *CounterBehavior) Stop() error {
	return nil
}

func (cb *CounterBehavior) Counter(id string) int {
	if v, ok := cb.counters[id]; ok {
		return v
	}
	return -1
}

//--------------------
// THRESHOLD BEHAVIOR
//--------------------

// ThresholdEvent signals any threshold passing or value changing.
type ThresholdEvent struct {
	reason    string
	counter   int64
	threshold int64
	context   *Context
}

// Topic returns the topic of the event, here "threshold([reason])".
func (te ThresholdEvent) Topic() string {
	return "threshold(" + te.reason + ")"
}

// Payload return the payload as an array with counter and threshold.
func (te ThresholdEvent) Payload() interface{} {
	return [2]int64{te.counter, te.threshold}
}

// Context returns the context of a set of event processings.
func (te ThresholdEvent) Context() *Context {
	return te.context
}

// SetContext set the context of a set of event processings.
func (te *ThresholdEvent) SetContext(c *Context) {
	te.context = c
}

// ThresholdBehavior fires an event if a threshold has been 
// passed depending on the configuration.
type ThresholdBehavior struct {
	initialCounter   int64
	tickerDifference int64
	tickerDirection  int
	counter          int64
	upperThreshold   int64
	lowerThreshold   int64
}

// NewThresholdBehavior creates a threshold event processor.
func NewThresholdBehavior(ic, ut, lt, td int64, dir int) *ThresholdBehavior {
	return &ThresholdBehavior{
		initialCounter:   ic,
		tickerDifference: td,
		tickerDirection:  dir,
		counter:          ic,
		upperThreshold:   ut,
		lowerThreshold:   lt,
	}
}

// Init the behavior.
func (tb *ThresholdBehavior) Init(env *Environment, id string) error {
	return nil
}

// ProcessEvent processes an event.
func (tb *ThresholdBehavior) ProcessEvent(e Event, emitChan EventChannel) {
	if _, ok := e.(*TickerEvent); ok {
		// It's a ticker event.
		switch {
		case tb.tickerDirection > 0:
			// Count up.
			tb.counter += tb.tickerDifference
		case tb.tickerDirection < 0:
			// Count down.
			tb.counter -= tb.tickerDifference
		default:
			// Count towards zero.
			if tb.counter > 0 {
				tb.counter -= tb.tickerDifference
			} else {
				tb.counter += tb.tickerDifference
			}
		}
	} else {
		// Check the payload for counter changing. Accept only
		// integers.
		switch p := e.Payload().(type) {
		case int:
			tb.counter += int64(p)
		case int16:
			tb.counter += int64(p)
		case int32:
			tb.counter += int64(p)
		case int64:
			tb.counter += p
		}
	}
	// Check the counter.
	switch {
	case tb.counter >= tb.upperThreshold:
		emitChan <- &ThresholdEvent{"upper", tb.counter, tb.upperThreshold, nil}
	case tb.counter <= tb.lowerThreshold:
		emitChan <- &ThresholdEvent{"lower", tb.counter, tb.lowerThreshold, nil}
	default:
		emitChan <- &ThresholdEvent{"ticker", tb.counter, 0, nil}
	}
}

// Recover from an error. Counter will be set back to the initial counter.
func (tb *ThresholdBehavior) Recover(err interface{}, e Event) {
	tb.counter = tb.initialCounter
}

// Stop the behavior.
func (tb *ThresholdBehavior) Stop() error {
	return nil
}

// EOF
