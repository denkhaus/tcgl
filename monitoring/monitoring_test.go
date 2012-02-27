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
	"code.google.com/p/tcgl/asserts"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

//--------------------
// TESTS
//--------------------

// Test of the ETM monitor.
func TestEtmMonitor(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	// Generate measurings.
	for i := 0; i < 500; i++ {
		n := rand.Intn(10)
		id := fmt.Sprintf("mp:task:%d", n)
		m := BeginMeasuring(id)
		work(n * 5000)
		m.EndMeasuring()
	}
	// Need some time to let that backend catch up queued mesurings.
	time.Sleep(1e7)
	// Asserts.
	mp, err := ReadMeasuringPoint("foo")
	assert.ErrorMatch(err, `measuring point "foo" does not exist`, "Reading non-existent measuring point.")
	mp, err = ReadMeasuringPoint("mp:task:5")
	assert.Nil(err, "No error expected.")
	assert.Equal(mp.Id, "mp:task:5", "Should get the right one.")
	assert.True(mp.Count > 0, "Should be measured several times.")
	assert.Match(mp.String(), `Measuring Point "mp:task:5" (.*)`, "String representation should look fine.")
	MeasuringPointsDo(func(mp *MeasuringPoint) {
		assert.Match(mp.Id, "mp:task:[0-9]", "Id has to match the pattern.")
		assert.True(mp.MinDuration <= mp.AvgDuration && mp.AvgDuration <= mp.MaxDuration, 
			"Avg should be somewhere between min and max.")
		assert.True(mp.TtlDuration > 0, "Duration should be greater 0.")
	})
}

// Test of the SSI monitor.
func TestSsiMonitor(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	// Generate values.
	for i := 0; i < 500; i++ {
		n := rand.Intn(10)
		id := fmt.Sprintf("ssv:value:%d", n)
		SetVariable(id, rand.Int63n(2001)-1000)
	}
	// Need some time to let that backend catch up queued mesurings.
	time.Sleep(1e7)
	// Asserts.
	ssv, err := ReadVariable("foo")
	assert.ErrorMatch(err, `stay-set variable "foo" does not exist`, "Reading non-existent variable.")
	ssv, err = ReadVariable("ssv:value:5")
	assert.Nil(err, "No error expected.")
	assert.Equal(ssv.Id, "ssv:value:5", "Should get the right one.")
	assert.True(ssv.Count > 0, "Should be set several times.")
	assert.Match(ssv.String(), `Stay-Set Variable "ssv:value:5" (.*)`, "String representation should look fine.")
	StaySetVariablesDo(func(ssv *StaySetVariable) {
		assert.Match(ssv.Id, "ssv:value:[0-9]", "Id has to match the pattern.")
		assert.True(ssv.MinValue <= ssv.AvgValue && ssv.AvgValue <= ssv.MaxValue, 
			"Avg should be somewhere between min and max.")
	})
}

// Test of the DSR monitor.
func TestDsrMonitor(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	// Register monitoring funcs.
	Register("dsr:a", func() string { return "A" })
	Register("dsr:b", func() string { return "4711" })
	Register("dsr:c", func() string { return "2012-02-15" })
	Register("dsr:d", func() string { a := 1; a = a / (a-a); return fmt.Sprintf("%d", a) })
	// Need some time to let that backend catch up queued registerings.
	time.Sleep(1e7)
	// Asserts.
	dsv, err := ReadStatus("foo")
	assert.ErrorMatch(err, `dynamic status "foo" does not exist`, "Reading non-existent status.")
	dsv, err = ReadStatus("dsr:b")
	assert.Nil(err, "No error expected.")
	assert.Equal(dsv, "4711", "Status value should be correct.")
	dsv, err = ReadStatus("dsr:d")
	assert.NotNil(err, "Error should be returned.")
	assert.ErrorMatch(err, "status error: .*", "Error inside retrieval has to be catched.")
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