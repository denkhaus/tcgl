// Tideland Common Go Library - Write once read multiple - Unit Tests
//
// Copyright (C) 2012 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed 
// by the new BSD license.

package worm_test

//--------------------
// IMPORTS
//--------------------

import (
	"github.com/denkhaus/tcgl/asserts"
	"github.com/denkhaus/tcgl/worm"
	"testing"
)

//--------------------
// TESTS
//--------------------

// TestDict tests the usage of the dictionary.
func TestDict(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	dv := worm.DictValues{
		"bool":         true,
		"bytes":        []byte{1, 2, 3, 4},
		"string":       "the quick brown fox yadda yadda ...",
		"int":          4711,
		"int64":        int64(-12345),
		"uint64":       uint64(12345),
		"float64":      float64(47.11),
		"complex128":   complex(47.11, 8.15),
		"struct":       Outer{"yadda", &Inner{47, 11}},
		"string-slice": []string{"one", "two", "three"},
		"map":          map[string]int{"one": 1, "two": 2, "three": 3},
	}
	d, err := worm.NewDict(dv)
	assert.Nil(err, "dict created")
	assert.Length(d, 11, "dict has right length")

	// Test valid access.
	b, err := d.Bool("bool")
	assert.Nil(err, "access ok")
	assert.Equal(b, dv["bool"], "right value")

	bs, err := d.Bytes("bytes")
	assert.Nil(err, "access ok")
	assert.Equal(bs, dv["bytes"], "right value")

	s, err := d.String("string")
	assert.Nil(err, "access ok")
	assert.Equal(s, dv["string"], "right value")

	i, err := d.Int("int")
	assert.Nil(err, "access ok")
	assert.Equal(i, dv["int"], "right value")

	i64, err := d.Int64("int64")
	assert.Nil(err, "access ok")
	assert.Equal(i64, dv["int64"], "right value")

	ui64, err := d.Uint64("uint64")
	assert.Nil(err, "access ok")
	assert.Equal(ui64, dv["uint64"], "right value")

	f64, err := d.Float64("float64")
	assert.Nil(err, "access ok")
	assert.Equal(f64, dv["float64"], "right value")

	c128, err := d.Complex128("complex128")
	assert.Nil(err, "access ok")
	assert.Equal(c128, dv["complex128"], "right value")

	var rss []string
	err = d.Read("string-slice", &rss)
	assert.Nil(err, "access ok")
	assert.Equal(rss, dv["string-slice"], "right value")

	var rm map[string]int
	err = d.Read("map", &rm)
	assert.Nil(err, "access ok")
	assert.Equal(rm, dv["map"], "right value")

	// Test string default.
	s = d.StringDefault("bool", "foo")
	assert.Equal(s, "true", "valid key, so string representation")

	s = d.StringDefault("bar", "foo")
	assert.Equal(s, "foo", "invalid key, so default value")

	// Test key access.
	keys := d.Keys()
	assert.Length(keys, d.Len(), "same length")
	assert.True(d.ContainsKeys(), "empty keys are ok")
	assert.True(d.ContainsKeys("bool", "int", "struct"), "keys detected")
	assert.False(d.ContainsKeys("bool", "foo", "struct"), "invalid keys recognized")

	// Test dictionary copy.
	nd := d.Copy()
	assert.Length(nd, 0, "copied nothing")

	nd = d.Copy("string-slice", "map")
	assert.Length(nd, 2, "copied both")
	err = nd.Read("string-slice", &rss)
	assert.Nil(err, "access ok")
	assert.Equal(rss, dv["string-slice"], "right value")
	err = nd.Read("map", &rm)
	assert.Nil(err, "access ok")
	assert.Equal(rm, dv["map"], "right value")

	nd = d.CopyAll()
	assert.Length(nd, d.Len(), "same length")

	// Test illegal keys and types.
	_, err = d.Bool("yadda")
	assert.True(worm.IsInvalidKeyError(err), "invalid key detected")

	_, err = d.Int("bool")
	assert.True(worm.IsInvalidTypeError(err), "invalid type detected")

	// Test applying more values.
	av := worm.DictValues{
		"int": 1234,
		"foo": "foo",
		"bar": "bar",
	}
	ad, err := d.Apply(av)
	assert.Nil(err, "apply ok")
	assert.Length(ad, d.Len()+2, "two new values in the dictionary")

	i, err = ad.Int("int")
	assert.Nil(err, "access ok")
	assert.Equal(i, av["int"], "right value, int has been changed")

	s, err = ad.String("foo")
	assert.Nil(err, "access ok")
	assert.Equal(s, av["foo"], "right new foo value")

	s, err = ad.String("bar")
	assert.Nil(err, "access ok")
	assert.Equal(s, av["bar"], "right new bar value")
}

