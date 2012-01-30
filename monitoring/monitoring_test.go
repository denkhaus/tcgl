// Tideland Common Go Library - Monitoring - Unit Tests
//
// Copyright (C) 2009-2012 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed 
// by the new BSD license.

package monitoring

//--------------------
// IMPORTS
//--------------------

import (
	"fmt"
	"math/rand"
	"testing"
)

//--------------------
// TESTS
//--------------------

// Test of the ETM monitor.
func TestEtmMonitor(t *testing.T) {
	// Generate measurings.
	for i := 0; i < 500; i++ {
		n := rand.Intn(10)
		id := fmt.Sprintf("Work %d", n)
		m := BeginMeasuring(id)
		work(n * 5000)
		m.EndMeasuring()
	}
	// Print.
	MeasuringPointsPrintAll()
}

// Test of the SSI monitor.
func TestSsiMonitor(t *testing.T) {
	// Generate values.
	for i := 0; i < 500; i++ {
		n := rand.Intn(10)
		id := fmt.Sprintf("Work %d", n)

		SetValue(id, rand.Int63n(2001)-1000)
	}
	// Print.
	StaySetVariablesPrintAll()
}

// Test of the DSR monitor.
func TestDsrMonitor(t *testing.T) {
	Register("monitor:dsr:a", func() string { return "A" })
	Register("monitor:dsr:b", func() string { return "4711" })
	Register("monitor:dsr:c", func() string { return "2011-05-07" })
	DynamicStatusValuesPrintAll()
}

//--------------------
// HELPERS
//--------------------

// Do some work.
func work(n int) int {
	if n < 0 {
		return 0
	}
	return n * work(n-1)
}

// EOF