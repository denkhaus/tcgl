// Tideland Common Go Library - Cells
//
// Copyright (C) 2010-2011 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed 
// by the new BSD license.

package cells

//--------------------
// IMPORTS
//--------------------

import (
	"runtime"
	"tcgl.googlecode.com/hg/identifier"
	"tcgl.googlecode.com/hg/monitoring"
)

//--------------------
// CONST
//--------------------

const RELEASE = "Tideland Common Go Library - Cells - Release 2011-12-22"

//--------------------
// BASIC INTERFACES AND TYPES
//--------------------

// Handler is the interface for event handlers like input,
// output and cells.
type Handler interface {
	HandleEvent(e Event)
}

// Event is anything that has a topic and a payload.
type Event interface {
	Topic() string
	Payload() interface{}
}

// simpleEvent can be used if no own event implementation is
// wanted or needed.
type simpleEvent struct {
	topic   string
	payload interface{}
}

// NewSimpleEvent creates a simple event.
func NewSimpleEvent(t string, p interface{}) Event {
	return &simpleEvent{t, p}
}

// Topic returns the topic of the simple event.
func (se simpleEvent) Topic() string {
	return se.topic
}

// Payload returns the payload of the simple event.
func (se simpleEvent) Payload() interface{} {
	return se.payload
}

// EventChannel is a channel to pass events to other goroutines.
// So it will be used by the HandleEvent() methods of behaviors
// to emit events.
type EventChannel chan Event

// Behavior is the interface that has to be implemented
// of those behaviors which can be plugged into the cell.
// They do the real event processing.
type Behavior interface {
	ProcessEvent(e Event, emitChan EventChannel)
	Recover(err interface{}, e Event)
	Stop()
}

//--------------------
// CELL 
//--------------------

// Cell for event processing.
type Cell struct {
	behavior        Behavior
	filtered        bool
	subscriptions   subscriptionMap
	eventChan       EventChannel
	subscribeChan   chan Handler
	unsubscribeChan chan Handler
	stopChan        chan bool
	measuringId     string
}

// Create a new cell with a given configuration
func NewCell(b Behavior, ecLen int) *Cell {
	part := identifier.TypeAsIdentifierPart(b)
	c := &Cell{
		behavior:        b,
		subscriptions:   newSubscriptionMap(),
		eventChan:       make(EventChannel, ecLen),
		subscribeChan:   make(chan Handler),
		unsubscribeChan: make(chan Handler),
		stopChan:        make(chan bool),
		measuringId:     identifier.Identifier("cgl", "cell", part),
	}

	runtime.SetFinalizer(c, (*Cell).stop)

	go c.backend()

	return c
}

// HandleEvent tells the cell to handle an event.
func (c *Cell) HandleEvent(e Event) {
	c.eventChan <- e
}

// Subscribe a handler for emitted events.
func (c *Cell) Subscribe(h Handler) {
	c.subscribeChan <- h
}

// Unsubscribe a handler for emitted events.
func (c *Cell) Unsubscribe(h Handler) {
	c.unsubscribeChan <- h
}

// Stop the cell.
func (c *Cell) Stop() {
	c.stopChan <- true
}

// backend function of the cell.
func (c *Cell) backend() {
	// Add an error handler. Here simply do a restart.
	defer func() {
		if err := recover(); err != nil {
			go c.backend()
		}
	}()

	// Main event loop.
	for {
		select {
		case e := <-c.eventChan:
			// Handle an event.
			c.handle(e)
		case h := <-c.subscribeChan:
			// Subscribe a new handler.
			c.subscriptions.subscribe(h)
		case h := <-c.unsubscribeChan:
			// Unsubscribe a handler.
			c.subscriptions.unsubscribe(h)
		case <-c.stopChan:
			// Received stop signal.
			c.stop()

			return
		}
	}
}

// handle an event, including error recovery and measuring.
func (c *Cell) handle(e Event) {
	// Error recovering.
	defer func() {
		if err := recover(); err != nil {
			c.behavior.Recover(err, e)
		}
	}()

	// Create a channel to let the behavior emit
	// events to the subscribed handlers. Those
	// will handle it in the background.
	emitChan := make(EventChannel)

	defer close(emitChan)

	go func() {
		for ee := range emitChan {
			c.subscriptions.handleEvent(ee)
		}
	}()

	// Handle the event inside a measuring.
	measuring := monitoring.Monitor().BeginMeasuring(c.measuringId)

	c.behavior.ProcessEvent(e, emitChan)

	measuring.EndMeasuring()
}

// Terminate the cell.
func (c *Cell) stop() {
	c.behavior.Stop()
	c.subscriptions.unsubscribeAll()
	runtime.SetFinalizer(c, nil)
}

// EOF
