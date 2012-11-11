// Tideland Common Go Library - Event Bus - Errors
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
	"fmt"
)

//--------------------
// ERRORS
//--------------------

// DuplicateTickerError will be returned if a ticker id already exists.
type DuplicateTickerError struct {
	Id string
}

// Error returns the error as string.
func (e *DuplicateTickerError) Error() string {
	return fmt.Sprintf("ticker %q already exists", e.Id)
}

// IsDuplicateTickerError tests the error type.
func IsDuplicateTickerError(err error) bool {
	_, ok := err.(*DuplicateTickerError)
	return ok
}

// TickerNotFoundError will be returned if a ticker id does not exist.
type TickerNotFoundError struct {
	Id string
}

// Error returns the error as string.
func (e *TickerNotFoundError) Error() string {
	return fmt.Sprintf("ticker %q not found", e.Id)
}

// IsTickerNotFoundError tests the error type.
func IsTickerNotFoundError(err error) bool {
	_, ok := err.(*TickerNotFoundError)
	return ok
}

// DuplicateAgentIdError will be returned if an agent id is already
// known at registration.
type DuplicateAgentIdError struct {
	Id string
}

// Error returns the error as string.
func (e *DuplicateAgentIdError) Error() string {
	return fmt.Sprintf("agent %q already exists", e.Id)
}

// IsDuplicateAgentIdError tests the error type.
func IsDuplicateAgentIdError(err error) bool {
	_, ok := err.(*DuplicateAgentIdError)
	return ok
}

// AgentNotRegisteredError will be returned if an agent is not
// registered and shall be unregistered, subscribing or unsubscribing.
type AgentNotRegisteredError struct {
	Id string
}

// Error returns the error as string.
func (e *AgentNotRegisteredError) Error() string {
	return fmt.Sprintf("agent %q is not registered", e.Id)
}

// IsAgentNotRegisteredError tests the error type.
func IsAgentNotRegisteredError(err error) bool {
	_, ok := err.(*AgentNotRegisteredError)
	return ok
}

// NoSubscriberError will be returned if no agent has subscribed 
// to the topic.
type NoSubscriberError struct {
	Topic string
}

// Error returns the error as string.
func (e *NoSubscriberError) Error() string {
	return fmt.Sprintf("no agent subscribed to topic %q", e.Topic)
}

// IsNoSubscriberError tests the error type.
func IsNoSubscriberError(err error) bool {
	_, ok := err.(*NoSubscriberError)
	return ok
}

// EOF