// TestIntList tests the usage of int lists.
func TestIntList(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	values := worm.Ints{1, 5, 6, 2, 5, 2, 9}
	i := worm.NewIntList(values)

	// Test length.
	assert.Length(i, 7, "correct length")

	// Tast retrieving the values.
	values = i.Values()
	assert.Length(values, 7, "correct length")
	assert.Equal(values, worm.Ints{1, 5, 6, 2, 5, 2, 9}, "values are right")

	// Tast retrieving the values sorted.
	values = i.SortedValues()
	assert.Length(values, 7, "correct length")
	assert.Equal(values, worm.Ints{1, 2, 2, 5, 5, 6, 9}, "values are right")

	// Test containing test.
	// assert.True(i.Contains(), "emtpy values are ok")
	// assert.True(i.Contains(6, 5), "values detected")
	// assert.False(i.Contains(1, 2, 7), "invalid values recognized")

	// Test appending more values.
	av := worm.Ints{2, 6, 1001, 1010, 1005}
	ai := i.Append(av)
	assert.Length(ai, i.Len()+5, "five more values in the new list")
}

// TestIntSet tests the usage of int sets.
func TestIntSet(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	values := worm.Ints{1, 5, 6, 2, 5, 2, 9}
	i := worm.NewIntSet(values)

	// Test length.
	assert.Length(i, 5, "correct length")

	// Tast retrieving the values.
	values = i.Values()
	assert.Length(values, 5, "correct length")
	assert.Equal(values, worm.Ints{1, 2, 5, 6, 9}, "values are right")

	// Test containing test.
	assert.True(i.Contains(), "emtpy values are ok")
	assert.True(i.Contains(6, 5), "values detected")
	assert.False(i.Contains(1, 2, 7), "invalid values recognized")

	// Test applying more values.
	av := worm.Ints{2, 6, 1001, 1010, 1005}
	ai := i.Apply(av)
	assert.Length(ai, i.Len()+3, "three more values in the new set")
}

// TestStringSet tests the usage of string sets.
func TestStringSet(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	values := worm.Strings{"foo", "bar", "baz", "yadda", "foo", "yadda", "argle"}
	s := worm.NewStringSet(values)

	// Test length.
	assert.Length(s, 5, "correct length")

	// Tast retrieving the values.
	values = s.Values()
	assert.Length(values, 5, "correct length")
	assert.Equal(values, worm.Strings{"argle", "bar", "baz", "foo", "yadda"}, "values are right")

	// Test containing test.
	assert.True(s.Contains(), "emtpy values are ok")
	assert.True(s.Contains("baz", "foo"), "values detected")
	assert.False(s.Contains("argle", "yadda", "zapper"), "invalid values recognized")

	// Test applying more values.
	av := worm.Strings{"foo", "bar", "alpha", "bravo", "charlie"}
	as := s.Apply(av)
	assert.Length(as, s.Len()+3, "three more values in the new set")
}

//--------------------
// HELPER
//--------------------

type Inner struct {
	Foo int
	Bar int
}

type Outer struct {
	Yadda string
	Inner *Inner
}

// EOF
