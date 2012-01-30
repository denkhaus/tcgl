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
	"reflect"
	"strings"
	"time"
)

//--------------------
// CONST
//--------------------

const RELEASE = "Tideland Common Go Library - Finite State Machine - Release 2012-01-23"

//--------------------
// HELPER TYPES
//--------------------

// Condition type.
type Condition struct {
	Now     time.Time
	Payload interface{}
}

// Transition type.
type transition struct {
	payload    interface{}
	resultChan chan interface{}
}

// Timeout type.
type Timeout time.Time

//--------------------
// FINITE STATE MACHINE
//--------------------

// Handler interface.
type Handler interface {
	Init() string
	Terminate(string, interface{}) string
}

// State machine type.
type FSM struct {
	Handler        Handler
	handlerValue   reflect.Value
	handlerFuncs   map[string]reflect.Value
	state          string
	transitionChan chan *transition
	timeoutChan    <-chan time.Time
}

// Create a new finite state machine.
func New(h Handler, timeout time.Duration) *FSM {
	var bufferSize int

	if timeout > 0 {
		bufferSize = int(timeout / 1e3)
	} else {
		bufferSize = 10
	}

	fsm := &FSM{
		Handler:        h,
		handlerFuncs:   make(map[string]reflect.Value),
		state:          h.Init(),
		transitionChan: make(chan *transition, bufferSize),
	}

	if timeout > 0 {
		fsm.timeoutChan = time.After(timeout)
	}

	fsm.analyze()

	go fsm.backend()

	return fsm
}

// Send a payload to handle and return the result.
func (fsm *FSM) SendWithResult(payload interface{}) interface{} {
	t := &transition{payload, make(chan interface{})}

	fsm.transitionChan <- t

	return <-t.resultChan
}

// Send a payload with no result.
func (fsm *FSM) Send(payload interface{}) {
	t := &transition{payload, nil}

	fsm.transitionChan <- t
}

// Send a payload with no result after a given time.
func (fsm *FSM) SendAfter(payload interface{}, after time.Duration) {
	saf := func() {
		time.Sleep(after)
		fsm.Send(payload)
	}
	go saf()
}

// Return the current state.
func (fsm *FSM) State() string {
	return fsm.state
}

// Analyze the event handler and prepare the state table.
func (fsm *FSM) analyze() {
	prefix := "HandleState"

	fsm.handlerValue = reflect.ValueOf(fsm.Handler)

	num := fsm.handlerValue.Type().NumMethod()

	for i := 0; i < num; i++ {
		meth := fsm.handlerValue.Type().Method(i)

		if (meth.PkgPath == "") && (strings.HasPrefix(meth.Name, prefix)) {
			if (meth.Type.NumIn() == 2) && (meth.Type.NumOut() == 2) {
				state := meth.Name[len(prefix):len(meth.Name)]

				fsm.handlerFuncs[state] = meth.Func
			}
		}
	}
}

// State machine backend.
func (fsm *FSM) backend() {
	// Message loop.
	for {
		select {
		case t := <-fsm.transitionChan:
			// Regular transition.

			if nextState, ok := fsm.handle(t); ok {
				// Continue.

				fsm.state = nextState
			} else {
				// Stop processing.

				fsm.state = fsm.Handler.Terminate(fsm.state, nextState)

				return
			}
		case to := <-fsm.timeoutChan:
			// Timeout signal resent to let it be handled.

			t := &transition{Timeout(to), nil}

			fsm.transitionChan <- t
		}
	}
}

// Handle a transition.
func (fsm *FSM) handle(t *transition) (string, bool) {
	condition := &Condition{time.Now(), t.payload}
	handlerFunc := fsm.handlerFuncs[fsm.state]
	handlerArgs := make([]reflect.Value, 2)

	handlerArgs[0] = fsm.handlerValue
	handlerArgs[1] = reflect.ValueOf(condition)

	results := handlerFunc.Call(handlerArgs)

	nextState := results[0].Interface().(string)
	result := results[1].Interface()

	// Return a result if wanted.

	if t.resultChan != nil {
		t.resultChan <- result
	}

	// Check for termination.

	if nextState == "Terminate" {
		return nextState, false
	}

	return nextState, true
}

// EOF
