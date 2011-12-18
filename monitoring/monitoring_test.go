// Tideland Common Go Library - Monitoring - Unit Tests
//
// Copyright (C) 2009-2011 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed 
// by the new BSD license.

package monitoring

//--------------------
// IMPORTS
//--------------------

import (
	"fmt"
	"rand"
	"testing"
)

//--------------------
// TESTS
//--------------------

// Test of the ETM monitor.
func TestEtmMonitor(t *testing.T) {
	mon := Monitor()

	// Generate measurings.
	for i := 0; i < 500; i++ {
		n := rand.Intn(10)
		id := fmt.Sprintf("Work %d", n)
		m := mon.BeginMeasuring(id)

		work(n * 5000)

		m.EndMeasuring()
	}

	// Print, process with error, and print again.
	mon.MeasuringPointsPrintAll()

	mon.MeasuringPointsDo(func(mp *MeasuringPoint) {
		if mp.Count >= 25 {
			// Divide by zero.
			mp.Count = mp.Count / (mp.Count - mp.Count)
		}
	})

	mon.MeasuringPointsPrintAll()
}

// Test of the SSI monitor.
func TestSsiMonitor(t *testing.T) {
	mon := Monitor()

	// Generate values.
	for i := 0; i < 500; i++ {
		n := rand.Intn(10)
		id := fmt.Sprintf("Work %d", n)

		mon.SetValue(id, rand.Int63n(2001)-1000)
	}

	// Print, process with error, and print again.
	mon.StaySetVariablesPrintAll()

	mon.StaySetVariablesDo(func(ssv *StaySetVariable) {
		if ssv.Count >= 25 {
			// Divide by zero.
			ssv.Count = ssv.Count / (ssv.Count - ssv.Count)
		}
	})

	mon.StaySetVariablesPrintAll()
}

// Test of the DSR monitor.
func TestDsrMonitor(t *testing.T) {
	mon := Monitor()

	mon.Register("monitor:dsr:a", func() string { return "A" })
	mon.Register("monitor:dsr:b", func() string { return "4711" })
	mon.Register("monitor:dsr:c", func() string { return "2011-05-07" })

	mon.DynamicStatusValuesPrintAll()
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
