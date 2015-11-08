// Tideland Common Go Library - Cells - Behavior
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
	"github.com/denkhaus/tcgl/applog"
)

//--------------------
// COLLECTOR BEHAVIOR
//--------------------

// EventCollector defines the interface for a behavior
// collecting events.
type EventCollector interface {
	// Events returns the collected list of events.
	Events() []Event
	// Reset clears the list of events.
	Reset()
}

// collectorBehavior collects events for debugging
type collectorBehavior struct {
	events []Event
}

// CollectorBehaviorFactory creates a collector behavior. It collects 
// all events emitted directly or by subscription. The event is passed
// through.
func CollectorBehaviorFactory() Behavior {
	return &collectorBehavior{[]Event{}}
}

// Init the behavior.
func (b *collectorBehavior) Init(env *Environment, id Id) error {
	return nil
}

// ProcessEvent processes an event.
func (b *collectorBehavior) ProcessEvent(e Event, emitter EventEmitter) {
	b.events = append(b.events, e)
	emitter.Emit(e)
}

// Recover from an error.
func (b *collectorBehavior) Recover(err interface{}, e Event) {}

// Stop the behavior.
func (b *collectorBehavior) Stop() {}

// Event returns the collected events.
func (b *collectorBehavior) Events() []Event {
	return b.events
}

// Reset clears the list of events.
func (b *collectorBehavior) Reset() {
	b.events = []Event{}
}

//--------------------
// LOG BEHAVIOR
//--------------------

// logBehavior is a behaior for the logging of events.
type logBehavior struct {
	id Id
}

// LogBehaviorFactory creates a logging behavior. It logs emitted
// events with info level.
func LogBehaviorFactory() Behavior {
	return &logBehavior{}
}

// Init the behavior.
func (b *logBehavior) Init(env *Environment, id Id) error {
	b.id = id
	return nil
}

// ProcessEvent processes an event.
func (b *logBehavior) ProcessEvent(e Event, emitter EventEmitter) {
	applog.Infof("cell: '%s' event topic: '%s' payload: '%v'", b.id, e.Topic(), e.Payload())
}

// Recover from an error. Can't even log, it's a logging problem.
func (b *logBehavior) Recover(err interface{}, e Event) {}

// Stop the behavior.
func (b *logBehavior) Stop() {}

//--------------------
// BROADCAST BEHAVIOR
//--------------------

// broadcastBehavior is a simple repeater.
type broadcastBehavior struct{}

// BroadcastBehaviorFactory creates a broadcast behavior that just emits every 
// received event. It's intended to work as an entry point for events, which shall 
// be immediately processed by several subscribers.
func BroadcastBehaviorFactory() Behavior {
	return &broadcastBehavior{}
}

// Init the behavior.
func (b *broadcastBehavior) Init(env *Environment, id Id) error {
	return nil
}

// ProcessEvent processes an event.
func (b *broadcastBehavior) ProcessEvent(e Event, emitter EventEmitter) {
	emitter.Emit(e)
}

// Recover from an error.
func (b *broadcastBehavior) Recover(err interface{}, e Event) {}

// Stop the behavior.
func (b *broadcastBehavior) Stop() {}

//--------------------
// FILTERED BROADCAST BEHAVIOR
//--------------------

// Filter is a function type checking if an event shall be broadcasted.
type FilterFunc func(e Event) bool

// filteredBroadcastBehavior is a simple repeater using the filter 
// function.
type filteredBroadcastBehavior struct {
	filterFunc FilterFunc
}

// NewFilteredBroadcastBehaviorFactory creates the constructor for a filtered
// broadcast behavior based on the passed function. It emits every received event
// for which the filter function returns true.
func NewFilteredBroadcastBehaviorFactory(ff FilterFunc) BehaviorFactory {
	return func() Behavior { return &filteredBroadcastBehavior{ff} }
}

// Init the behavior.
func (b *filteredBroadcastBehavior) Init(env *Environment, id Id) error {
	return nil
}

// ProcessEvent processes an event.
func (b *filteredBroadcastBehavior) ProcessEvent(e Event, emitter EventEmitter) {
	if b.filterFunc(e) {
		emitter.Emit(e)
	}
}

// Recover from an error.
func (b *filteredBroadcastBehavior) Recover(err interface{}, e Event) {}

// Stop the behavior.
func (b *filteredBroadcastBehavior) Stop() {}

//--------------------
// SIMPLE ACTION BEHAVIOR
//--------------------

// SimpleActionFunc is a function type for simple event handling. 
type SimpleActionFunc func(e Event, emitter EventEmitter)

// NewSimpleActionBehaviorFactory creates the factory for a simple
// action behavior based on the passed function. It doesn't care for
// init, recovering or stopping, only event processing.
func NewSimpleActionBehaviorFactory(saf SimpleActionFunc) BehaviorFactory {
	return func() Behavior { return &simpeActionBehavior{saf} }
}

