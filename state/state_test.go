// Tideland Common Go Library - Finite State Machine - Unit Tests
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
	"cgl.tideland.biz/asserts"
	"log"
	"testing"
	"time"
)

//--------------------
// TESTS
//--------------------

// Test the finite state machine successfully.
func TestFsmSuccess(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	// Create some test data.
	fsm := New(NewLoginHandler(), 5*time.Minute)
	fsm.Handle("login", &LoginData{"yadda", "yadda"})
	fsm.Handle("prepare", &LoginData{"foo", "bar"})
	fsm.Handle("login", &LoginData{"foo", "yadda"})
	fsm.Handle("login", &LoginData{"foo", "yadda"})
	fsm.Handle("login", &LoginData{"foo", "yadda"})
	fsm.Handle("login", &LoginData{"foo", "yadda"})

	assert.Equal(fsm.State(), "locked", "FSM is locked.")

	fsm.Handle("unlock", &LoginData{"foo", ""})
	fsm.Handle("login", &LoginData{"foo", "bar"})

	assert.Equal(fsm.State(), "terminated", "FSM terminated.")
}

// Test the finite state machine with timeout.
func TestFsmTimeout(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	// Create some test data.
	fsm := New(NewLoginHandler(), 250*time.Millisecond)
	fsm.Handle("prepare", &LoginData{"foo", "bar"})
	fsm.Handle("login", &LoginData{"foo", "yadda"})
	fsm.Handle("login", &LoginData{"foo", "yadda"})

	time.Sleep(2e9)
	assert.Equal(fsm.State(), "new", "FSM is timed-out.")

	fsm.Handle("prepare", &LoginData{"foo", "bar"})
	fsm.Handle("login", &LoginData{"foo", "bar"})

	assert.Equal(fsm.State(), "terminated", "FSM terminated.")
}

// Test the finite state machine having an error.
func TestFsmError(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	// Create some test data.
	fsm := New(NewLoginHandler(), 250*time.Millisecond)
	fsm.Handle("prepare", &LoginData{"foo", "bar"})
	fsm.Handle("bullshit", &LoginData{"", ""})
	fsm.Handle("login", &LoginData{"foo", "yadda"})

	assert.Equal(fsm.State(), "terminated", "FSM terminated after error.")
}

//--------------------
// HELPER: TEST LOGIN EVENT HANDLER
//--------------------

// LoginData encapsulates the data for the login handler.
type LoginData struct {
	UserId   string
	Password string
}

// Login handler tyoe.
type LoginHandler struct {
	userId              string
	password            string
	illegalLoginCounter int
	locked              bool
	ticks               int
}

// Create a new login handler.
func NewLoginHandler() *LoginHandler {
	return &LoginHandler{}
}

// Return the initial state.
func (lh *LoginHandler) Init() (*HandlerMap, string) {
	hm := NewHandlerMap(lh)
	hm.Assign("new", "HandleNew")
	hm.Assign("authenticating", "HandleAuthenticating")
	hm.Assign("locked", "HandleLocked")
	return hm, "new"
}

func (lh *LoginHandler) Error(t *Transition, err error) string {
	log.Printf("Handle error: %v", err)
	lh.init()
	return "terminate"
}

func (lh *LoginHandler) Terminate() {
	log.Printf("Terminating.")
}

// Handler for state: "new".
func (lh *LoginHandler) HandleNew(t *Transition) string {
	ld, _ := t.Payload.(*LoginData)
	switch t.Command {
	case "prepare":
		lh.userId = ld.UserId
		lh.password = ld.Password
		lh.illegalLoginCounter = 0
		lh.locked = false
		log.Printf("User '%v' prepared.", lh.userId)
		return "authenticating"
	case "login":
		log.Printf("Illegal login, handler not initialized!")
		return "new"
	case "tick":
		log.Printf("Got a new tick.")
		return "new"
	}
	log.Printf("Illegal command %q during state 'new'!", t.Command)
	return "new"
}

// Handler for state: "authenticating".
func (lh *LoginHandler) HandleAuthenticating(t *Transition) string {
	ld, _ := t.Payload.(*LoginData)
	switch t.Command {
	case "login":
		if ld.Password == lh.password {
			lh.illegalLoginCounter = 0
			lh.locked = false
			lh.ticks = 0
			log.Printf("User '%v' logged in.", lh.userId)
			return "terminate"
		}
		log.Printf("User '%v' used illegal password.", lh.userId)
		lh.illegalLoginCounter++
		if lh.illegalLoginCounter == 3 {
			lh.locked = true
			log.Printf("User '%v' locked!", lh.userId)
			return "locked"
		}
		return "authenticating"
	case "unlock":
		log.Printf("No need to unlock user '%v'!", lh.userId)
		return "authenticating"
	case "reset":
		lh.illegalLoginCounter = 0
		lh.locked = false
		lh.ticks = 0
		log.Printf("User '%v' resetted.", lh.userId)
		return "authenticating"
	case "bullshit":
		log.Printf("Got bullshit commmand.")
		return "i-dont-know-what-to-do"
	case "tick":
		log.Printf("Got a tick.")
		lh.ticks++
		if lh.ticks > 5 {
			lh.init()
			return "new"
		}
		return "authenticating"
	}
	log.Printf("Illegal command %q during state 'authenticating'!", t.Command)
	return "authenticating"
}

// Handler for state: "Locked".
func (lh *LoginHandler) HandleLocked(t *Transition) string {
	ld, _ := t.Payload.(*LoginData)
	switch t.Command {
	case "login":
		log.Printf("User %q login rejected, user is locked!", ld.UserId)
		return "locked"
	case "reset", "unlock":
		lh.illegalLoginCounter = 0
		lh.locked = false
		lh.ticks = 0
		log.Printf("User %q resetted / unlocked.", ld.UserId)
		return "authenticating"
	case "tick":
		log.Printf("Got a locked tick.")
		return "locked"
	}
	log.Printf("Illegal command %q during state 'loacked'!", t.Command)
	return "locked"
}

func (lh *LoginHandler) init() {
	lh.userId = ""
	lh.password = ""
	lh.illegalLoginCounter = 0
	lh.locked = false
	lh.ticks = 0
}

// EOF
