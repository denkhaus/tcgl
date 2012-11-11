// Tideland Common Go Library - Write once read multiple - Sets
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
// INT SET
//--------------------

// IntSet contains ints only once.
type IntSet struct {
	values map[int]struct{}
}

// NewIntSet creates a new set of ints.
func NewIntSet(values Ints) IntSet {
	i := IntSet{make(map[int]struct{}, len(values))}
	if values != nil {
		for _, value := range values {
			if _, ok := i.values[value]; !ok {
				i.values[value] = struct{}{}
			}
		}
	}
	return i
}

// Len returns the number of values in the set.
func (i IntSet) Len() int {
	return len(i.values)
}

// Values returns the values of the set.
func (i IntSet) Values() Ints {
	values := make(Ints, len(i.values))
	ctr := 0
	for value := range i.values {
		values[ctr] = value
		ctr++
	}
	sort.Ints(values)
	return values
}

// Apply creates a new set with all passed values and those
// of this set which are not in the values.
func (i IntSet) Apply(values Ints) IntSet {
	ni := NewIntSet(values)
	for value := range i.values {
		if _, ok := ni.values[value]; !ok {
			ni.values[value] = struct{}{}
		}
	}
	return ni
}

// Contains tests if all the passed values are in the set.
func (i IntSet) Contains(values ...int) bool {
	for _, value := range values {
		if _, ok := i.values[value]; !ok {
			return false
		}
	}
	return true
}

//--------------------
// STRING SET
//--------------------

// StringSet contains strings only once.
type StringSet struct {
	values map[string]struct{}
}

// NewStringSet creates a new set of strings.
func NewStringSet(values Strings) StringSet {
	s := StringSet{make(map[string]struct{}, len(values))}
	if values != nil {
		for _, value := range values {
			if _, ok := s.values[value]; !ok {
				s.values[value] = struct{}{}
			}
		}
	}
	return s
}

// Len returns the number of values in the set.
func (s StringSet) Len() int {
	return len(s.values)
}

// Values returns the values of the set.
func (s StringSet) Values() Strings {
	values := make(Strings, len(s.values))
	ctr := 0
	for value := range s.values {
		values[ctr] = value
		ctr++
	}
	sort.Strings(values)
	return values
}

// Apply creates a new set with all passed values and those
// of this set which are not in the values.
func (s StringSet) Apply(values Strings) StringSet {
	ns := NewStringSet(values)
	for value := range s.values {
		if _, ok := ns.values[value]; !ok {
			ns.values[value] = struct{}{}
		}
	}
	return ns
}

// Contains tests if all the passed values are in the set.
func (s StringSet) Contains(values ...string) bool {
	for _, value := range values {
		if _, ok := s.values[value]; !ok {
			return false
		}
	}
	return true
}

// EOF
