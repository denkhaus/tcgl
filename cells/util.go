// Tideland Common Go Library - Cells - Utilities
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
	"code.google.com/p/tcgl/identifier"
	"fmt"
	"sync"
	"time"
)

//--------------------
// IDENTIFIER
//--------------------

// Id is used to identify context values and cells.
type Id string

// NewId generates an identifier based on the given parts.
func NewId(parts ...interface{}) Id {
	return Id(identifier.Identifier(parts...))
}

//--------------------
// CELL MAP
//--------------------

// cellMap is a map from id to cells for subscriptions and
// subscribers.
type cellMap map[Id]*cell

// subset returns a cell map containing the cell with the given ids.
func (cm cellMap) subset(ids ...Id) (cellMap, error) {
	scm := make(cellMap)
	for _, id := range ids {
		c, ok := cm[id]
		if !ok {
			return nil, CellDoesNotExistError{id}
		}
		scm[id] = c
	}
	return scm, nil
}

//--------------------
// CELL MESSAGE QUEUE
//--------------------

// cellMessage is a message that's handled by the cells 
// backend loops.
type cellMessage struct {
	event Event
	cells cellMap
	add   bool
}

// String returns a readable representation of the message.
func (m cellMessage) String() string {
	ids := []string{}
	for id := range m.cells {
		ids = append(ids, string(id))
	}
	return fmt.Sprintf("<%v %v %v>", EventString(m.event), ids, m.add)
}

// cellMessageQueue provides an unlimitted message queue for the cells.
type cellMessageQueue struct {
	cond   *sync.Cond
	buffer []*cellMessage
}

// newCellMessageQueue creates an empty message queue.
func newCellMessageQueue() *cellMessageQueue {
	var locker sync.Mutex
	return &cellMessageQueue{sync.NewCond(&locker), make([]*cellMessage, 0)}
}

// push appends a new message to the queue.
func (q *cellMessageQueue) push(event Event, cells cellMap, add bool) error {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	if q.buffer == nil {
		return QueueClosedError{}
	}
	q.buffer = append(q.buffer, &cellMessage{event, cells, add})
	q.cond.Signal()
	return nil
}

// pull retrieves a message out of the queue. If it's empty pull
// is waiting.
func (q *cellMessageQueue) pull() (msg *cellMessage) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	for {
		if len(q.buffer) == 0 {
			q.cond.Wait()
		} else {
			msg = q.buffer[0]
			q.buffer = q.buffer[1:]
			break
		}
	}
	return
}

// close tells the queue to stop working.
func (q *cellMessageQueue) close() {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	q.buffer = nil
}

//--------------------
// CONTEXT
//--------------------

// Context allows a number of coherent event processings to store
// and retrieves values useful for an event emitter and wait for all
// cells to end processing their events in this context.
type Context struct {
	mutex           sync.RWMutex
	values          map[Id]interface{}
	activityCounter int
	doneChan        chan bool
}

// newContext creates a new event processing context.
func newContext() *Context {
	return &Context{
		values:          make(map[Id]interface{}),
		activityCounter: 1,
		doneChan:        make(chan bool, 1),
	}
}

// Set a value in the context.
func (c *Context) Set(key Id, value interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.values[key] = value
}

// Get the value of 'key'.
func (c *Context) Get(key Id) (interface{}, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	v := c.values[key]
	if v == nil {
		return nil, fmt.Errorf("no context value for %q", key)
	}
	return v, nil
}

