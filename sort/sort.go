// Tideland Common Go Library - Sort
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
	"runtime"
	"sort"
)

//--------------------
// CONST
//--------------------

const RELEASE = "Tideland Common Go Library - Sort - Release 2012-02-21"

//--------------------
// CONTROL VALUES
//--------------------

// Threshold for switching from sequential quick sort to insertion sort.
var sequentialThreshold int = runtime.NumCPU() * 4 - 1

// Threshold for switching from parallel to sequential quick sort.
var parallelThreshold int = runtime.NumCPU() * 2048 - 1

//--------------------
// HELPING FUNCS
//--------------------

// Simple insertion sort for smaller data collections.
func insertionSort(data sort.Interface, lo, hi int) {
	for i := lo + 1; i < hi+1; i++ {
		for j := i; j > lo && data.Less(j, j-1); j-- {
			data.Swap(j, j-1)
		}
	}
}

// Get the median based on Tukey's ninther.
func median(data sort.Interface, lo, hi int) int {
	m := (lo + hi) / 2
	d := (hi - lo) / 8
	// Move median into the middle.
	mot := func(ml, mm, mh int) {
		if data.Less(mm, ml) {
			data.Swap(mm, ml)
		}
		if data.Less(mh, mm) {
			data.Swap(mh, mm)
		}
		if data.Less(mm, ml) {
			data.Swap(mm, ml)
		}
	}
	// Get low, middle, and high median.
	if hi-lo > 40 {
		mot(lo+d, lo, lo+2*d)
		mot(m-d, m, m+d)
		mot(hi-d, hi, hi-2*d)
	}
	// Get combined median.
	mot(lo, m, hi)
	return m
}

// Partition the data based on the median.
func partition(data sort.Interface, lo, hi int) (int, int) {
	med := median(data, lo, hi)
	idx := lo
	data.Swap(med, hi)
	for i := lo; i < hi; i++ {
		if data.Less(i, hi) {
			data.Swap(i, idx)
			idx++
		}
	}
	data.Swap(idx, hi)
	return idx - 1, idx + 1
}

// Sequential quicksort using itself recursively.
func sequentialQuickSort(data sort.Interface, lo, hi int) {
	if hi-lo > sequentialThreshold {
		// Use sequential quicksort.
		plo, phi := partition(data, lo, hi)
		sequentialQuickSort(data, lo, plo)
		sequentialQuickSort(data, phi, hi)
	} else {
		// Use insertion sort.
		insertionSort(data, lo, hi)
	}
}

// Parallel quicksort using itself recursively
// and concurrent.
func parallelQuickSort(data sort.Interface, lo, hi int, done chan bool) {
	if hi-lo > parallelThreshold {
		// Parallel QuickSort.
		plo, phi := partition(data, lo, hi)
		partDone := make(chan bool)
		go parallelQuickSort(data, lo, plo, partDone)
		go parallelQuickSort(data, phi, hi, partDone)
		// Wait for the end of both sorts.
		<-partDone
		<-partDone
	} else {
		// Sequential QuickSort.
		sequentialQuickSort(data, lo, hi)
	}
	// Signal that it's done.
	done <- true
}

//--------------------
// PARALLEL QUICKSORT
//--------------------

// Sort is the single function for sorting data according
// to the standard sort interface. Internally it uses the
// parallel quicksort.
func Sort(data sort.Interface) {
	done := make(chan bool)

	go parallelQuickSort(data, 0, data.Len()-1, done)

	<-done
}

// EOF
