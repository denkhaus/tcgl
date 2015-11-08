// Tideland Common Go Library - Sort - Unit Tests
//
// Copyright (C) 2009-2012 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed 
// by the new BSD license.

package sort

//--------------------
// IMPORTS
//--------------------

import (
	"github.com/denkhaus/tcgl/asserts"
	"math/rand"
	"sort"
	"testing"
	"time"
)

//--------------------
// TESTS
//--------------------

// Test pivot.
func TestPivot(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	// Make some test data.
	td := ByteSlice{17, 20, 13, 15, 51, 6, 21, 11, 23, 47, 59, 88, 78, 67, 94}
	plh, puh := partition(td, 0, len(td)-1)
	// Asserts.
	assert.Equal(plh, 3, "Pivot lower half.")
	assert.Equal(puh, 5, "Pivot upper half.")
	assert.Equal(td[puh-1], byte(17), "Data at median.")
	assert.Equal(td, ByteSlice{11, 13, 15, 6, 17, 20, 21, 94, 23, 47, 59, 88, 78, 67, 51}, "Prepared data.")
}

// Test sort shootout.
func TestSort(t *testing.T) {
	// Make some test data.
	isa := generateIntSlice(100000)
	isb := generateIntSlice(100000)
	isc := generateIntSlice(100000)
	isd := generateIntSlice(100000)
	// No use different sorts.
	dqs := duration(func() { sort.Sort(isb) })
	dis := duration(func() { insertionSort(isc, 0, len(isc)-1) })
	dsqs := duration(func() { sequentialQuickSort(isd, 0, len(isd)-1) })
	dpqs := duration(func() { Sort(isa) })
	// Log durations.
	t.Logf("            QS: %v", dqs)
	t.Logf("Insertion Sort: %v", dis)
	t.Logf(" Sequential QS: %v", dsqs)
	t.Logf("   Parallel QS: %v", dpqs)
}

//--------------------
// HELPERS
//--------------------

// ByteSlice is a number of bytes for sorting implementing the sort.Interface.
type ByteSlice []byte

func (bs ByteSlice) Len() int           { return len(bs) }
func (bs ByteSlice) Less(i, j int) bool { return bs[i] < bs[j] }
func (bs ByteSlice) Swap(i, j int)      { bs[i], bs[j] = bs[j], bs[i] }

// generateIntSlice generates a slice of ints.
func generateIntSlice(count int) sort.IntSlice {
	is := make([]int, count)
	for i := 0; i < count; i++ {
		is[i] = rand.Int()
	}
	return is
}

// duration measures the duration of a function execution.
func duration(f func()) time.Duration {
	start := time.Now()
	f()
	return time.Now().Sub(start)
}

// EOF