// Do iterates over all stored values and calls the passed function for each pair.
func (c *Context) Do(f func(key Id, value interface{})) {
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
func (c *Context) Wait(timeout time.Duration) error {
	select {
	case <-c.doneChan:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("timeout during context wait")
	}
	return nil
}

//--------------------
// TICKER
//--------------------

// ticker provides periodic events raised at a defined id.
type ticker struct {
	env      *Environment
	id       Id
	emitId   Id
	period   time.Duration
	stopChan chan bool
}

// startTicker starts a new ticker in the background.
func startTicker(env *Environment, id, emitId Id, period time.Duration) *ticker {
	t := &ticker{env, id, emitId, period, make(chan bool)}
	go t.backend()
	return t
}

// stop lets the backend goroutine stop working.
func (t *ticker) stop() {
	t.stopChan <- true
}

// backend is the goroutine running the ticker.
func (t *ticker) backend() {
	for {
		select {
		case <-time.After(t.period):
			t.env.Emit(t.emitId, NewTickerEvent(t.id))
		case <-t.stopChan:
			return
		}
	}
}

// TickerEvent signals a tick to ticker subscribers.
type TickerEvent struct {
	id      Id
	time    time.Time
	context *Context
}

// NewTickerEvent creates a new ticker event instance with a 
// given id and the current time.
func NewTickerEvent(id Id) *TickerEvent {
	return &TickerEvent{id, time.Now(), nil}
}

// Topic returns the topic of the event, here "ticker([id])".
func (te TickerEvent) Topic() string {
	return fmt.Sprintf("ticker(%s)", te.id)
}

// Payload returns the payload of the event, here the time in
// nanoseconds.
func (te TickerEvent) Payload() interface{} {
	return te.time
}

// Context returns the context of a set of event processings.
func (te TickerEvent) Context() *Context {
	return te.context
}

// SetContext set the context of a set of event processings.
func (te *TickerEvent) SetContext(c *Context) {
	te.context = c
}

//--------------------
// HELPER FUNCTIONS
//--------------------

// EventString returns an event as a readable string.
func EventString(e Event) string {
	if e == nil {
		return "none"
	}
	return fmt.Sprintf("<event topic: %q payload: %+v>", e.Topic(), e.Payload())
}

//--------------------
// ERRORS
//--------------------

// CellInitError will be returned if a cell behaviors init method
// returns an error.
type CellInitError struct {
	Id  Id
	Err error
}

// Error returns the error as string.
func (e CellInitError) Error() string {
	return fmt.Sprintf("cell %q can't initialize: %v", e.Id, e.Err)
}

// IsCellInitError checks if an error is a cell init error.
func IsCellInitError(err error) bool {
	_, ok := err.(CellInitError)
	return ok
}

// CellAlreadyExistsError will be returned if a cell already exists.
type CellAlreadyExistsError struct {
	Id Id
}

// Error returns the error as string.
func (e CellAlreadyExistsError) Error() string {
	return fmt.Sprintf("cell %q already exists", e.Id)
}

// IsCellAlreadyExistsError checks if an error is a cell already exists error.
func IsCellAlreadyExistsError(err error) bool {
	_, ok := err.(CellAlreadyExistsError)
	return ok
}

// CellDoesNotExistError will be returned if a cell does not exist.
type CellDoesNotExistError struct {
	Id Id
}

// Error returns the error as string.
func (e CellDoesNotExistError) Error() string {
	return fmt.Sprintf("cell %q does not exist", e.Id)
}

// IsCellDoesNotExistError checks if an error is a cell does not exist error.
func IsCellDoesNotExistError(err error) bool {
	_, ok := err.(CellDoesNotExistError)
	return ok
}

// CellStoppedError will be returned if a subscribed cell has been
// stopped and so removed from a cell subscription map.
type CellStoppedError struct {
	Id Id
}

// Error returns the error as string.
func (e CellStoppedError) Error() string {
	return fmt.Sprintf("cell %q has been stopped", e.Id)
}

// IsCellStoppedError checks if an error is a cell stopped error.
func IsCellStoppedError(err error) bool {
	_, ok := err.(CellStoppedError)
	return ok
}

// QueueClosedError will be returned if a cell message queue is
// closed and a message shall be pushed or pulled.
type QueueClosedError struct{}

// Error returns the error as string.
func (e QueueClosedError) Error() string {
	return fmt.Sprintf("cell message queue has been closed")
}

// IsQueueClosedError checks if an error is a queue closed error.
func IsQueueClosedError(err error) bool {
	_, ok := err.(QueueClosedError)
	return ok
}

// EOF
