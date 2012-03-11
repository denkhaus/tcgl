// Tideland Common Go Library - Asserts
//
// Copyright (C) 2012 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed 
// by the new BSD license.

package asserts

//--------------------
// IMPORTS
//--------------------

import (
	"bytes"
	"fmt"
	"path"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

//--------------------
// CONST
//--------------------

const RELEASE = "Tideland Common Go Library - Asserts - Release 2012-03-01"

//--------------------
// TEST
//--------------------

// Test represents the test inside an assert.
type Test uint

const (
	Invalid Test = iota
	True
	False
	Nil
	NotNil
	Equal
	Different
	About
	Match
	ErrorMatch
	Implements
	Assignable
	Unassignable
	Empty
	NotEmpty
	Length
)

var testNames = []string{
	Invalid:      "invalid",
	True:         "true",
	False:        "false",
	Nil:          "nil",
	NotNil:       "not nil",
	Equal:        "equal",
	Different:    "different",
	About:	      "about",
	Match:        "match",
	ErrorMatch:   "error match",
	Implements:   "implements",
	Assignable:   "assignable",
	Unassignable: "unassignable",
	Empty:        "empty",
	NotEmpty:     "not empty",
	Length:       "length",
}

func (t Test) String() string {
	if int(t) < len(testNames) {
		return testNames[t]
	}
	return "invalid"
}

//--------------------
// FAIL FUNC
//--------------------

// FailFunc is a user defined function that will be call by an assert if
// a test fails.
type FailFunc func(test Test, obtained, expected interface{}, msg string) bool

// panicFailFunc just panics if an assert fails.
func panicFailFunc(test Test, obtained, expected interface{}, msg string) bool {
	var obex string
	switch test {
	case True, False, Nil, NotNil, Empty, NotEmpty:
		obex = fmt.Sprintf("'%v'", obtained)
	case Implements, Assignable, Unassignable:
		obex = fmt.Sprintf("'%v' <> '%v'", ValueDescription(obtained), ValueDescription(expected))
	default:
		obex = fmt.Sprintf("'%v' <> '%v'", obtained, expected)
	}
	panic(fmt.Sprintf("assert '%s' failed: %s (%s)", test, obex, msg))
	return false
}

// generateTestingFailFunc creates a fail func bound to a testing.T.
func generateTestingFailFunc(t *testing.T, fail bool) FailFunc {
	return func(test Test, obtained, expected interface{}, msg string) bool {
		pc, file, line, _ := runtime.Caller(2)
		_, fileName := path.Split(file)
		funcNameParts := strings.Split(runtime.FuncForPC(pc).Name(), ".")
		funcNamePartsIdx := len(funcNameParts) - 1
		funcName := funcNameParts[funcNamePartsIdx]
		buffer := &bytes.Buffer{}
		fmt.Fprintf(buffer, "--------------------------------------------------------------------------------\n")
		fmt.Fprintf(buffer, "Assert '%s' failed!\n\n", test)
		fmt.Fprintf(buffer, "Filename: %s\n", fileName)
		fmt.Fprintf(buffer, "Function: %s()\n", funcName)
		fmt.Fprintf(buffer, "Line    : %d\n", line)
		switch test {
		case True, False, Nil, NotNil, Empty, NotEmpty:
			fmt.Fprintf(buffer, "Obtained: %v\n", obtained)
		case Implements, Assignable, Unassignable:
			fmt.Fprintf(buffer, "Obtained: %v\n", ValueDescription(obtained))
			fmt.Fprintf(buffer, "Expected: %v\n", ValueDescription(expected))
		default:
			fmt.Fprintf(buffer, "Obtained: %v\n", obtained)
			fmt.Fprintf(buffer, "Expected: %v\n", expected)
		}
		fmt.Fprintf(buffer, "Message : %s\n", msg)
		fmt.Fprintf(buffer, "--------------------------------------------------------------------------------\n")
		fmt.Print(buffer)
		if fail {
			t.Fail()
		}
		return false
	}
}

//--------------------
// ASSERT
//--------------------

// Asserts instances provide the test methods.
type Asserts struct {
	failFunc FailFunc
}

// NewAsserts creates a new asserts instance.
func NewAsserts(ff FailFunc) *Asserts {
	return &Asserts{ff}
}

// NewPanicAsserts creates a new asserts instance which panics if a test fails.
func NewPanicAsserts() *Asserts {
	return NewAsserts(panicFailFunc)
}

// NewTestingAsserts creates a new asserts instance for use with the testing package.
func NewTestingAsserts(t *testing.T, fail bool) *Asserts {
	return NewAsserts(generateTestingFailFunc(t, fail))
}

// True tests if obtained is true.
func (a Asserts) True(obtained bool, msg string) bool {
	if obtained == false {
		return a.failFunc(True, obtained, true, msg)
	}
	return true
}

// False tests if obtained is false.
func (a Asserts) False(obtained bool, msg string) bool {
	if obtained == true {
		return a.failFunc(False, obtained, false, msg)
	}
	return true
}

// Nil tests if obtained is nil.
func (a Asserts) Nil(obtained interface{}, msg string) bool {
	if !isNil(obtained) {
		return a.failFunc(Nil, obtained, nil, msg)
	}
	return true
}

// NotNil tests if obtained is not nil.
func (a Asserts) NotNil(obtained interface{}, msg string) bool {
	if isNil(obtained) {
		return a.failFunc(NotNil, obtained, nil, msg)
	}
	return true
}

// Equal tests if obtained and expected are equal.
func (a Asserts) Equal(obtained, expected interface{}, msg string) bool {
	if !reflect.DeepEqual(obtained, expected) {
		return a.failFunc(Equal, obtained, expected, msg)
	}
	return true
}

// Different tests if obtained and expected are different.
func (a Asserts) Different(obtained, expected interface{}, msg string) bool {
	if reflect.DeepEqual(obtained, expected) {
		return a.failFunc(Different, obtained, expected, msg)
	}
	return true
}

// About tests if obtained and expected are near to each other (within the 
// given extend).
func (a Asserts) About(obtained, expected, extend float64, msg string) bool {
	if extend < 0.0 {
		extend = extend * (-1)
	}
	expectedMin := expected - extend
	expectedMax := expected + extend
	if obtained < expectedMin || obtained > expectedMax {
		return a.failFunc(About, obtained, expected, msg)
	}
	return true
}

// Match tests if the obtained string matches a regular expression.
func (a Asserts) Match(obtained, regex, msg string) bool {
	matches, err := regexp.MatchString("^"+regex+"$", obtained)
	if err != nil {
		return a.failFunc(Match, obtained, regex, "can't compile regex: "+err.Error())
	}
	if !matches {
		return a.failFunc(Match, obtained, regex, msg)
	}
	return true
}

// ErrorMatch tests if the obtained error as string matches a regular expression.
func (a Asserts) ErrorMatch(obtained error, regex, msg string) bool {
	matches, err := regexp.MatchString("^"+regex+"$", obtained.Error())
	if err != nil {
		return a.failFunc(ErrorMatch, obtained, regex, "can't compile regex: "+err.Error())
	}
	if !matches {
		return a.failFunc(ErrorMatch, obtained, regex, msg)
	}
	return true
}

// Implements tests if obtained implements the expected interface variable pointer.
func (a Asserts) Implements(obtained, expected interface{}, msg string) bool {
	obtainedValue := reflect.ValueOf(obtained)
	expectedValue := reflect.ValueOf(expected)
	if !obtainedValue.IsValid() {
		return a.failFunc(Implements, obtained, expected, "obtained value is invalid")
	}
	if !expectedValue.IsValid() || expectedValue.Kind() != reflect.Ptr || expectedValue.Elem().Kind() != reflect.Interface {
		return a.failFunc(Implements, obtained, expected, "expected value is no interface variable pointer")
	}
	if !obtainedValue.Type().Implements(expectedValue.Elem().Type()) {
		return a.failFunc(Implements, obtained, expected, msg)
	}
	return true
}

// Assignable tests if the types of expected and obtained are assignable.
func (a Asserts) Assignable(obtained, expected interface{}, msg string) bool {
	obtainedValue := reflect.ValueOf(obtained)
	expectedValue := reflect.ValueOf(expected)
	if !obtainedValue.Type().AssignableTo(expectedValue.Type()) {
		return a.failFunc(Assignable, obtained, expected, msg)
	}
	return true
}

// Unassignable tests if the types of expected and obtained are not assignable.
func (a Asserts) Unassignable(obtained, expected interface{}, msg string) bool {
	obtainedValue := reflect.ValueOf(obtained)
	expectedValue := reflect.ValueOf(expected)
	if obtainedValue.Type().AssignableTo(expectedValue.Type()) {
		return a.failFunc(Unassignable, obtained, expected, msg)
	}
	return true
}

// Empty tests if the len of the obtained string, array, slice
// map or channel is 0.
func (a Asserts) Empty(obtained interface{}, msg string) bool {
	// Check using the interface.
	if l, ok := obtained.(lenable); ok {
		if l.Len() != 0 {
			return a.failFunc(Empty, l.Len(), 0, msg)
		}
		return true
	}
	// Check the standard types.
	obtainedValue := reflect.ValueOf(obtained)
	obtainedKind := obtainedValue.Kind()
	switch obtainedKind {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		obtainedLen := obtainedValue.Len()
		if obtainedLen != 0 {
			return a.failFunc(Empty, obtainedLen, 0, msg)
		}
	default:
		return a.failFunc(Empty, ValueDescription(obtained), 0, 
			"obtained type is no array, chan, map, slice, string or has method Len()")
	}
	return true
}

// NotEmpty tests if the len of the obtained string, array, slice
// map or channel is greater than 0.
func (a Asserts) NotEmpty(obtained interface{}, msg string) bool {
	// Check using the interface.
	if l, ok := obtained.(lenable); ok {
		if l.Len() == 0 {
			return a.failFunc(Empty, l.Len(), 0, msg)
		}
		return true
	}
	// Check the standard types.
	obtainedValue := reflect.ValueOf(obtained)
	obtainedKind := obtainedValue.Kind()
	switch obtainedKind {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		obtainedLen := obtainedValue.Len()
		if obtainedLen == 0 {
			return a.failFunc(NotEmpty, obtainedLen, nil, msg)
		}
	default:
		return a.failFunc(NotEmpty, ValueDescription(obtained), nil, 
			"obtained type is no array, chan, map, slice, string or has method Len()")
	}
	return true
}

// Length tests if the len of the obtained string, array, slice
// map or channel is equal to the expected one.
func (a Asserts) Length(obtained interface{}, expected int, msg string) bool {
	// Check using the interface.
	if l, ok := obtained.(lenable); ok {
		if l.Len() != expected {
			return a.failFunc(Length, l.Len(), expected, msg)
		}
		return true
	}
	// Check the standard types.
	obtainedValue := reflect.ValueOf(obtained)
	obtainedKind := obtainedValue.Kind()
	switch obtainedKind {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		obtainedLen := obtainedValue.Len()
		if obtainedLen != expected {
			return a.failFunc(Length, obtainedLen, expected, msg)
		}
	default:
		return a.failFunc(Length, ValueDescription(obtained), 0, 
			"obtained type is no array, chan, map, slice, string or has method Len()")
	}
	return true
}

//--------------------
// HELPER
//--------------------

// ValueDescription returns a description of a value as string.
func ValueDescription(value interface{}) string {
	rvalue := reflect.ValueOf(value)
	kind := rvalue.Kind()
	switch kind {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice:
		return kind.String() + " of " + rvalue.Type().Elem().String()
	case reflect.Func:
		return kind.String() + " " + rvalue.Type().Name() + "()"
	case reflect.Interface, reflect.Struct:
		return kind.String() + " " + rvalue.Type().Name()
	case reflect.Ptr:
		return kind.String() + " to " + rvalue.Type().Elem().String()
	}
	// Default.
	return kind.String()
}

// lenable is an interface for the Len() mehod.
type lenable interface {
	Len() int
}

// isNil is a safer way to test if a value is nil.
func isNil(value interface{}) bool {
	if value == nil {
		// Standard test.
		return true
	} else {
		// Some types have to be tested via reflection.
		rvalue := reflect.ValueOf(value)
		kind := rvalue.Kind()
		switch kind {
		case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
			return rvalue.IsNil()
		}
	}
	return false
}

// EOF
