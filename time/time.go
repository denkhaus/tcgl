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
	"time"
)

//--------------------
// CONST
//--------------------

const RELEASE = "Tideland Common Go Library - Time - Release 2012-01-24"

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
func YearInList(time time.Time, years []int) bool {
	for _, year := range years {
		if time.Year() == year {
			return true
		}
	}

	return false
}

// YearInRange tests if a year of a time is in a given range.
func YearInRange(time time.Time, minYear, maxYear int) bool {
	return (minYear <= time.Year()) && (time.Year() <= maxYear)
}

// MonthInList tests if the month of a time is in a given list.
func MonthInList(time time.Time, months []time.Month) bool {
	for _, month := range months {
		if time.Month() == month {
			return true
		}
	}
	return false
}

// MonthInRange tests if a month of a time is in a given range.
func MonthInRange(time time.Time, minMonth, maxMonth time.Month) bool {
	return (minMonth <= time.Month()) && (time.Month() <= maxMonth)
}

// DayInList tests if the day of a time is in a given list.
func DayInList(time time.Time, days []int) bool {
	for _, day := range days {
		if time.Day() == day {
			return true
		}
	}
	return false
}

// DayInRange tests if a day of a time is in a given range.
func DayInRange(time time.Time, minDay, maxDay int) bool {
	return (minDay <= time.Day()) && (time.Day() <= maxDay)
}

// HourInList tests if the hour of a time is in a given list.
func HourInList(time time.Time, hours []int) bool {
	for _, hour := range hours {
		if time.Hour() == hour {
			return true
		}
	}
	return false
}

// HourInRange tests if a hour of a time is in a given range.
func HourInRange(time time.Time, minHour, maxHour int) bool {
	return (minHour <= time.Hour()) && (time.Hour() <= maxHour)
}

// MinuteInList tests if the minute of a time is in a given list.
func MinuteInList(time time.Time, minutes []int) bool {
	for _, minute := range minutes {
		if time.Minute() == minute {
			return true
		}
	}
	return false
}

// MinuteInRange tests if a minute of a time is in a given range.
func MinuteInRange(time time.Time, minMinute, maxMinute int) bool {
	return (minMinute <= time.Minute()) && (time.Minute() <= maxMinute)
}

// SecondInList tests if the second of a time is in a given list.
func SecondInList(time time.Time, seconds []int) bool {
	for _, second := range seconds {
		if time.Second() == second {
			return true
		}
	}
	return false
}

// SecondInRange tests if a second of a time is in a given range.
func SecondInRange(time time.Time, minSecond, maxSecond int) bool {
	return (minSecond <= time.Second()) && (time.Second() <= maxSecond)
}

// WeekdayInList tests if the weekday of a time is in a given list.
func WeekdayInList(time time.Time, weekdays []time.Weekday) bool {
	for _, weekday := range weekdays {
		if time.Weekday() == weekday {
			return true
		}
	}
	return false
}

// WeekdayInRange tests if a weekday of a time is in a given range.
func WeekdayInRange(time time.Time, minWeekday, maxWeekday time.Weekday) bool {
	return (minWeekday <= time.Weekday()) && (time.Weekday() <= maxWeekday)
}

//--------------------
// CRONJOB
//--------------------

// cronCommand operates on a crontab.
type cronCommand func() bool

// CheckFunc is the function type for checking if a job
// shall be performed now. It also returns if a job shall
// be deleted after execution.
type CheckFunc func(time.Time) (bool, bool)

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
func (j *job) checkAndPerform(time time.Time) bool {
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
		delete(c.jobs, id)
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
	now := time.Now().UTC()
	deletes := make(map[string]*job)
	// Check and perform jobs.
	for id, job := range c.jobs {
		delete := job.checkAndPerform(now)

		if delete {
			deletes[id] = job
		}
	}
	// Delete those marked for deletion.
	for id, _ := range deletes {
		delete(c.jobs, id)
	}
}

// EOF
