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
	"testing"
)

//--------------------
// TESTS
//--------------------

// Test debug statement.
func TestDebug(t *testing.T) {
	Debug("Hello, I'm debugging %v!", "here")
}

// Test the integer generator.
func TestLazyIntEvaluator(t *testing.T) {
	fibFunc := func(s interface{}) (interface{}, interface{}) {
		os := s.([]int)
		v1 := os[0]
		v2 := os[1]
		ns := []int{v2, v1 + v2}

		return v1, ns
	}

	fib := BuildLazyIntEvaluator(fibFunc, []int{0, 1})

	var fibs [25]int

	for i := 0; i < 25; i++ {
		fibs[i] = fib()
	}

	t.Logf("FIBS: %v", fibs)
}

// EOF
