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
	"fmt"
)

//--------------------
// INTERFACES
//--------------------

// Event represents anything that happens in the system and has
// to be processed.
type Event interface {
	// Payload returns the payload of the event into
	// the value. It is passed as a serialized copy to
	// avoid concurrent changes and to provide the same
	// content when it is distributed via networks.
	Payload(value interface{}) error
	// Topic returns the topic of the event.
	Topic() string
}

// Agent is the interface that has to be implemented
// of those agent behaviors which can be deployed to the
// event bus.
type Agent interface {
	// Id returns the unique identifier of the agent.
	Id() string
	// Process processes an event.
	Process(event Event) error
	// Recover from an error during the processing of an event.
	Recover(r interface{}, event Event) error
	// Stop tells the agent to cleanup.
	Stop()
	// Err returns the error the agent possibly stopped with.
	Err() error
}

//--------------------
// BACKEND
//--------------------

// backend defines the methods a backend has to implement to be
// used as event bus.
type backend interface {
	Init(config *config.Configuration) error
	Stop() error
	Register(agent Agent) (Agent, error)
	Deregister(agent Agent) error
	Lookup(id string) (Agent, error)
	Subscribe(agent Agent, topic string) error
	Unsubscribe(agent Agent, topic string) error
	Emit(event Event) error
}

// eventBus is the backend used by the API functions.
var eventBus backend

//--------------------
// FUNCTIONS
//--------------------

// Init initializes the event bus with the given configuration. If this
// isn't done all further operation will fail.
func Init(config *config.Configuration) error {
	backend, err := config.GetDefault("backend", "single")
	if err != nil {
		return err
	}
	switch backend {
	case "single":
		eventBus = newSingleNodeBackend()
	default:
		panic(fmt.Sprintf("invalid backend %q", backend))
	}
	return eventBus.Init(config)
}

// Stop shuts the event bus down.
func Stop() error {
	if eventBus == nil {
		panic("event bus is not initialized")
	}
	return eventBus.Stop()
}

// Register adds an agent.
func Register(agent Agent) (Agent, error) {
	if eventBus == nil {
		panic("event bus is not initialized")
	}
	return eventBus.Register(agent)
}

// Deregister stops and removes the agent.
func Deregister(agent Agent) error {
	if eventBus == nil {
		panic("event bus is not initialized")
	}
	return eventBus.Deregister(agent)
}

// Lookup retrieves a registered agent by id.
func Lookup(id string) (Agent, error) {
	if eventBus == nil {
		panic("event bus is not initialized")
	}
	return eventBus.Lookup(id)
}

// Subscribe subscribes the agent to the topic created out of 
// the stem and the parts.
func Subscribe(agent Agent, stem string, parts ...interface{}) error {
	if eventBus == nil {
		panic("event bus is not initialized")
	}
	return eventBus.Subscribe(agent, Id(stem, parts...))
}

// Unsubscribe removes the subscription of the agent from the topic 
// created out of the stem and the parts.
func Unsubscribe(agent Agent, stem string, parts ...interface{}) error {
	if eventBus == nil {
		panic("event bus is not initialized")
	}
	return eventBus.Unsubscribe(agent, Id(stem, parts...))
}

// Emit emits new event with the given payload and the topic
// created out of the stem and the parts to the event bus.
func Emit(payload interface{}, stem string, parts ...interface{}) error {
	if eventBus == nil {
		panic("event bus is not initialized")
	}
	event, err := newSimpleEvent(payload, Id(stem, parts...))
	if err != nil {
		return err
	}
	return eventBus.Emit(event)
}

// EOF
