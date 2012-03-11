// Tideland Common Go Library - Cells
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
	"code.google.com/p/tcgl/applog"
	"code.google.com/p/tcgl/identifier"
	"code.google.com/p/tcgl/monitoring"
	"fmt"
	"runtime"
	"sync"
	"time"
)

//--------------------
// CONST
//--------------------

const RELEASE = "Tideland Common Go Library - Cells - Release 2012-03-11"

//--------------------
// BASIC INTERFACES AND TYPES
//--------------------

// Context allows a number of coherent event processings to store
// values useful for an event emitter.
type Context struct {
	mutex           sync.RWMutex
	values          map[string]interface{}
	activityCounter int
	doneChan        chan bool
}

// newContext creates a new event processing context.
func newContext() *Context {
	return &Context{
		values:          make(map[string]interface{}),
		activityCounter: 0,
		doneChan:        make(chan bool, 1),
	}
}

// Set a value in the context.
func (c *Context) Set(key string, value interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.values[key] = value
}

// Get the value of 'key'.
func (c Context) Get(key string) (interface{}, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	v := c.values[key]
	if v == nil {
		return nil, fmt.Errorf("no context value for %q", key)
	}
	return v, nil
}

// Do iterates over all stored values and calls the passed function for each pair.
func (c Context) Do(f func(key string, value interface{})) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	for k, v := range c.values {
		f(k, v)
	}
}

// incrActivity indicates, that one more cell is working in the context.
func (c *Context) incrActivity() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.activityCounter++
}

// decrActivity indicates, that one more cell has stopped processing an event
// in the context. If the counter is down to zero the context signals that all
// work is done.
func (c *Context) decrActivity() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.activityCounter--
	if c.activityCounter <= 0 {
		select {
		case c.doneChan <- true:
		default:
		}
	}
}

// Wait lets a caller wait until the processing of the context is done.
func (c Context) Wait(d time.Duration) error {
	select {
	case <-c.doneChan:
		return nil
	case <-time.After(d):
		return fmt.Errorf("timeout during context wait")
	}
	return nil
}

// Event is anything that has a topic and a payload.
type Event interface {
	// Topic returns the topic of the simple event.
	Topic() string
	// Payload returns the payload of the simple event.
	Payload() interface{}
	// Context returns the context of a set of event processings.
	Context() *Context
	// SetContext set the context of a set of event processings.
	SetContext(c *Context)
}

// simpleEvent can be used if no own event implementation is
// wanted or needed.
type simpleEvent struct {
	topic   string
	payload interface{}
	context *Context
}

// NewSimpleEvent creates a simple event.
func NewSimpleEvent(t string, p interface{}) Event {
	return &simpleEvent{t, p, nil}
}

// Topic returns the topic of the simple event.
func (se simpleEvent) Topic() string {
	return se.topic
}

// Payload returns the payload of the simple event.
func (se simpleEvent) Payload() interface{} {
	return se.payload
}

// Context returns the context of a set of event processings.
func (se simpleEvent) Context() *Context {
	return se.context
}

// SetContext set the context of a set of event processings.
func (se *simpleEvent) SetContext(c *Context) {
	se.context = c
}

// EventChannel is a channel to pass events to other handlers.
type EventChannel chan Event

// Behavior is the interface that has to be implemented
// of those behaviors which can be plugged into an environment
// for the real event processing.
type Behavior interface {
	Init(env *Environment, id string) error
	ProcessEvent(e Event, emitChan EventChannel)
	Recover(r interface{}, e Event)
	Stop() error
}

//--------------------
// ENVIRONMENT
//--------------------

// Environment defines a common set of cells.
type Environment struct {
	mutex         sync.RWMutex
	id            string
	cells         map[string]*cell
	subscribers   assignments
	subscriptions assignments
}

// NewEnvironment creates a new environment.
func NewEnvironment(id string) *Environment {
	env := &Environment{
		id:            id,
		cells:         make(map[string]*cell),
		subscribers:   make(assignments),
		subscriptions: make(assignments),
	}
	runtime.SetFinalizer(env, (*Environment).Shutdown)
	return env
}

// AddCell adds a handler with a given id and an input queue length.
func (env *Environment) AddCell(id string, b Behavior, ql int) (Behavior, error) {
	env.mutex.Lock()
	defer env.mutex.Unlock()
	// Check registry before adding.
	if _, ok := env.cells[id]; ok {
		return nil, fmt.Errorf("cell behavior with id %q already added", id)
	}
	c, err := newCell(env, id, b, ql)
	if err != nil {
		return nil, err
	}
	env.cells[id] = c
	return b, nil
}

// RemoveCell removes the behavior with the given id. If it doesn't
// exist no error is returned.
func (env *Environment) RemoveCell(id string) error {
	env.mutex.Lock()
	defer env.mutex.Unlock()
	// Retrieve and remove it with all subscriptions.
	if c, ok := env.cells[id]; ok {
		env.subscribers.drop(id)
		env.subscriptions.dropAll(id, env.subscribers)
		delete(env.cells, id)
		return c.stop()
	}
	return nil
}

