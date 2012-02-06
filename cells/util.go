// Tideland Common Go Library - Cells - Utilities
//
// Copyright (C) 2010-2012 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed 
// by the new BSD license.

package cells

//--------------------
// IMPORTS
//--------------------

import (
	"code.google.com/p/tcgl/util"
)

//--------------------
// LOGGING
//--------------------

// logger is the util.Logger used by cells.
var logger util.Logger

// Logger returns the configured used by cells.
func Logger() util.Logger {
	return logger
}

// SetLogger sets the logger that's used by cells.
func SetLogger(l util.Logger) {
	logger = l
}

// init configures the standard logger.
func init() {
	logger = util.NewDefaultLogger("cells")
}

//--------------------
// SUBSCRIPTION MAP
//--------------------

// subscription is a pair of id and handler for subscribe
// and unsubscribe operations.
type subscription struct {
	id string
	h  Handler
}

// subscriptionMap manages the subsciptions of a handler to
// a cell or input.
type subscriptionMap map[string]Handler

// newSubscriptionMap creates a map.
func newSubscriptionMap() subscriptionMap {
	return make(subscriptionMap)
}

// subscribe adds a handler to a the map.
func (sm subscriptionMap) subscribe(s *subscription) {
	sm[s.id] = s.h
}

// unsubscribe removes a handler from a map.
func (sm subscriptionMap) unsubscribe(s *subscription) {
	delete(sm, s.id)
}

// unsubscribeAll removes all handlers from a map.
func (sm subscriptionMap) unsubscribeAll() {
	for id, _ := range sm {
		delete(sm, id)
	}
}

// handleEvent lets all subscribed handlers handle an event.
func (sm subscriptionMap) handleEvent(e Event) {
	if e.Targets() == nil {
		// Targets are all subscribed handlers.
		for id, h := range sm {
			sm.secureHandleEvent(id, h, e)
		}
	} else {
		// Target are some of the handlers.
		for _, id := range e.Targets() {
			if h, ok := sm[id]; ok {
				sm.secureHandleEvent(id, h, e)			
			}
		}
	}
}

// secureHandleEvent provides a secure wrapper of a handler call
// with logging in case of an error.
func (sm subscriptionMap) secureHandleEvent(id string, h Handler, e Event) {
	defer func() {
		if err := recover(); err != nil {
			logger.Errorf("'%s' cannot handle event '%s': %v", id, e.Topic(), err)
		}
	}()
	h.HandleEvent(e)
}

// EOF