// Tideland Common Go Library - Finite State Machine - Unit Tests
//
// Copyright (C) 2009-2011 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed 
// by the new BSD license.

package state

//--------------------
// IMPORTS
//--------------------

import (
	"log"
	"testing"
	"time"
)

//--------------------
// TESTS
//--------------------

// Test the finite state machine successfully.
func TestFsmSuccess(t *testing.T) {
	fsm := New(NewLoginHandler(), -1)

	fsm.Send(&LoginPayload{"yadda"})
	fsm.Send(&PreparePayload{"foo", "bar"})
	fsm.Send(&LoginPayload{"yaddaA"})
	fsm.Send(&LoginPayload{"yaddaB"})
	fsm.Send(&LoginPayload{"yaddaC"})
	fsm.Send(&LoginPayload{"yaddaD"})
	fsm.Send(&UnlockPayload{})
	fsm.Send(&LoginPayload{"bar"})

	time.Sleep(1e7)

	t.Logf("Status: '%v'.", fsm.State())
}

// Test the finite state machine with timeout.
func TestFsmTimeout(t *testing.T) {
	fsm := New(NewLoginHandler(), 1e5)

	fsm.Send(&LoginPayload{"yadda"})
	fsm.Send(&PreparePayload{"foo", "bar"})
	fsm.Send(&LoginPayload{"yaddaA"})
	fsm.Send(&LoginPayload{"yaddaB"})

	time.Sleep(1e8)

	fsm.Send(&LoginPayload{"yaddaC"})
	fsm.Send(&LoginPayload{"yaddaD"})
	fsm.Send(&UnlockPayload{})
	fsm.Send(&LoginPayload{"bar"})

	time.Sleep(1e7)

	t.Logf("Status: '%v'.", fsm.State())
}

//--------------------
// HELPER: TEST LOGIN EVENT HANDLER
//--------------------

// Prepare payload.
type PreparePayload struct {
	userId   string
	password string
}

// Login payload.
type LoginPayload struct {
	password string
}

// Reset payload.
type ResetPayload struct{}

// Unlock payload.
type UnlockPayload struct{}

// Login handler tyoe.
type LoginHandler struct {
	userId              string
	password            string
	illegalLoginCounter int
	locked              bool
}

// Create a new login handler.
func NewLoginHandler() *LoginHandler {
	return new(LoginHandler)
}

// Return the initial state.
func (lh *LoginHandler) Init() string {
	return "New"
}

// Terminate the handler.
func (lh *LoginHandler) Terminate(string, interface{}) string {
	return "LoggedIn"
}

// Handler for state: "New".
func (lh *LoginHandler) HandleStateNew(c *Condition) (string, interface{}) {
	switch pld := c.Payload.(type) {
	case *PreparePayload:
		lh.userId = pld.userId
		lh.password = pld.password
		lh.illegalLoginCounter = 0
		lh.locked = false

		log.Printf("User '%v' prepared.", lh.userId)

		return "Authenticating", nil
	case *LoginPayload:
		log.Printf("Illegal login, handler not initialized!")

		return "New", false
	case Timeout:
		log.Printf("Timeout, terminate handler!")

		return "Terminate", nil
	}

	log.Printf("Illegal payload '%v' during state 'new'!", c.Payload)

	return "New", nil
}

// Handler for state: "Authenticating".
func (lh *LoginHandler) HandleStateAuthenticating(c *Condition) (string, interface{}) {
	switch pld := c.Payload.(type) {
	case *LoginPayload:
		if pld.password == lh.password {
			lh.illegalLoginCounter = 0
			lh.locked = false

			log.Printf("User '%v' logged in.", lh.userId)

			return "Terminate", true
		}

		log.Printf("User '%v' used illegal password.", lh.userId)

		lh.illegalLoginCounter++

		if lh.illegalLoginCounter == 3 {
			lh.locked = true

			log.Printf("User '%v' locked!", lh.userId)

			return "Locked", false
		}

		return "Authenticating", false
	case *UnlockPayload:
		log.Printf("No need to unlock user '%v'!", lh.userId)

		return "Authenticating", nil
	case *ResetPayload, Timeout:
		lh.illegalLoginCounter = 0
		lh.locked = false

		log.Printf("User '%v' resetted.", lh.userId)

		return "Authenticating", nil
	}

	log.Printf("Illegal payload '%v' during state 'authenticating'!", c.Payload)

	return "Authenticating", nil
}

// Handler for state: "Locked".
func (lh *LoginHandler) HandleStateLocked(c *Condition) (string, interface{}) {
	switch pld := c.Payload.(type) {
	case *LoginPayload:
		log.Printf("User '%v' login rejected, user is locked!", lh.userId)

		return "Locked", false
	case *ResetPayload, *UnlockPayload, Timeout:
		lh.illegalLoginCounter = 0
		lh.locked = false

		log.Printf("User '%v' resetted / unlocked.", lh.userId)

		return "Authenticating", nil
	}

	log.Printf("Illegal payload '%v' during state 'loacked'!", c.Payload)

	return "Locked", nil
}

// EOF
