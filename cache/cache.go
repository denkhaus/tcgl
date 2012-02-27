// Tideland Common Go Library - Cache
//
// Copyright (C) 2009-2012 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed 
// by the new BSD license.

package cache

//--------------------
// IMPORTS
//--------------------

import (
	"sync"
	"time"
)

//--------------------
// CONST
//--------------------

const RELEASE = "Tideland Common Go Library - Cache - Release 2012-02-15"

//--------------------
// CONST
//--------------------

// Signals to stop or restart a cached value.
const (
	sigStop    = true
	sigRestart = false
)

//--------------------
// CACHED VALUE
//--------------------

// RetrievalFunc is the signature of a function responsible for the retrieval
// of the cached value from somewhere else in the system, e.g. a database.
type RetrievalFunc func() (interface{}, error)

// CachedValue retrieves and caches a value for a defined time. After
// that time the cache is cleared and set again automatically when accessed.
type CachedValue struct {
	value         interface{}
	retrievalFunc RetrievalFunc
	ttl           time.Duration
	mutex         sync.Mutex
	ticker        *time.Ticker
	signalChan    chan bool
}

// NewCachedValue creates a new cache.
func NewCachedValue(r RetrievalFunc, ttl time.Duration) *CachedValue {
	cv := &CachedValue{
		retrievalFunc: r,
		ttl:           ttl,
		ticker:        time.NewTicker(ttl),
		signalChan:    make(chan bool),
	}
	go cv.backend()
	return cv
}

// Value returns the cached value. If an error occurred during
// retrieval that will be returned too.
func (cv *CachedValue) Value() (v interface{}, err error) {
	cv.mutex.Lock()
	defer cv.mutex.Unlock()

	if cv.value != nil {
		// Everything is fine.
		return cv.value, nil
	}
	// Retrieve the value.
	if cv.value, err = cv.retrievalFunc(); err != nil {
		// Ooops, something went wrong. Reset value
		// and return the error.
		cv.value = nil
		return nil, err
	}
	// Retrieval has been ok, restart ticker and return value.
	cv.ticker.Stop()
	cv.ticker = time.NewTicker(cv.ttl)
	cv.signalChan <- sigRestart
	return cv.value, nil
}

// Clear removes the value from the cache.
func (cv *CachedValue) Clear() {
	cv.mutex.Lock()
	defer cv.mutex.Unlock()

	cv.value = nil
}

// Stop removes everything and sends a stop signal to the backend.
func (cv *CachedValue) Stop() {
	cv.mutex.Lock()
	defer cv.mutex.Unlock()

	cv.value = nil
	cv.retrievalFunc = nil
	cv.ticker.Stop()
	cv.signalChan <- sigStop
}

// backend clears the cache in intervals until it's told to stop.
func (cv *CachedValue) backend() {
	for {
		select {
		case <-cv.ticker.C:
			// Just clear it.
			cv.Clear()
		case stop := <-cv.signalChan:
			if stop {
				// Leave the endless loop.
				return
			}
		}
	}
}

// EOF
