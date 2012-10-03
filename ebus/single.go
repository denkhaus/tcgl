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
	"cgl.tideland.biz/config"
	"sync"
)

//--------------------
// SINGLE NODE BACKEND
//--------------------

// singleBackend implements the event bus backend for a single node. 
type singleBackend struct{}

func newSingleBackend() backend {
	return &singleBackend{}
}

// Init initializes the single event bus with the given configuration. If this
// isn't done all further operation will fail.
func (b *singleBackend) Init(config *config.Configuration) error {
	return nil
}

// Register adds an agent factory with an id. The agent instance will
// only be created if the id is not yet used.
func (b *singleBackend) Register(id Id, factory AgentFactory) error {
	return nil
}

// Unregister stops and removes the agent with the given id.
func (b *singleBackend) Unregister(id Id) error {
	return nil
}

// NewContext creates an initial context.
func (b *singleBackend) NewContext() (Context, error) {
	return &singleContext{
		values: make(map[Id]Value),
	}, nil
}

// RaiseEvent raises an event with a context.
func (b *singleBackend) RaiseEvent(topic string, payload Value, ctx Context) error {
	return nil
}

//--------------------
// SINGLE NODE CONTEXT
//--------------------

// singleContext implements the context for a single node. 
// It just works in memory.
type singleContext struct {
	mutex  sync.RWMutex
	values map[Id]Value
}

// Set a value in the context.
func (ctx *singleContext) Set(key Id, value Value) error {
	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()
	ctx.values[key] = Value
	return nil
}

// Get a value out of the context.
func (ctx *singleContext) Get(key Id) (Value, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	v := ctx.values[key]
	if v == nil {
		return nil, &ContextValueNotFoundError{key}
	}
	return v, nil
}

// Delete a value from the context.
func (ctx *singleContext) Delete(key Id) error {
	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()
	delete(ctx.values, key)
	return nil
}

// EOF
