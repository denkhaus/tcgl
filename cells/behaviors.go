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
	"io"
	"log"
)

//--------------------
// LOG CELL BEHAVIOUR
//--------------------

// LogBehavior just logs events and raises nothing.
type LogBehavior struct {
	prefix string
	logger *log.Logger
}

// NewLogBehavior creates a log cell behavior with the
// given writer as target.
func NewLogBehavior(p string, o io.Writer) *LogBehavior {
	return &LogBehavior{
		prefix: p,
		logger: log.New(o, "", log.Ldate|log.Ltime),
	}
}

// ProcessEvent processes an event.
func (lb LogBehavior) ProcessEvent(e Event, emitChan EventChannel) {
	lb.logger.Printf("[%v] event topic: '%s' payload: '%v'", lb.prefix, e.Topic(), e.Payload())
}

// Recover from an error. Can't even log, it's a logging problem.
func (lb *LogBehavior) Recover(err interface{}, e Event) {}

// Stop the behavior.
func (lb *LogBehavior) Stop() {}

//--------------------
// SIMPLE ACTION BEHAVIOUR
//--------------------

// SimpleActionFunc is a function type for simple event handling. 
// To use any function with this signature as cell just do 
// NewCell(SimpleActionFunc(myFunc), ...).
type SimpleActionFunc func(e Event, emitChan EventChannel)

// ProcessEvent fulfills the behavior interface for the simple
// action.
func (saf SimpleActionFunc) ProcessEvent(e Event, emitChan EventChannel) {
	saf(e, emitChan)
}

// Recover from an error.
func (saf SimpleActionFunc) Recover(err interface{}, e Event) {
	log.Printf("[cells] cannot perform simple action func: '%v'", err)
}

// Stop the behavior.
func (saf SimpleActionFunc) Stop() {}

//--------------------
// FILTERED SIMPLE ACTION BEHAVIOUR
//--------------------

// Filter is a function type checking if an event shall be handled.
type Filter func(e Event) bool

// FilteredSimpleActionBehavior takes a function for
// the processing of an event.
type FilteredSimpleActionBehavior struct {
	filter Filter
	action SimpleActionFunc
}

// NewFilteredSimpleActionBehavior creates a filtered simple action cell behavior.
func NewFilteredSimpleActionBehavior(f Filter, a SimpleActionFunc) *FilteredSimpleActionBehavior {
	return &FilteredSimpleActionBehavior{f, a}
}

// ProcessEvent processes an event.
func (fsab FilteredSimpleActionBehavior) ProcessEvent(e Event, emitChan EventChannel) {
	if fsab.filter(e) {
		fsab.action(e, emitChan)
	}
}

// Recover from an error.
func (fsab FilteredSimpleActionBehavior) Recover(err interface{}, e Event) {}

// Stop the behavior.
func (fsab FilteredSimpleActionBehavior) Stop() {}

//--------------------
// STATE ACTION BEHAVIOUR
//--------------------

// StateAction is a function type for the event handling with a state.
type StateActionFunc func(s int, e Event, emitChan EventChannel) int

// StateActionBehavior manages its event handling based on an 
// internal state represented by an integer. The StateActionFunc function
// manages changes.
type StateActionBehavior struct {
	initialState int
	state        int
	action       StateActionFunc
}

// NewStateActionBehavior creates a simple action cell behavior.
func NewStateActionBehavior(s int, a StateActionFunc) *StateActionBehavior {
	return &StateActionBehavior{s, s, a}
}

// ProcessEvent processes an event.
func (sab *StateActionBehavior) ProcessEvent(e Event, emitChan EventChannel) {
	sab.state = sab.action(sab.state, e, emitChan)
}

// Recover from an error. State will be set back to the initial state.
func (sab *StateActionBehavior) Recover(err interface{}, e Event) {
	sab.state = sab.initialState
}

// Stop the behavior.
func (sab *StateActionBehavior) Stop() {}

//--------------------
// THRESHOLD BEHAVIOUR
//--------------------

// ThresholdEvent signals any threshold passing or value changing.
type ThresholdEvent struct {
	reason    string
	counter   int64
	threshold int64
}

// Topic returns the topic of the event, here "threshold([reason])".
func (te ThresholdEvent) Topic() string {
	return "threshold(" + te.reason + ")"
}

// Targets returns the reason of this ticker event.
func (te ThresholdEvent) Targets() []string {
	return []string{te.reason}
}

// Payload return the payload as an array with counter and threshold.
func (te ThresholdEvent) Payload() interface{} {
	return [2]int64{te.counter, te.threshold}
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

// ProcessEvent processes an event.
func (tb *ThresholdBehavior) ProcessEvent(e Event, emitChan EventChannel) {
	if _, ok := e.(TickerEvent); ok {
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
		emitChan <- &ThresholdEvent{"upper", tb.counter, tb.upperThreshold}
	case tb.counter <= tb.lowerThreshold:
		emitChan <- &ThresholdEvent{"lower", tb.counter, tb.lowerThreshold}
	default:
		emitChan <- &ThresholdEvent{"ticker", tb.counter, 0}
	}
}

// Recover from an error. Counter will be set back to the initial counter.
func (tb *ThresholdBehavior) Recover(err interface{}, e Event) {
	tb.counter = tb.initialCounter
}

// Stop the behavior.
func (tb ThresholdBehavior) Stop() {}

// EOF
