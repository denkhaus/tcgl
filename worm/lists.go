// Tideland Common Go Library - Write once read multiple - Lists
//
// Copyright (C) 2012 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed 
// by the new BSD license.

package worm

//--------------------
// IMPORTS
//--------------------

import (
	"sort"
)

//--------------------
// INT LIST
//--------------------

// Ints is an int slice for the data exchange with lists and sets.
type Ints []int

// IntList contains ints in a ordered list.
type IntList struct {
	values []int
}

// NewIntList creates a new list of ints.
func NewIntList(values Ints) IntList {
	i := IntList{make([]int, len(values))}
	copy(i.values, values)
	return i
}

// Len returns the number of values in the list.
func (i IntList) Len() int {
	return len(i.values)
}

// Values returns the values of the list.
func (i IntList) Values() Ints {
	values := make(Ints, len(i.values))
	copy(values, i.values)
	return values
}

// SortedValues returns the values of the list sorted.
func (i IntList) SortedValues() Ints {
	values := make(Ints, len(i.values))
	copy(values, i.values)
	sort.Ints(values)
	return values
}

// Append creates a new list with the passed values appended.
func (i IntList) Append(values Ints) IntList {
	ni := NewIntList(i.values)
	ni.values = append(ni.values, values...)
	return ni
}

// Filter creates a new list out of those values for which the
// filter function returns true.
func (i IntList) Filter(f func(int) bool) IntList {
	values := []int{}
	for _, v := range i.values {
		if f(v) {
			values = append(values, v)
		}
	}
	return NewIntList(values)
}

//--------------------
// STRING LIST
//--------------------

// Strings is an int slice for the data exchange with lists and sets.
type Strings []string

// StringList contains strings in a ordered list.
type StringList struct {
	values []string
}

// NewStringList creates a new list of strings.
func NewStringList(values Strings) StringList {
	s := StringList{make([]string, len(values))}
	copy(s.values, values)
	return s
}

// Len returns the number of values in the list.
func (s StringList) Len() int {
	return len(s.values)
}

// Values returns the values of the list.
func (s StringList) Values() Strings {
	values := make(Strings, len(s.values))
	copy(values, s.values)
	return values
}

// SortedValues returns the values of the list sorted.
func (s StringList) SortedValues() Strings {
	values := make(Strings, len(s.values))
	copy(values, s.values)
	sort.Strings(values)
	return values
}

// Append creates a new list with the passed values appended.
func (s StringList) Append(values Strings) StringList {
	ns := NewStringList(s.values)
	ns.values = append(ns.values, values...)
	return ns
}

// Filter creates a new list out of those values for which the
// filter function returns true.
func (s StringList) Filter(f func(string) bool) StringList {
	values := []string{}
	for _, v := range s.values {
		if f(v) {
			values = append(values, v)
		}
	}
	return NewStringList(values)
}

// EOF
