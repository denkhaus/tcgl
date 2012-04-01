// Tideland Common Go Library - Finite State Machine
//
// Copyright (C) 2009-2012 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed 
// by the new BSD license.

package state

//--------------------
// IMPORTS
//--------------------

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

//--------------------
// FINITE STATE MACHINE
//--------------------

// Transition type.
type Transition struct {
	Timestamp  time.Time
	Command    string
	State      string
	Payload    interface{}
	ResultChan chan interface{}
}

// HandlerMap maps states to handler methods.
type HandlerMap struct {
	handler reflect.Value
	methods map[string]reflect.Value
}

// NewHandlerMap creates a new handler map with initial state
// to method assignments.
func NewHandlerMap(h Handler) *HandlerMap {
	hm := &HandlerMap{
		handler: reflect.ValueOf(h),
		methods: make(map[string]reflect.Value),
	}
	return hm
}

// Assign adds an assignment of a state to a handler method.
func (hm *HandlerMap) Assign(state, method string) error {
	mv := hm.handler.MethodByName(method)
	mvt := mv.Type()
	// Check the method.
	if mvt.NumIn() != 2 && mvt.NumOut() != 1 {
		return fmt.Errorf("%q is no valid handler method", method)
	}
	// Assign the method.
	hm.methods[strings.ToLower(state)] = mv
	return nil
}

// call does the call of a handler method for a state.
func (hm *HandlerMap) call(state string, t *Transition) (next string, err error) {
	defer func() {
		if e := recover(); e != nil {
			next = ""
			err = fmt.Errorf("state runtime error: %v", e)
		}
	}()
	if method, ok := hm.methods[state]; ok {
		args := []reflect.Value{reflect.ValueOf(t)}
		results := method.Call(args)
		return strings.ToLower(results[0].Interface().(string)), nil
	}
	// Illegal state.
	return "", fmt.Errorf("tried to handle illegal state %q", state)
}

// Handler interface.
type Handler interface {
	Init() (*HandlerMap, string)
	Error(*Transition, error) string
	Terminate()
}

// State machine type.
type FSM struct {
	handler        Handler
	handlerMap     *HandlerMap
	state          string
	transitionChan chan *Transition
	tickChan       <-chan time.Time
	stateChan      chan chan string
}

// Create a new finite state machine.
func New(h Handler, tick time.Duration) *FSM {
	hm, s := h.Init()
	fsm := &FSM{
		handler:        h,
		handlerMap:     hm,
		state:          s,
		transitionChan: make(chan *Transition),
		tickChan:       time.Tick(tick),
		stateChan:      make(chan chan string),
	}
	// Start working.
	go fsm.backend()
	return fsm
}

// HandleWithResult lets the FSM handle a command and payload and 
// returns a channel for a possible result.
func (fsm *FSM) HandleWithResult(cmd string, payload interface{}) chan interface{} {
	t := &Transition{time.Now(), cmd, "", payload, make(chan interface{})}
	fsm.transitionChan <- t
	return t.ResultChan
}

// Handle lets the FSM handle a command and payload.
func (fsm *FSM) Handle(cmd string, payload interface{}) {
	t := &Transition{time.Now(), cmd, "", payload, nil}
	fsm.transitionChan <- t
}

// HandeAfter lets the FSM handle a command and payload after a given duration.
func (fsm *FSM) HandleAfter(cmd string, payload interface{}, after time.Duration) {
        haf := func() {
                time.Sleep(after)
                fsm.Handle(cmd, payload)
        }
        go haf()
}

// State returns the current state.
func (fsm *FSM) State() string {
	stateChan := make(chan string)
	fsm.stateChan <- stateChan
	return <-stateChan
}

// backend is the state machines backend.
func (fsm *FSM) backend() {
	// Handle one transition.
	handle := func(t *Transition) {
		var err error
		fsm.state, err = fsm.handlerMap.call(fsm.state, t)
		if err != nil {
			fsm.state = fsm.handler.Error(t, err)
		}
		if fsm.state == "terminate" {
			fsm.handler.Terminate()
			fsm.state = "terminated"
		}
	}
	// Message loop.
	for {
		select {
		case t := <-fsm.transitionChan:
			// Regular transition.
			handle(t)
		case <-fsm.tickChan:
			// Received a tick.
			handle(&Transition{time.Now(), "tick", fsm.state, nil, nil})
		case stateChan := <-fsm.stateChan:
			// Send the current state.
			stateChan <- fsm.state
		}
	}
}

// EOF
