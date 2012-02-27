// Tideland Common Go Library - Asserts - Unit Test
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
	"errors"
	"io"
	"testing"
)

//--------------------
// FAIL FUNCS
//--------------------

// createValueAsserts returns an assert with a value logging fail func.
func createValueAsserts(t *testing.T) *Asserts {
	return NewAsserts(func(test Test, obtained, expected interface{}, msg string) bool {
		t.Logf("testing assert '%s' failed: '%v' <> '%v' (%s)", test, obtained, expected, msg)
		return false
	})
}

// createTypeAsserts returns an assert with a value description (type) logging fail func.
func createTypeAsserts(t *testing.T) *Asserts {
	return NewAsserts(func(test Test, obtained, expected interface{}, msg string) bool {
		t.Logf("testing assert '%s' failed: '%v' <> '%v' (%s)",
			test, ValueDescription(obtained), ValueDescription(expected), msg)
		return false
	})
}

//--------------------
// TESTS
//--------------------

// TestIsNilHelper test sthe isNil() helper.
func TestIsNilHelper(t *testing.T) {
	if !isNil(nil) {
		t.Errorf("nil is not nil?")
	}
	if isNil("nil") {
		t.Errorf("'nil' is nil?")
	}
	var c chan int
	if !isNil(c) {
		t.Errorf("channel is not nil?")
	}
	var f func()
	if !isNil(f) {
		t.Errorf("func is not nil?")
	}
	var i interface{}
	if !isNil(i) {
		t.Errorf("interface is not nil?")
	}
	var m map[string]string
	if !isNil(m) {
		t.Errorf("map is not nil?")
	}
	var p *bool
	if !isNil(p) {
		t.Errorf("pointer is not nil?")
	}
	var s []string
	if !isNil(s) {
		t.Errorf("slice is not nil?")
	}
}

// TestAssertTrue tests the True() assertion.
func TestAssertTrue(t *testing.T) {
	a := createValueAsserts(t)

	a.True(true, "should not fail")
	if a.True(false, "should fail and be logged") {
		t.Errorf("True() returned true")
	}
}

// TestAssertFalse tests the False() assertion.
func TestAssertFalse(t *testing.T) {
	a := createValueAsserts(t)

	a.False(false, "should not fail")
	if a.False(true, "should fail and be logged") {
		t.Errorf("False() returned true")
	}
}

// TestAssertNil tests the Nil() assertion.
func TestAssertNil(t *testing.T) {
	a := createValueAsserts(t)
	a.Nil(nil, "should not fail")
	if a.Nil("not nil", "should fail and be logged") {
		t.Errorf("Nil() returned true")
	}
}

// TestAssertNotNil tests the NotNil() assertion.
func TestAssertNotNil(t *testing.T) {
	a := createValueAsserts(t)

	a.NotNil("not nil", "should not fail")
	if a.NotNil(nil, "should fail and be logged") {
		t.Errorf("NotNil() returned true")
	}
}

// TestAssertEqual tests the Equal() assertion.
func TestAssertEqual(t *testing.T) {
	a := createValueAsserts(t)
	m := map[string]int{"one": 1, "two": 2, "three": 3}

	a.Equal(nil, nil, "should not fail")
	a.Equal(true, true, "should not fail")
	a.Equal(1, 1, "should not fail")
	a.Equal("foo", "foo", "should not fail")
	a.Equal(map[string]int{"one": 1, "three": 3, "two": 2}, m, "should not fail")
	if a.Equal("one", 1, "should fail and be logged") {
		t.Errorf("Equal() returned true")
	}
	if a.Equal("two", "2", "should fail and be logged") {
		t.Errorf("Equal() returned true")
	}
}

// TestAssertDifferent tests the Different() assertion.
func TestAssertDifferent(t *testing.T) {
	a := createValueAsserts(t)
	m := map[string]int{"one": 1, "two": 2, "three": 3}

	a.Different(nil, "nil", "should not fail")
	a.Different("true", true, "should not fail")
	a.Different(1, 2, "should not fail")
	a.Different("foo", "bar", "should not fail")
	a.Different(map[string]int{"three": 3, "two": 2}, m, "should not fail")
	if a.Different("one", "one", "should fail and be logged") {
		t.Errorf("Different() returned true")
	}
	if a.Different(2, 2, "should fail and be logged") {
		t.Errorf("Different() returned true")
	}
}

