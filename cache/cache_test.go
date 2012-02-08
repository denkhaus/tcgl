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
	"code.google.com/p/tcgl/asserts"
	"testing"
	"time"
)

//--------------------
// TESTS
//--------------------

// Test cache.
func TestCache(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	ctr := 0
	count := func() (interface{}, error) {
		ctr++

		return ctr, nil
	}
	cv := NewCachedValue(count, 5e8)
	retrieve := func() int { v, _ := cv.Value(); return v.(int) }

	assert.Equal(retrieve(), 1, "1st cache access")
	assert.Equal(retrieve(), 1, "2nd cache access")
	time.Sleep(1e9)
	assert.Equal(retrieve(), 2, "3rd cache access")
	time.Sleep(1e9)
	assert.Equal(retrieve(), 3, "4th cache access")
	assert.Equal(retrieve(), 3, "5th cache access")
}

// EOF
