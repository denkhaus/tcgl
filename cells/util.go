// Tideland Common Go Library - Cells - Utilities
//
// Copyright (C) 2010-2011 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed 
// by the new BSD license.

package cells

//--------------------
// IMPORTS
//--------------------

import (
	"log"
	"unsafe"
)

//--------------------
// SUBSCRIPTION MAP
//--------------------

// subscriptionMap manages the subsciptions of a handler to
// a cell or input.
type subscriptionMap map[uintptr]Handler

// newSubscriptionMap creates a map.
func newSubscriptionMap() subscriptionMap {
	return make(subscriptionMap)
}

// subscribe adds a handler to a the map.
func (sm subscriptionMap) subscribe(h Handler) {
	sm[sm.handlerId(h)] = h
}

// unsubscribe removes a handler from a map.
func (sm subscriptionMap) unsubscribe(h Handler) {
	hid := sm.handlerId(h)
	smh, ok := sm[hid]

	if ok {
		sm[hid] = smh, false
	}
}

// unsubscribeAll removes all handlers from a map.
func (sm subscriptionMap) unsubscribeAll() {
	for id, h := range sm {
		sm[id] = h, false
	}
}

// handleEvent lets all subscribed handlers handle an event.
func (sm subscriptionMap) handleEvent(e Event) {
	for _, h := range sm {
		sm.secureHandleEvent(h, e)
	}
}

// secureHandleEvent provides a secure wrapper of a handler call
// with logging in case of an error.
func (sm subscriptionMap) secureHandleEvent(h Handler, e Event) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("[cells] error handling event '%s': %v", e.Topic(), err)
		}
	}()

	h.HandleEvent(e)
}

// Generate the map id of the handler.
func (sm subscriptionMap) handlerId(h Handler) uintptr {
	return uintptr(unsafe.Pointer(&h))
}

// EOF
