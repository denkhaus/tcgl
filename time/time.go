// Tideland Common Go Library - Time
//
// Copyright (C) 2009-2011 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed 
// by the new BSD license.

package time

//--------------------
// IMPORTS
//--------------------

import (
	"time"
)

//--------------------
// CONST
//--------------------

const RELEASE = "Tideland Common Go Library - Time - Release 2011-12-18"

//--------------------
// DATE AND TIME
//--------------------

// NsMicroseconds calcs nanoseconds from microseconds.
func NsMicroseconds(count int64) int64 { return count * 1e3 }

// NsMilliseconds calcs nanoseconds from milliseconds.
func NsMilliseconds(count int64) int64 { return NsMicroseconds(count * 1e3) }

// NsSeconds calcs nanoseconds from seconds.
func NsSeconds(count int64) int64 { return NsMilliseconds(count * 1e3) }

// NsMinutes calcs nanoseconds from minutes.
func NsMinutes(count int64) int64 { return NsSeconds(count * 60) }

// NsHours calcs nanoseconds from hours.
func NsHours(count int64) int64 { return NsMinutes(count * 60) }

// NsDays calcs nanoseconds from days.
func NsDays(count int64) int64 { return NsHours(count * 24) }

// NsWeeks calcs nanoseconds from weeks.
func NsWeeks(count int64) int64 { return NsDays(count * 7) }

// Test if the year of a time is in a given list.
func YearInList(time *time.Time, years []int64) bool {
	for _, year := range years {
		if time.Year == year {
			return true
		}
	}

	return false
}

// YearInRange tests if a year of a time is in a given range.
func YearInRange(time *time.Time, minYear, maxYear int64) bool {
	return (minYear <= time.Year) && (time.Year <= maxYear)
}

// MonthInList tests if the month of a time is in a given list.
func MonthInList(time *time.Time, months []int) bool {
	return fieldInList(time.Month, months)
}

// MonthInRange tests if a month of a time is in a given range.
func MonthInRange(time *time.Time, minMonth, maxMonth int) bool {
	return fieldInRange(time.Month, minMonth, maxMonth)
}

// DayInList tests if the day of a time is in a given list.
func DayInList(time *time.Time, days []int) bool {
	return fieldInList(time.Day, days)
}

// DayInRange tests if a day of a time is in a given range.
func DayInRange(time *time.Time, minDay, maxDay int) bool {
	return fieldInRange(time.Day, minDay, maxDay)
}

// HourInList tests if the hour of a time is in a given list.
func HourInList(time *time.Time, hours []int) bool {
	return fieldInList(time.Hour, hours)
}

// HourInRange tests if a hour of a time is in a given range.
func HourInRange(time *time.Time, minHour, maxHour int) bool {
	return fieldInRange(time.Hour, minHour, maxHour)
}

// MinuteInList tests if the minute of a time is in a given list.
func MinuteInList(time *time.Time, minutes []int) bool {
	return fieldInList(time.Minute, minutes)
}

// MinuteInRange tests if a minute of a time is in a given range.
func MinuteInRange(time *time.Time, minMinute, maxMinute int) bool {
	return fieldInRange(time.Minute, minMinute, maxMinute)
}

// SecondInList tests if the second of a time is in a given list.
func SecondInList(time *time.Time, seconds []int) bool {
	return fieldInList(time.Second, seconds)
}

// SecondInRange tests if a second of a time is in a given range.
func SecondInRange(time *time.Time, minSecond, maxSecond int) bool {
	return fieldInRange(time.Second, minSecond, maxSecond)
}

// WeekdayInList tests if the weekday of a time is in a given list.
func WeekdayInList(time *time.Time, weekdays []int) bool {
	return fieldInList(time.Weekday, weekdays)
}

// WeekdayInRange tests if a weekday of a time is in a given range.
func WeekdayInRange(time *time.Time, minWeekday, maxWeekday int) bool {
	return fieldInRange(time.Weekday, minWeekday, maxWeekday)
}

//--------------------
// CRONJOB
//--------------------

// cronCommand operates on a crontab.
type cronCommand func() bool

// CheckFunc is the function type for checking if a job
// shall be performed now. It also returns if a job shall
// be deleted after execution.
type CheckFunc func(*time.Time) (bool, bool)

// TaskFunc is the function type that will be performed 
// if a jobs check func returns true.
type TaskFunc func(string)

// job represents one cronological job.
type job struct {
	id    string
	check CheckFunc
	task  TaskFunc
}

// checkAndPerform checks, if a job shall be performed. If true the
// task function will be called.
func (j *job) checkAndPerform(time *time.Time) bool {
	perform, delete := j.check(time)

	if perform {
		go j.task(j.id)
	}

	return perform && delete
}

// Crontab is one cron server. A system can run multiple in
// parallel.
type Crontab struct {
	jobs        map[string]*job
	commandChan chan cronCommand
	ticker      *time.Ticker
}

// NewCrontab creates a cron server.
func NewCrontab() *Crontab {
	c := &Crontab{
		jobs:        make(map[string]*job),
		commandChan: make(chan cronCommand),
		ticker:      time.NewTicker(1e9),
	}

	go c.backend()

	return c
}

// Stop terminates the server.
func (c *Crontab) Stop() {
	c.commandChan <- func() bool {
		return false
	}
}

// AddJob adds a new job to the server.
func (c *Crontab) AddJob(id string, cf CheckFunc, tf TaskFunc) {
	c.commandChan <- func() bool {
		c.jobs[id] = &job{id, cf, tf}

		return false
	}
}

// DeleteJob removes a job from the server.
func (c *Crontab) DeleteJob(id string) {
	c.commandChan <- func() bool {
		job, _ := c.jobs[id]
		c.jobs[id] = job, false

		return false
	}
}

// Crontab backend.
func (c *Crontab) backend() {
	for {
		select {
		case cmd := <-c.commandChan:
			// A server command.
			if cmd() {
				c.ticker.Stop()

				return
			}
		case <-c.ticker.C:
			// One tick every second.
			c.tick()
		}
	}
}

// Handle one server tick.
func (c *Crontab) tick() {
	now := time.UTC()
	deletes := make(map[string]*job)

	// Check and perform jobs.
	for id, job := range c.jobs {
		delete := job.checkAndPerform(now)

		if delete {
			deletes[id] = job
		}
	}

	// Delete those marked for deletion.
	for id, job := range deletes {
		c.jobs[id] = job, false
	}
}

//--------------------
// HELPERS
//--------------------

// fieldInList tests if a field is contained in a list.
func fieldInList(field int, list []int) bool {
	for _, item := range list {
		if field == item {
			return true
		}
	}

	return false
}

// fieldInRange tests if a field is in a given int range.
func fieldInRange(field int, min, max int) bool {
	return (min <= field) && (field <= max)
}

// EOF
