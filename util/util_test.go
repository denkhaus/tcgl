// Tideland Common Go Library - Utilities - Unit Tests
//
// Copyright (C) 2009-2012 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed 
// by the new BSD license.

package util

//--------------------
// IMPORTS
//--------------------

import (
	"code.google.com/p/tcgl/asserts"
	"testing"
)

//--------------------
// TESTS
//--------------------

// Test the method dispatch function.
func TestDispatch(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	on := Switch{true}
	result, err := Dispatch(&on, "String")
	assert.Nil(err, "Dispatch String() should return no error")
	assert.Equal(result, "on", "Active switch as string is 'on'")
	result, err = Dispatch(&on, "Set", false)
	assert.Nil(err, "Dispatch Set() should return no error")
	assert.Equal(on.String(), "off", "Inactive switch as string is 'off'")
}

// Test the integer generator.
func TestLazyIntEvaluator(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	fibFunc := func(s interface{}) (interface{}, interface{}) {
		os := s.([]int)
		v1 := os[0]
		v2 := os[1]
		ns := []int{v2, v1 + v2}
		return v1, ns
	}
	fib := BuildLazyIntEvaluator(fibFunc, []int{0, 1})
	// Assert the first calls.
	assert.Equal(fib(), 0, "1st fib call.")
	assert.Equal(fib(), 1, "2nd fib call.")
	assert.Equal(fib(), 1, "3rd fib call.")
	assert.Equal(fib(), 2, "4th fib call.")
	assert.Equal(fib(), 3, "5th fib call.")
	assert.Equal(fib(), 5, "6th fib call.")
	assert.Equal(fib(), 8, "7th fib call.")
	assert.Equal(fib(), 13, "8th fib call.")
	assert.Equal(fib(), 21, "9th fib call.")
}

//--------------------
// HELPER
//--------------------

type Switch struct {
	value bool
}

func (s *Switch) Set(v bool) {
	s.value = v
}

func (s Switch) String() string {
	if s.value {
		return "on"
	}
	return "off"
}

// EOF
