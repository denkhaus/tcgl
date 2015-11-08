// Tideland Common Go Library - Event Bus - Single Node Backend
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
	"github.com/denkhaus/tcgl/config"
)

//--------------------
// SINGLE NODE BACKEND
//--------------------

// singleNodeBackend implements the event bus backend for a single node. 
type singleNodeBackend struct {
	router *nodeRouter
}

func newSingleNodeBackend() backend {
	return &singleNodeBackend{newNodeRouter()}
}

// Init initializes the single event bus with the given configuration. If this
// isn't done all further operation will fail.
func (b *singleNodeBackend) Init(config *config.Configuration) error {
	return nil
}

// Stop shuts the event bus down.
func (b *singleNodeBackend) Stop() error {
	stopTickers()
	b.router.stop()
	return nil
}

// Register adds an agent.
func (b *singleNodeBackend) Register(agent Agent) (Agent, error) {
	err := b.router.register(agent)
	return agent, err
}

// Deregister stops and removes the agent.
func (b *singleNodeBackend) Deregister(agent Agent) error {
	return b.router.deregister(agent)
}

// Lookup retrieves a registered agent by id.
func (b *singleNodeBackend) Lookup(id string) (Agent, error) {
	return b.router.lookup(id)
}

// Subscribe subscribes the agent to the topic.
func (b *singleNodeBackend) Subscribe(agent Agent, topic string) error {
	return b.router.subscribe(agent, topic)
}

// Unsubscribe removes the subscription of the agent from the topic. 
func (b *singleNodeBackend) Unsubscribe(agent Agent, topic string) error {
	return b.router.unsubscribe(agent, topic)
}

// Emit emits new event to the event bus.
func (b *singleNodeBackend) Emit(event Event) error {
	b.router.push(event)
	return nil
}

// EOF
