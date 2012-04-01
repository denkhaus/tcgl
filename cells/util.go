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
	"sort"
	"time"
)

//--------------------
// TICKER
//--------------------

// ticker provides periodic events raised at a defined id.
type ticker struct {
	env *Environment
	id string
	raiseId string
	period time.Duration
	stopChan chan bool
}

// startTicker starts a new ticker in the background.
func startTicker(env *Environment, id, raiseId string, period time.Duration) *ticker {
	t := &ticker{env, id, raiseId, period, make(chan bool)}
	go t.backend()
	return t
}

// stop lets the backend goroutine stop working.
func (t *ticker) stop() {
	t.stopChan <- true
}

// backend is the goroutine running the ticker.
func (t *ticker) backend() {
	for {
		select {
		case <-time.After(t.period):
			t.env.Raise(t.raiseId, NewTickerEvent(t.id))
		case <-t.stopChan:
			return
		}
	}
}

// TickerEvent signals a tick to ticker subscribers.
type TickerEvent struct {
	id      string
	time    time.Time
	context *Context
}

// NewTickerEvent creates a new ticker event instance with a 
// given id and the current time.
func NewTickerEvent(id string) *TickerEvent {
	return &TickerEvent{id, time.Now(), nil}
}

// Topic returns the topic of the event, here "ticker([id])".
func (te TickerEvent) Topic() string {
	return "ticker(" + te.id + ")"
}

// Payload returns the payload of the event, here the time in
// nanoseconds.
func (te TickerEvent) Payload() interface{} {
	return te.time
}

// Context returns the context of a set of event processings.
func (te TickerEvent) Context() *Context {
	return te.context
}

// SetContext set the context of a set of event processings.
func (te *TickerEvent) SetContext(c *Context) {
	te.context = c
}

//--------------------
// CELL ASSIGNMENTS
//--------------------

// assignments manages assignments of one cell id to 
// many (subscribers / subscriptions).
type assignments map[string][]string

// add some assignments between cells.
func (a assignments) add(eid string, aids ...string) {
	// Get list of ids.
	ids := a[eid]
	if ids == nil {
		ids = []string{}
	}
	// Check and append them.
	tmp := map[string]struct{}{}
	for _, id := range ids {
		tmp[id] = struct{}{}
	}
	for _, id := range aids {
		tmp[id] = struct{}{}
	}
	ids = []string{}
	for id, _ := range tmp {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	a[eid] = ids
}

// remove some assignments between cells.
func (a assignments) remove(eid string, rids ...string) {
	// Get list of ids.
	ids := a[eid]
	if ids == nil {
		return
	}
	// Check and remove them.
	tmp := map[string]struct{}{}
	for _, id := range ids {
		tmp[id] = struct{}{}
	}
	for _, id := range rids {
		delete(tmp, id)
	}
	ids = []string{}
	for id, _ := range tmp {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	a[eid] = ids
}

// all returns all assigned ids to an id.
func (a assignments) all(id string) []string {
	if ids, ok := a[id]; ok {
		return ids
	}
	return []string{}
}

// drop removes all assignments for one id.
func (a assignments) drop(id string) {
	delete(a, id)
}

// dropAll drops the cell id from all assignments.
func (a assignments) dropAll(id string, da assignments) {
	if aids, ok := a[id]; ok {
		for _, aid := range aids {
			da.remove(aid, id)
		}
	}
}

// EOF