// simpleActionBehavior simply processes events based on a function.
type simpeActionBehavior struct {
	simpleActionFunc SimpleActionFunc
}

// Init the behavior.
func (b *simpeActionBehavior) Init(env *Environment, id Id) error {
	return nil
}

// ProcessEvent fulfills the behavior interface for the simple
// action.
func (b *simpeActionBehavior) ProcessEvent(e Event, emitter EventEmitter) {
	b.simpleActionFunc(e, emitter)
}

// Recover from an error.
func (b *simpeActionBehavior) Recover(err interface{}, e Event) {
	applog.Errorf("cells", "cannot perform simple action func: '%v'", err)
}

// Stop the behavior.
func (b *simpeActionBehavior) Stop() {}

//--------------------
// COUNTER BEHAVIOR
//--------------------

// CounterFunc is the signature of a function which analyzis
// an event and returns, which counters shall be incremented.
type CounterFunc func(e Event) []string

// NewCounterBehaviorFactory creates a constructor for a counter behavior
// based on the passed function. It increments and emits those counters named
// by the result of the counter function.
func NewCounterBehaviorFactory(cf CounterFunc) BehaviorFactory {
	return func() Behavior { return &counterBehavior{cf, make(map[string]int64)} }
}

// counterBehavior counts events based on the counter function.
type counterBehavior struct {
	counterFunc CounterFunc
	counters    map[string]int64
}

// Init the behavior.
func (b *counterBehavior) Init(env *Environment, id Id) error {
	return nil
}

// ProcessEvent processes an event.
func (b *counterBehavior) ProcessEvent(e Event, emitter EventEmitter) {
	cids := b.counterFunc(e)
	if cids != nil {
		for _, cid := range cids {
			v, ok := b.counters[cid]
			if ok {
				b.counters[cid] = v + 1
			} else {
				b.counters[cid] = 1
			}
			emitter.EmitSimple("counter:"+cid, b.counters[cid])
		}
	}
}

// Recover from an error.
func (b *counterBehavior) Recover(err interface{}, e Event) {}

// Stop the behavior.
func (b *counterBehavior) Stop() {}

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

// ThresholdBehavior fires an event if the upper or lower threshold has been 
// passed depending on the configuration. A ticker event can also increase
// (direction > 0) or decrease (direction < 0) the counter or move it back to 
// zero (direction = 0).
type thresholdBehavior struct {
	initialCounter   int64
	tickerDifference int64
	tickerDirection  int64
	counter          int64
	upperThreshold   int64
	lowerThreshold   int64
}

// NewThresholdBehaviorFactory creates a constructor for a threshold behavior.
func NewThresholdBehaviorFactory(initialCounter, tickerDifference, tickerDirection, upperThreshold, lowerThreshold int64) BehaviorFactory {
	return func() Behavior {
		return &thresholdBehavior{
			initialCounter,
			tickerDifference,
			tickerDirection,
			initialCounter,
			upperThreshold,
			lowerThreshold,
		}
	}
}

// Init the behavior.
func (b *thresholdBehavior) Init(env *Environment, id Id) error {
	return nil
}

// ProcessEvent processes an event.
func (b *thresholdBehavior) ProcessEvent(e Event, emitter EventEmitter) {
	if _, ok := e.(*TickerEvent); ok {
		// It's a ticker event.
		switch {
		case b.tickerDirection > 0:
			// Count up.
			b.counter += b.tickerDifference
		case b.tickerDirection < 0:
			// Count down.
			b.counter -= b.tickerDifference
		default:
			// Count towards zero.
			if b.counter > 0 {
				b.counter -= b.tickerDifference
			} else {
				b.counter += b.tickerDifference
			}
		}
	} else {
		// Check the payload for counter changing. Accept only
		// integers.
		switch p := e.Payload().(type) {
		case bool:
			if p {
				b.counter++
			} else {
				b.counter--
			}
		case int:
			b.counter += int64(p)
		case int16:
			b.counter += int64(p)
		case int32:
			b.counter += int64(p)
		case int64:
			b.counter += p
		}
	}
	// Check the counter.
	switch {
	case b.counter >= b.upperThreshold:
		emitter.Emit(&ThresholdEvent{"upper", b.counter, b.upperThreshold, nil})
	case b.counter <= b.lowerThreshold:
		emitter.Emit(&ThresholdEvent{"lower", b.counter, b.lowerThreshold, nil})
	default:
		emitter.Emit(&ThresholdEvent{"ticker", b.counter, 0, nil})
	}
}

// Recover from an error. Counter will be set back to the initial counter.
func (b *thresholdBehavior) Recover(err interface{}, e Event) {
	b.counter = b.initialCounter
}

// Stop the behavior.
func (b *thresholdBehavior) Stop() {}

// EOF
