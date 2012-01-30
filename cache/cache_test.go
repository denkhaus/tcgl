// Tideland Common Go Library - Cache - Unit Test
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
	"log"
	"testing"
	"time"
)

//--------------------
// TESTS
//--------------------

// Test cache.
func TestCache(t *testing.T) {
	ctr := 0
	count := func() (interface{}, error) {
		ctr++

		return ctr, nil
	}
	cv := NewCachedValue(count, 1e9)
	retrieve := func() int { v, _ := cv.Value(); return v.(int) }

	log.Printf("1st cache access: %v", retrieve())
	log.Printf("2nd cache access: %v", retrieve())
	time.Sleep(15e8)
	log.Printf("3rd cache access: %v", retrieve())
	time.Sleep(15e8)
	log.Printf("4th cache access: %v", retrieve())
	log.Printf("5th cache access: %v", retrieve())
}

// EOF