// TestAssertMatch tests the Match() assertion.
func TestAssertMatch(t *testing.T) {
	a := createValueAsserts(t)

	a.Match("this is a test", "this.*test", "should not fail")
	a.Match("this is 1 test", "this is [0-9] test", "should not fail")
	if a.Match("this is a test", "foo", "should fail and be logged") {
		t.Errorf("Match() returned true")
	}
	if a.Match("this is a test", "this*test", "should fail and be logged") {
		t.Errorf("Match() returned true")
	}
}

// TestAssertErrorMatch tests the ErrorMatch() assertion.
func TestAssertErrorMatch(t *testing.T) {
	a := createValueAsserts(t)
	err := errors.New("oops, an error")

	a.ErrorMatch(err, "oops, an error", "should not fail")
	a.ErrorMatch(err, "oops,.*", "should not fail")
	if a.ErrorMatch(err, "foo", "should fail and be logged") {
		t.Errorf("ErrorMatch() returned true")
	}
}

// TestAssertImplements tests the Implements() assertion.
func TestAssertImplements(t *testing.T) {
	a := createTypeAsserts(t)

	var err error
	var w io.Writer

	a.Implements(errors.New("error test"), &err, "should not fail")
	if a.Implements("string test", &err, "should fail and be logged") {
		t.Errorf("Implements() returned true")
	}
	if a.Implements(errors.New("error test"), &w, "should fail and be logged") {
		t.Errorf("Implements() returned true")
	}
}

// TestAssertAssignable tests the Assignable() assertion.
func TestAssertAssignable(t *testing.T) {
	a := createTypeAsserts(t)

	a.Assignable(1, 5, "should not fail")
	if a.Assignable("one", 5, "should fail and be logged") {
		t.Errorf("Assignable() returned true")
	}
}

// TestAssertUnassignable tests the Unassignable() assertion.
func TestAssertUnassignable(t *testing.T) {
	a := createTypeAsserts(t)

	a.Unassignable("one", 5, "should not fail")
	if a.Unassignable(1, 5, "should fail and be logged") {
		t.Errorf("Unassignable() returned true")
	}
}

// TestAssertEmpty tests the Empty() assertion.
func TestAssertEmpty(t *testing.T) {
	a := createValueAsserts(t)

	a.Empty("", "should not fail")
	a.Empty([]bool{}, "should also not fail")
	if a.Empty("not empty", "should fail and be logged") {
		t.Errorf("Empty() returned true")
	}
	if a.Empty([3]int{1, 2, 3}, "should also fail and be logged") {
		t.Errorf("Empty() returned true")
	}
	if a.Empty(true, "illegal type has to fail") {
		t.Errorf("Empty() returned true")
	}
}

// TestAssertNotEmpty tests the NotEmpty() assertion.
func TestAsserNotEmpty(t *testing.T) {
	a := createValueAsserts(t)

	a.NotEmpty("not empty", "should not fail")
	a.NotEmpty([3]int{1, 2, 3}, "should also not fail")
	if a.NotEmpty("", "should fail and be logged") {
		t.Errorf("NotEmpty() returned true")
	}
	if a.NotEmpty([]int{}, "should also fail and be logged") {
		t.Errorf("NotEmpty() returned true")
	}
	if a.NotEmpty(true, "illegal type has to fail") {
		t.Errorf("NotEmpty() returned true")
	}
}

// TestAssertLength tests the Length() assertion.
func TestAssertLength(t *testing.T) {
	a := createValueAsserts(t)

	a.Length("", 0, "should not fail")
	a.Length([]bool{true, false}, 2, "should also not fail")
	if a.Length("not empty", 0, "should fail and be logged") {
		t.Errorf("Length() returned true")
	}
	if a.Length([3]int{1, 2, 3}, 10, "should also fail and be logged") {
		t.Errorf("Length() returned true")
	}
	if a.Length(true, 1, "illegal type has to fail") {
		t.Errorf("Length() returned true")
	}
}

// TestPanicAssert tests if the panic assert panics when failing.
func TestPanicAssert(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			t.Logf("panic worked: '%v'", err)
		}
	}()

	a := NewPanicAsserts()
	foo := func() {}

	a.Assignable(47, 11, "should not fail")
	a.Assignable(47, foo, "should fail")

	t.Errorf("should not be reached")
}

// TestTestingAssert tests the testing assert.
func TestTestingAssert(t *testing.T) {
	a := NewTestingAsserts(t, false)
	foo := func() {}
	bar := 4711

	a.Assignable(47, 11, "should not fail")
	a.Assignable(foo, bar, "should fail")
}

// EOF
