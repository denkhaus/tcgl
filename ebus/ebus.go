// Tideland Common Go Library - Event Bus - Main API
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
	"cgl.tideland.biz/config"
	"time"
)

//--------------------
// INTERFACES
//--------------------

// Value represents the values that can be stored in the context
// of used as event payload.
type Value interface{}

// Context is passed with the events during processing
// to manage shared values.
type Context interface {
	Set(key Id, value Value) error
	Get(key Id) (Value, error)
	Delete(key Id) error
}

// Event represents anything that happens in the system and has
// to be processed.
type Event interface {
	// Time returns the time the event is created.
	Time() time.Time
	// Topic returns the topic of the event.
	Topic() string
	// Payload returns the payload of the event.
	Payload() Value
}

// Agent is the interface that has to be implemented
// of those agent behaviors which can be deployed to the
// event bus.
type Agent interface {
	// Init initializes the deployed agent. It returns a slice
	// of topics it subscribes to.
	Init(id Id) ([]string, error)
	// Process processes an event.
	Process(evt Event, ctx Context) error
	// Recover from an error during the processing of an event.
	Recover(r interface{}, evt Event) error
	// Stop tells the agent to cleanup.
	Stop() error
}

// AgentFactory is a function that creates an agent instance.
type AgentFactory func() Agent

//--------------------
// BACKEND
//--------------------

// backend defines the methods a backend has to implement to be
// used as event bus.
type backend interface {
	Init(config *config.Configuration) error
	Register(id Id, factory AgentFactory) error
	Unregister(id Id) error
	NewContext() (Context, error)
	RaiseEvent(topic string, payload Value, ctx Context) error
}

// eventBus is the backend used by the API functions.
var eventBus *backend

//--------------------
// FUNCTIONS
//--------------------

// Init initializes the event bus with the given configuration. If this
// isn't done all further operation will fail.
func Init(config *config.Configuration) error {
	eventBus = newSingleBackend()

	return eventBus.Init(config)
}

// Register adds an agent factory with an id. The agent instance will
// only be created if the id is not yet used.
func Register(id Id, factory AgentFactory) error {
	if eventBus == nil {
		panic("event bus is not initialized")
	}
	return eventBus.Register(id, factory)
}

// Unregister stops and removes the agent with the given id.
func Unregister(id Id) error {
	if eventBus == nil {
		panic("event bus is not initialized")
	}
	return eventBus.Unregister(id)
}

// NewContext creates an initial context.
func NewContext() (Context, error) {
	if eventBus == nil {
		panic("event bus is not initialized")
	}
	return eventBus.NewContext(id)
}

// RaiseEvent raises an event with a context.
func RaiseEvent(topic string, payload Value, ctx Context) error {
	if eventBus == nil {
		panic("event bus is not initialized")
	}
	return eventBus.RaiseEvent(topic, payload, ctx)
}

// EOF
