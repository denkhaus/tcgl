// Tideland Common Go Library - Time
//
// Copyright (C) 2009-2012 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed 
// by the new BSD license.

package time

//--------------------
// IMPORTS
//--------------------

import (
	"cgl.tideland.biz/asserts"
	"testing"
	"time"
)

//--------------------
// TESTS
//--------------------

// Test time containments.
func TestTimeContainments(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	// Create some test data.
	ts := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	years := []int{2008, 2009, 2010}
	months := []time.Month{10, 11, 12}
	days := []int{10, 11, 12, 13, 14}
	hours := []int{20, 21, 22, 23}
	minutes := []int{0, 5, 10, 15, 20, 25}
	seconds := []int{0, 15, 30, 45}
	weekdays := []time.Weekday{time.Monday, time.Tuesday, time.Wednesday}

	assert.True(YearInList(ts, years), "Go time in year list.")
	assert.True(YearInRange(ts, 2005, 2015), "Go time in year range.")
	assert.True(MonthInList(ts, months), "Go time in month list.")
	assert.True(MonthInRange(ts, 7, 12), "Go time in month range.")
	assert.True(DayInList(ts, days), "Go time in day list.")
	assert.True(DayInRange(ts, 5, 15), "Go time in day range .")
	assert.True(HourInList(ts, hours), "Go time in hour list.")
	assert.True(HourInRange(ts, 20, 31), "Go time in hour range .")
	assert.True(MinuteInList(ts, minutes), "Go time in minute list.")
	assert.True(MinuteInRange(ts, 0, 5), "Go time in minute range .")
	assert.True(SecondInList(ts, seconds), "Go time in second list.")
	assert.True(SecondInRange(ts, 0, 5), "Go time in second range .")
	assert.True(WeekdayInList(ts, weekdays), "Go time in weekday list.")
	assert.True(WeekdayInRange(ts, time.Monday, time.Friday), "Go time in weekday range .")
}

// Test crontab keeping the job.
func TestCrontabKeep(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	// Create test crontab with job.
	counter := 0
	c := NewCrontab()
	cf := func(now time.Time) (bool, bool) { return now.Unix()%2 == 0, false }
	tf := func(id string) { counter++ }

	c.AddJob("keep", cf, tf)
	time.Sleep(5 * time.Second)
	c.Stop()

	assert.Equal(counter, 2, "Counter should be increased two times.")
}

// Test crontab deleting the job.
func TestCrontabDelete(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	// Create test crontab with job.
	counter := 0
	c := NewCrontab()
	cf := func(now time.Time) (bool, bool) { return now.Unix()%2 == 0, true }
	tf := func(id string) { counter++ }

	c.AddJob("keep", cf, tf)
	time.Sleep(5 * 1e9)
	c.Stop()

	assert.Equal(counter, 1, "Counter should be increased only once.")
}

// EOF