// Cell returns the behavior with the id or nil (take care).
func (env *Environment) Cell(id string) (Behavior, error) {
	env.mutex.RLock()
	defer env.mutex.RUnlock()
	// Retrieve the cell first
	if c, ok := env.cells[id]; ok {
		return c.behavior, nil
	}
	return nil, fmt.Errorf("cell behavior with id %q does not exist")
}

// Subscribe assigns cells as receivers of the emitted 
// events of the first cell.
func (env *Environment) Subscribe(eid string, sids ...string) error {
	env.mutex.Lock()
	defer env.mutex.Unlock()
	// Check both ids.
	if env.cells[eid] == nil {
		return fmt.Errorf("emitter cell %q does not exist", eid)
	}
	for _, sid := range sids {
		if env.cells[sid] == nil {
			return fmt.Errorf("subscriber cell %q does not exist", sid)
		}
	}
	// Store assignments in both directions.
	env.subscribers.add(eid, sids...)
	for _, sid := range sids {
		env.subscriptions.add(sid, eid)
	}
	return nil
}

// Unsubscribe removes the assignment of emitting und subscribed cells. 
func (env *Environment) Unsubscribe(eid string, sids ...string) error {
	env.mutex.Lock()
	defer env.mutex.Unlock()
	// Check both ids.
	if env.cells[eid] == nil {
		return fmt.Errorf("emitter cell %q does not exist", eid)
	}
	for _, sid := range sids {
		if env.cells[sid] == nil {
			return fmt.Errorf("subscriber cell %q does not exist", sid)
		}
	}
	// Remove assignments in both directions.
	env.subscribers.remove(eid, sids...)
	for _, sid := range sids {
		env.subscriptions.remove(sid, eid)
	}
	return nil
}

// Raise raises an event to the cell for a given id.
func (env *Environment) Raise(id string, e Event) (*Context, error) {
	env.mutex.RLock()
	defer env.mutex.RUnlock()
	// Retrieve the cell.
	if c, ok := env.cells[id]; ok {
		// Check the context.
		if e.Context() == nil {
			e.SetContext(newContext())
		}
		c.processEvent(e)
		return e.Context(), nil
	}
	return nil, fmt.Errorf("cell %q not found", id)
}

// RaiseSimpleEvent is a convenience method raising a simple event in one call.
func (env *Environment) RaiseSimpleEvent(id, t string, p interface{}) (*Context, error) {
	return env.Raise(id, NewSimpleEvent(t, p))
}

// raiseSubscribers raises an event to the subscribers of a cell. If the 
// event has a target only this cell will receive the event, otherwise all.
func (env *Environment) raiseSubscribers(id string, e Event) {
	env.mutex.RLock()
	defer env.mutex.RUnlock()
	// Raise event.
	for _, sid := range env.subscribers.all(id) {
		if c, ok := env.cells[sid]; ok {
			c.processEvent(e)
		}
	}
}

// Shutdown manages the proper finalization of an environment.
func (env *Environment) Shutdown() error {
	env.mutex.Lock()
	defer env.mutex.Unlock()
	// Stop all cells and delete registry.
	for id, c := range env.cells {
		err := c.stop()
		if err != nil {
			return fmt.Errorf("shutdown of cell %q failed", id)
		}
	}
	env.cells = nil
	env.subscribers = nil
	env.subscriptions = nil
	runtime.SetFinalizer(env, nil)
	return nil
}

//--------------------
// CELL 
//--------------------

// cell for event processing.
type cell struct {
	env         *Environment
	id          string
	behavior    Behavior
	eventChan   EventChannel
	measuringId string
}

// newCell create a new cell around a behavior.
func newCell(env *Environment, id string, b Behavior, ql int) (*cell, error) {
	c := &cell{
		env:         env,
		id:          id,
		behavior:    b,
		eventChan:   make(EventChannel, ql),
		measuringId: identifier.Identifier("cell", env.id, id),
	}
	go c.backend()
	err := b.Init(env, id)
	if err != nil {
		return nil, fmt.Errorf("cell behavior init error: %v", err)
	}
	return c, nil
}

// processEvent tells the cell to handle an event.
func (c *cell) processEvent(e Event) {
	e.Context().incrActivity()
	c.eventChan <- e
}

// stop terminates the cell.
func (c *cell) stop() error {
	close(c.eventChan)
	return c.behavior.Stop()
}

// backend function of the cell.
func (c *cell) backend() {
	// Main event loop.
	for e := range c.eventChan {
		c.process(e)
	}
}

// process encapsulates event processing including error 
// recovery and measuring.
func (c *cell) process(e Event) {
	// Error recovering.
	defer func() {
		if r := recover(); r != nil {
			applog.Errorf("cells", "cell '%s' has error '%v' after '%v'", c.id, r, e)
			c.behavior.Recover(r, e)
		}
	}()
	// Create a channel to let the behavior emit
	// events to the subscribed handlers. Those
	// will handle it in the background.
	emitChan := make(EventChannel)
	defer close(emitChan)
	go func() {
		for ee := range emitChan {
			ee.SetContext(e.Context())
			c.env.raiseSubscribers(c.id, ee)
		}
		e.Context().decrActivity()
	}()
	// Handle the event inside a measuring.
	measuring := monitoring.BeginMeasuring(c.measuringId)
	c.behavior.ProcessEvent(e, emitChan)
	measuring.EndMeasuring()
}

// EOF
