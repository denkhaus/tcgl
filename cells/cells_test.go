// Tideland Common Go Library - Cells - Unit Tests
//
// Copyright (C) 2010-2011 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed 
// by the new BSD license.

package cells

//--------------------
// IMPORTS
//--------------------

import (
	"os"
	"strconv"
	"testing"
	"time"
	"tcgl.googlecode.com/hg/monitoring"
)

//--------------------
// CONST
//--------------------

//--------------------
// TESTS
//--------------------

// Test simple scenario.
func TestSimpleScenario(t *testing.T) {
	in := NewInput(10)
	logCell := NewCell(NewLogBehavior("simple:log", os.Stdout), 10)
	sepCell := NewCell(SimpleActionFunc(SeparatorAction), 10)
	evenCell := NewCell(NewFilteredSimpleActionBehavior(EvenFilter, ItoaAction), 10)
	oddCell := NewCell(NewFilteredSimpleActionBehavior(OddFilter, ItoaAction), 10)
	out := NewLoggingFunctionOutput("simple:out")

	// Build network.
	in.Subscribe(sepCell)
	sepCell.Subscribe(evenCell)
	sepCell.Subscribe(oddCell)
	sepCell.Subscribe(logCell)
	evenCell.Subscribe(out)
	evenCell.Subscribe(logCell)
	oddCell.Subscribe(out)
	oddCell.Subscribe(logCell)

	// Sending some events.
	in.HandleEvent(NewSimpleEvent("integer", 11))
	in.HandleEvent(NewSimpleEvent("integer", 12))
	in.HandleEvent(NewSimpleEvent("integer", 13))
	in.HandleEvent(NewSimpleEvent("integer", "foo"))
	in.HandleEvent(NewSimpleEvent("string", "foo"))

	time.Sleep(2e9)
	monitoring.Monitor().MeasuringPointsPrintAll()
	time.Sleep(1e9)

	// Stop all cells.
	oddCell.Stop()
	evenCell.Stop()
	sepCell.Stop()
	in.Stop()
}

// Test threshold scenario.
func TestThresholdScenario(t *testing.T) {
	in := NewInput(10)
	tickerA := NewTicker("ticker:a", 1e9, in)
	tickerB := NewTicker("ticker:b", 2e8, in)
	ttoiCell := NewCell(SimpleActionFunc(func(e Event, ec EventChannel) { ec <- NewSimpleEvent("int", 1) }), 10)
	logCell := NewCell(NewLogBehavior("threshold:log", os.Stdout), 10)
	thCell := NewCell(NewThresholdBehavior(5, 20, 0, 1, 0), 10)

	// Build network.
	in.Subscribe(logCell)
	in.Subscribe(thCell)
	in.Subscribe(logCell)
	in.Subscribe(ttoiCell)
	ttoiCell.Subscribe(thCell)
	thCell.Subscribe(logCell)
	thCell.Subscribe(logCell)

	time.Sleep(4e9)
	monitoring.Monitor().MeasuringPointsPrintAll()
	time.Sleep(1e9)

	// Stop all cells.
	tickerA.Stop()
	tickerB.Stop()
	ttoiCell.Stop()
	logCell.Stop()
	thCell.Stop()
}

//--------------------
// HELPERS
//--------------------

// EvenFilter filters even integer.
func EvenFilter(e Event) bool {
	if d, ok := e.Payload().(int); ok {
		return d%2 == 0
	}

	return false
}

// OddFilter filters even integer.
func OddFilter(e Event) bool {
	if d, ok := e.Payload().(int); ok {
		return d%2 != 0
	}

	return false
}

// SeparatorAction seperates integer events in odd and even.
func SeparatorAction(e Event, ec EventChannel) {
	if d, ok := e.Payload().(int); ok {
		if d%2 == 0 {
			ec <- NewSimpleEvent("even", d)
		} else {
			ec <- NewSimpleEvent("odd", d)
		}
	}
}

// ItoaAction maps an integer to an asciii string.
func ItoaAction(e Event, ec EventChannel) {
	if d, ok := e.Payload().(int); ok {
		ec <- NewSimpleEvent("itoa:"+e.Topic(), strconv.Itoa(d))
	}

	ec <- NewSimpleEvent("illegal-type", e)
}

// EOF
