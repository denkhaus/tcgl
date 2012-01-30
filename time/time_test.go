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
	"testing"
	"time"
)

//--------------------
// TESTS
//--------------------

// Test nanoseconds calculation.
func TestNanoseconds(t *testing.T) {
	t.Logf("Microseconds: %v\n", NsMicroseconds(4711))
	t.Logf("Milliseconds: %v\n", NsMilliseconds(4711))
	t.Logf("Seconds     : %v\n", NsSeconds(4711))
	t.Logf("Minutes     : %v\n", NsMinutes(4711))
	t.Logf("Hours       : %v\n", NsHours(4711))
	t.Logf("Days        : %v\n", NsDays(4711))
	t.Logf("Weeks       : %v\n", NsWeeks(4711))
}

// Test time containments.
func TestTimeContainments(t *testing.T) {
	now := time.Now().UTC()
	years := []int{2008, 2009, 2010}
	months := []time.Month{3, 6, 9, 12}
	days := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	hours := []int{20, 21, 22, 23}
	minutes := []int{5, 10, 15, 20, 25, 30}
	seconds := []int{0, 15, 30, 45}
	weekdays := []time.Weekday{time.Saturday, time.Sunday}

	t.Logf("Time is %s\n", now.Format(time.RFC822))
	t.Logf("Year in list    : %t\n", YearInList(now, years))
	t.Logf("Year in range   : %t\n", YearInRange(now, 2000, 2005))
	t.Logf("Month in list   : %t\n", MonthInList(now, months))
	t.Logf("Month in range  : %t\n", MonthInRange(now, 1, 6))
	t.Logf("Day in list     : %t\n", DayInList(now, days))
	t.Logf("Day in range    : %t\n", DayInRange(now, 15, 25))
	t.Logf("Hour in list    : %t\n", HourInList(now, hours))
	t.Logf("Hour in range   : %t\n", HourInRange(now, 9, 17))
	t.Logf("Minute in list  : %t\n", MinuteInList(now, minutes))
	t.Logf("Minute in range : %t\n", MinuteInRange(now, 0, 29))
	t.Logf("Second in list  : %t\n", SecondInList(now, seconds))
	t.Logf("Second in range : %t\n", SecondInRange(now, 30, 59))
	t.Logf("Weekday in list : %t\n", WeekdayInList(now, weekdays))
	t.Logf("Weekday in range: %t\n", WeekdayInRange(now, time.Monday, time.Friday))
}

// Test crontab keeping the job.
func TestCrontabKeep(t *testing.T) {
	c := NewCrontab()
	cf := func(now time.Time) (bool, bool) { return now.Unix()%2 == 0, false }
	tf := func(id string) { t.Logf("Performed 'keep job' %s\n", id) }

	c.AddJob("keep", cf, tf)

	time.Sleep(10 * 1e9)

	c.Stop()
}

// Test crontab deleting the job.
func TestCrontabDelete(t *testing.T) {
	c := NewCrontab()
	cf := func(now time.Time) (bool, bool) { return now.Unix()%2 == 0, true }
	tf := func(id string) { t.Logf("Performed 'keep job' %s\n", id) }

	c.AddJob("keep", cf, tf)

	time.Sleep(10 * 1e9)

	c.Stop()
}

// EOF
