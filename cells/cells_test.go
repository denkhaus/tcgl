// Tideland Common Go Library - Cells - Unit Tests
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
	"code.google.com/p/tcgl/monitoring"
	"strconv"
	"testing"
	"time"
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
	logCell := NewCell(NewLogBehavior("simple:log"), 10)
	sepCell := NewCell(SimpleActionFunc(SeparatorAction), 10)
	evenCell := NewCell(NewFilteredSimpleActionBehavior(EvenFilter, ItoaAction), 10)
	oddCell := NewCell(NewFilteredSimpleActionBehavior(OddFilter, ItoaAction), 10)
	out := NewLoggingFunctionOutput("simple:out")
	// Build network.
	in.Subscribe("seperator", sepCell)
	sepCell.Subscribe("even", evenCell)
	sepCell.Subscribe("odd", oddCell)
	sepCell.Subscribe("log", logCell)
	evenCell.Subscribe("out", out)
	evenCell.Subscribe("log", logCell)
	oddCell.Subscribe("out", out)
	oddCell.Subscribe("log", logCell)
	// Sending some events.
	in.HandleEvent(NewSimpleEvent("integer", nil, 11))
	in.HandleEvent(NewSimpleEvent("integer", nil, 12))
	in.HandleEvent(NewSimpleEvent("integer", nil, 13))
	in.HandleEvent(NewSimpleEvent("integer", nil, "foo"))
	in.HandleEvent(NewSimpleEvent("string", nil, "foo"))
	// Stop all cells.
	time.Sleep(2e9)
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
	ttoiCell := NewCell(SimpleActionFunc(func(e Event, ec EventChannel) { ec <- NewSimpleEvent("int", nil, 1) }), 10)
	logCell := NewCell(NewLogBehavior("threshold:log"), 10)
	thCell := NewCell(NewThresholdBehavior(5, 20, 0, 1, 0), 10)
	// Build network.
	in.Subscribe("log", logCell)
	in.Subscribe("th", thCell)
	in.Subscribe("log", logCell)
	in.Subscribe("ttoi", ttoiCell)
	ttoiCell.Subscribe("th", thCell)
	thCell.Subscribe("log", logCell)
	thCell.Subscribe("log", logCell)
	// Stop all cells.
	time.Sleep(2e9)
	tickerA.Stop()
	tickerB.Stop()
	ttoiCell.Stop()
	logCell.Stop()
	thCell.Stop()
}

// TestMonitoring just prints the measuring values.
func TestMonitoring(t *testing.T) {
	monitoring.MeasuringPointsPrintAll()
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
			ec <- NewSimpleEvent("even", []string{"even"}, d)
		} else {
			ec <- NewSimpleEvent("odd", []string{"odd"}, d)
		}
	}
}

// ItoaAction maps an integer to an asciii string.
func ItoaAction(e Event, ec EventChannel) {
	if d, ok := e.Payload().(int); ok {
		ec <- NewSimpleEvent("itoa:"+e.Topic(), nil, strconv.Itoa(d))
	} else {
		ec <- NewSimpleEvent("illegal-type", nil, e)
	}
}

// EOF
