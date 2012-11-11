// Tideland Common Go Library - Event Bus - Agents
//
// Copyright (C) 2012 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed 
// by the new BSD license.

package ebus

//--------------------
// IMPORTS
//--------------------

import (
	"cgl.tideland.biz/applog"
)

//--------------------
// LOG AGENT
//--------------------

// LogAgent logs events.
type LogAgent struct {
	id string
}

// NewLogAgent creates a new log agent.
func NewLogAgent(id string) *LogAgent {
	return &LogAgent{id}
}

// Id returns the unique identifier of the agent.
func (l *LogAgent) Id() string {
	return l.id
}

// Process processes an event.
func (l *LogAgent) Process(event Event) error {
	applog.Infof("agent: %q event topic: %q", l.id, event.Topic())
	return nil
}

// Recover from an error during the processing of an event.
func (l *LogAgent) Recover(r interface{}, event Event) error {
	return nil
}

// Stop tells the agent to cleanup.
func (l *LogAgent) Stop() {
}

// Err returns the error the agent possibly stopped with.
func (l *LogAgent) Err() error {
	return nil
}

//--------------------
// SIMPLE FUNCTION AGENT
//--------------------

// SimpleFunc is a function type for simple event handling. 
type SimpleFunc func(event Event) error

// SimpleFuncAgent takes a function to process events. So the development
// of an own agent may be avoided.
type SimpleFuncAgent struct {
	id  string
	f   SimpleFunc
	err error
}

// NewSimpleFuncAgent creates a new simple function agent.
func NewSimpleFuncAgent(id string, f SimpleFunc) *SimpleFuncAgent {
	return &SimpleFuncAgent{id, f, nil}
}

// Id returns the unique identifier of the agent.
func (s *SimpleFuncAgent) Id() string {
	return s.id
}

// Process processes an event.
func (s *SimpleFuncAgent) Process(event Event) error {
	s.err = s.f(event)
	return s.err
}

// Recover from an error during the processing of an event.
func (s *SimpleFuncAgent) Recover(r interface{}, event Event) error {
	return s.err
}

// Stop tells the agent to cleanup.
func (s *SimpleFuncAgent) Stop() {}

// Err returns the error the agent possibly stopped with.
func (s *SimpleFuncAgent) Err() error {
	return s.err
}

//--------------------
// WRAPPER AGENT
//--------------------

// PreprocessFunc is the signature of the preprocessing function.
type PreprocessFunc func(event Event) (Event, error)

// WrapperAgent preprocesses an event before it is passed to the
// wrapped agent.
type WrapperAgent struct {
	a  Agent
	pp PreprocessFunc
}

// NewWrapperAgent creates a new wrapper agent.
func NewWrapperAgent(a Agent, pp PreprocessFunc) *WrapperAgent {
	return &WrapperAgent{a, pp}
}

// Id returns the unique identifier of the agent.
func (w *WrapperAgent) Id() string {
	return w.a.Id()
}

// Process processes an event.
func (w *WrapperAgent) Process(event Event) error {
	e, err := w.pp(event)
	if err != nil {
		return err
	}
	return w.a.Process(e)
}

// Recover from an error during the processing of an event.
func (w *WrapperAgent) Recover(r interface{}, event Event) error {
	return w.a.Recover(r, event)
}

// Stop tells the agent to cleanup.
func (w *WrapperAgent) Stop() {}

// Err returns the error the agent possibly stopped with.
func (w *WrapperAgent) Err() error {
	return w.a.Err()
}

//--------------------
// COUNTER AGENT
//--------------------

type CounterOp int

const (
	CounterOpIncr = iota
	CounterOpReset
	CounterOpEmit
)

// CounterFunc is a function type returning names of counters to increment,
// the emitting of the counters to topics or the reset of all counters.
type CounterFunc func(event Event) (CounterOp, []string, error)

// CounterAgent counts events based on the counter function.
type CounterAgent struct {
	id       string
	f        CounterFunc
	counters map[string]int64
	err      error
}

// NewCounterAgent creates a new counter agent.
func NewCounterAgent(id string, f CounterFunc) *CounterAgent {
	return &CounterAgent{id, f, make(map[string]int64), nil}
}

// Id returns the unique identifier of the agent.
func (c *CounterAgent) Id() string {
	return c.id
}

// Process processes an event.
func (c *CounterAgent) Process(event Event) error {
	op, ids, err := c.f(event)
	if err != nil {
		c.err = err
		return err
	}
	switch op {
	case CounterOpIncr:
		if ids != nil {
			for _, id := range ids {
				c.counters[id]++
			}
		}
	case CounterOpReset:
		c.counters = make(map[string]int64)
	case CounterOpEmit:
		if ids != nil {
			for _, id := range ids {
				Emit(c.counters, id)
			}
		}
	default:
		applog.Errorf("illegal op code %d of counter agent %q", op, c.id)
	}
	return nil
}

// Recover from an error during the processing of an event.
func (c *CounterAgent) Recover(r interface{}, event Event) error {
	return c.err
}

// Stop tells the agent to cleanup.
func (c *CounterAgent) Stop() {}

// Err returns the error the agent possibly stopped with.
func (c *CounterAgent) Err() error {
	return c.err
}

// EOF
